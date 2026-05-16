#!/usr/bin/env python3
"""Identify vinyl albums from images and write a valuation CSV report."""

from __future__ import annotations

import argparse
import base64
import csv
import json
import mimetypes
import os
import re
import sys
import time
import urllib.error
import urllib.request
from dataclasses import asdict, dataclass
from pathlib import Path
from typing import Callable


SUPPORTED_EXTENSIONS = {".dng", ".heic", ".heif", ".jpg", ".jpeg", ".png", ".tif", ".tiff", ".webp"}
DEFAULT_SRC_DIR = Path("src")
DEFAULT_REPORT = Path("report") / "album_catalog.csv"
DEFAULT_GEMINI_MODEL = "gemini-2.5-flash-lite"
DEFAULT_LM_STUDIO_BASE_URL = "http://localhost:1234/v1"
DEFAULT_LM_STUDIO_VISION_MODEL = "qwen2.5-vl-7b-instruct"
CONFIDENCE_VALUES = {"high", "medium", "low", "manual-review"}
RETRYABLE_HTTP_STATUS = {429, 503}
CSV_COLUMNS = [
    "source_image",
    "artist",
    "title",
    "identification_confidence",
    "recommended_price_eur",
    "price_confidence",
    "price_basis",
    "notes",
]


@dataclass(frozen=True)
class AlbumCatalogData:
    artist: str
    title: str
    identification_confidence: str
    recommended_price_eur: str
    price_confidence: str
    price_basis: str
    notes: str


@dataclass(frozen=True)
class AlbumCatalogRow:
    source_image: str
    artist: str
    title: str
    identification_confidence: str
    recommended_price_eur: str
    price_confidence: str
    price_basis: str
    notes: str


@dataclass(frozen=True)
class RunConfig:
    image: Path | None
    all_images: bool
    replace: bool
    src_dir: Path = DEFAULT_SRC_DIR
    report: Path = DEFAULT_REPORT
    provider: str = "lm-studio"
    model: str = DEFAULT_LM_STUDIO_VISION_MODEL


class RetryableGeminiError(RuntimeError):
    def __init__(self, message: str, *, retry_after: float | None = None):
        super().__init__(message)
        self.retry_after = retry_after


class GeminiQuotaExhaustedError(RuntimeError):
    pass


def is_supported_image(path: Path) -> bool:
    return path.suffix.lower() in SUPPORTED_EXTENSIONS


def collect_work_items(src_dir: Path, *, image: Path | None, all_images: bool) -> list[Path]:
    if image is None and not all_images:
        raise ValueError("Choose a specific image or all images")
    if image is not None and all_images:
        raise ValueError("Choose only one image mode")

    if image is not None:
        resolved = image if image.exists() else src_dir / image
        if not resolved.exists():
            raise FileNotFoundError(f"Image not found: {image}")
        if not is_supported_image(resolved):
            raise ValueError(f"Unsupported image extension: {resolved.suffix}")
        return [resolved]

    if not src_dir.exists():
        raise FileNotFoundError(f"Source directory not found: {src_dir}")
    return sorted(path for path in src_dir.iterdir() if path.is_file() and is_supported_image(path))


def read_catalog_rows(catalog_path: Path) -> list[AlbumCatalogRow]:
    if not catalog_path.exists():
        return []
    with catalog_path.open(newline="", encoding="utf-8") as handle:
        rows = csv.DictReader(handle)
        return [AlbumCatalogRow(**{column: row.get(column, "") for column in CSV_COLUMNS}) for row in rows]


def write_catalog_csv(rows: list[AlbumCatalogRow], catalog_path: Path) -> None:
    catalog_path.parent.mkdir(parents=True, exist_ok=True)
    with catalog_path.open("w", newline="", encoding="utf-8") as handle:
        writer = csv.DictWriter(handle, fieldnames=CSV_COLUMNS)
        writer.writeheader()
        for row in rows:
            writer.writerow(asdict(row))


def pending_images(images: list[Path], catalog_path: Path, *, replace: bool) -> list[Path]:
    if replace:
        return images
    existing = {row.source_image for row in read_catalog_rows(catalog_path)}
    existing.update(Path(row.source_image).name for row in read_catalog_rows(catalog_path))
    return [image for image in images if str(image) not in existing and image.name not in existing]


def _json_object_from_text(text: str) -> dict:
    cleaned = text.strip()
    fenced = re.search(r"```(?:json)?\s*(\{.*?\})\s*```", cleaned, re.DOTALL)
    if fenced:
        cleaned = fenced.group(1)
    else:
        start = cleaned.find("{")
        end = cleaned.rfind("}")
        if start == -1 or end == -1 or end <= start:
            raise ValueError("Gemini response did not contain a JSON object")
        cleaned = cleaned[start : end + 1]
    return json.loads(cleaned)


def _confidence(value: object, default: str = "manual-review") -> str:
    normalized = str(value or default).strip().lower()
    return normalized if normalized in CONFIDENCE_VALUES else default


def parse_gemini_catalog_response(text: str) -> AlbumCatalogData:
    payload = _json_object_from_text(text)
    return AlbumCatalogData(
        artist=str(payload.get("artist") or "Unknown").strip() or "Unknown",
        title=str(payload.get("title") or "Unknown").strip() or "Unknown",
        identification_confidence=_confidence(payload.get("identification_confidence")),
        recommended_price_eur=str(payload.get("recommended_price_eur") or "").strip(),
        price_confidence=_confidence(payload.get("price_confidence"), default="low"),
        price_basis=str(payload.get("price_basis") or "").strip(),
        notes=str(payload.get("notes") or "").strip(),
    )


def extract_gemini_text(response: dict) -> str:
    candidates = response.get("candidates") or []
    if not candidates:
        raise ValueError("Gemini returned no candidates")
    parts = candidates[0].get("content", {}).get("parts", [])
    text = "\n".join(part.get("text", "") for part in parts if isinstance(part, dict)).strip()
    if not text:
        raise ValueError("Gemini returned no text")
    return text


def _duration_seconds(value: str) -> float | None:
    if value.endswith("ms"):
        try:
            return float(value[:-2]) / 1000
        except ValueError:
            return None
    if value.endswith("s"):
        try:
            return float(value[:-1])
        except ValueError:
            return None
    return None


def retry_delay_from_error_body(body: str, *, default: float) -> float:
    try:
        payload = json.loads(body)
    except json.JSONDecodeError:
        payload = {}
    for detail in payload.get("error", {}).get("details", []):
        retry_delay = detail.get("retryDelay") if isinstance(detail, dict) else None
        if isinstance(retry_delay, str):
            duration = _duration_seconds(retry_delay)
            if duration:
                return duration
    match = re.search(r"retry in ([0-9.]+)\s*(ms|s)", body, re.IGNORECASE)
    if match:
        value = float(match.group(1))
        return value / 1000 if match.group(2).lower() == "ms" else value
    return default


def body_has_daily_quota_exhausted(body: str) -> bool:
    try:
        payload = json.loads(body)
    except json.JSONDecodeError:
        return "per day" in body.lower()
    message = str(payload.get("error", {}).get("message", "")).lower()
    if "per day" in message:
        return True
    for detail in payload.get("error", {}).get("details", []):
        if not isinstance(detail, dict):
            continue
        for violation in detail.get("violations", []):
            if "PerDay" in str(violation.get("quotaId", "")):
                return True
    return False


def retry_operation(operation: Callable[[], object], *, max_retries: int, retry_delay: float) -> object:
    attempt = 0
    while True:
        try:
            return operation()
        except RetryableGeminiError as error:
            if attempt >= max_retries:
                raise
            time.sleep(error.retry_after if error.retry_after is not None else retry_delay)
            attempt += 1


def request_gemini_json(request: urllib.request.Request, *, timeout: int, retry_delay: float) -> dict:
    try:
        with urllib.request.urlopen(request, timeout=timeout) as response:
            return json.loads(response.read().decode("utf-8"))
    except urllib.error.HTTPError as error:
        body = error.read().decode("utf-8", errors="replace")
        message = f"Gemini HTTP {error.code}: {body}"
        if error.code == 429 and body_has_daily_quota_exhausted(body):
            raise GeminiQuotaExhaustedError(message) from error
        if error.code in RETRYABLE_HTTP_STATUS:
            delay = retry_delay_from_error_body(body, default=retry_delay)
            raise RetryableGeminiError(message, retry_after=delay) from error
        raise RuntimeError(message) from error
    except urllib.error.URLError as error:
        raise RuntimeError(f"Gemini request failed: {error}") from error


def identify_album_with_gemini(
    image_path: Path,
    *,
    api_key: str,
    model: str = DEFAULT_GEMINI_MODEL,
    timeout: int = 60,
    max_retries: int = 3,
    retry_delay: float = 7.0,
) -> AlbumCatalogData:
    mime_type = mimetypes.guess_type(image_path.name)[0] or "image/jpeg"
    if not mime_type.startswith("image/"):
        mime_type = "image/jpeg"
    prompt = (
        "Identify the vinyl album from this front cover/source image and estimate a conservative second-hand sale price. "
        "Return only JSON with this exact shape: "
        '{"artist":"string","title":"string",'
        '"identification_confidence":"high|medium|low|manual-review",'
        '"recommended_price_eur":"string",'
        '"price_confidence":"high|medium|low|manual-review",'
        '"price_basis":"string","notes":"string"}. '
        "Use Unknown for artist/title if the cover is unreadable or ambiguous. "
        "Price assumptions: Spain/EU market, EUR, media VG+, sleeve VG, normal second-hand sale. "
        "If uncertain, use low or manual-review confidence and explain in notes. "
        "Do not include markdown or commentary."
    )
    payload = {
        "contents": [
            {
                "parts": [
                    {"text": prompt},
                    {"inline_data": {"mime_type": mime_type, "data": base64.b64encode(image_path.read_bytes()).decode("ascii")}},
                ]
            }
        ],
        "generationConfig": {"temperature": 0, "responseMimeType": "application/json"},
    }
    url = f"https://generativelanguage.googleapis.com/v1beta/models/{model}:generateContent?key={api_key}"
    request = urllib.request.Request(
        url,
        data=json.dumps(payload).encode("utf-8"),
        headers={"Content-Type": "application/json"},
        method="POST",
    )
    response = retry_operation(
        lambda: request_gemini_json(request, timeout=timeout, retry_delay=retry_delay),
        max_retries=max_retries,
        retry_delay=retry_delay,
    )
    return parse_gemini_catalog_response(extract_gemini_text(response))


def identify_album_with_lm_studio(
    image_path: Path,
    *,
    model: str = DEFAULT_LM_STUDIO_VISION_MODEL,
    base_url: str = DEFAULT_LM_STUDIO_BASE_URL,
    timeout: int = 60,
) -> AlbumCatalogData:
    mime_type = mimetypes.guess_type(image_path.name)[0] or "image/jpeg"
    if not mime_type.startswith("image/"):
        mime_type = "image/jpeg"
    prompt = (
        "Identify the vinyl album from this front cover/source image and estimate a conservative second-hand sale price. "
        "Return only JSON with this exact shape: "
        '{"artist":"string","title":"string",'
        '"identification_confidence":"high|medium|low|manual-review",'
        '"recommended_price_eur":"string",'
        '"price_confidence":"high|medium|low|manual-review",'
        '"price_basis":"string","notes":"string"}. '
        "Use Unknown for artist/title if unreadable. Price assumptions: Spain/EU market, EUR, media VG+, sleeve VG. "
        "Do not include markdown or commentary."
    )
    payload = {
        "model": model,
        "temperature": 0,
        "messages": [
            {
                "role": "user",
                "content": [
                    {"type": "text", "text": prompt},
                    {
                        "type": "image_url",
                        "image_url": {"url": f"data:{mime_type};base64,{base64.b64encode(image_path.read_bytes()).decode('ascii')}"},
                    },
                ],
            }
        ],
    }
    request = urllib.request.Request(
        f"{base_url.rstrip('/')}/chat/completions",
        data=json.dumps(payload).encode("utf-8"),
        headers={"Content-Type": "application/json"},
        method="POST",
    )
    try:
        with urllib.request.urlopen(request, timeout=timeout) as response:
            response_payload = json.loads(response.read().decode("utf-8"))
    except urllib.error.URLError as error:
        raise RuntimeError(f"LM Studio request failed: {error}") from error
    choices = response_payload.get("choices") or []
    if not choices:
        raise ValueError("LM Studio returned no choices")
    text = str(choices[0].get("message", {}).get("content", "")).strip()
    if not text:
        raise ValueError("LM Studio returned no text")
    return parse_gemini_catalog_response(text)


def process_images(
    images: list[Path],
    report_path: Path,
    *,
    replace: bool,
    identify: Callable[[Path], AlbumCatalogData],
) -> list[AlbumCatalogRow]:
    rows = [] if replace else read_catalog_rows(report_path)
    for image in pending_images(images, report_path, replace=replace):
        try:
            data = identify(image)
        except GeminiQuotaExhaustedError:
            raise
        except Exception as error:  # noqa: BLE001 - keep per-image failures in the CSV.
            data = AlbumCatalogData(
                artist="Unknown",
                title="Unknown",
                identification_confidence="manual-review",
                recommended_price_eur="",
                price_confidence="manual-review",
                price_basis="",
                notes=f"identification failed: {error}",
            )
        rows.append(
            AlbumCatalogRow(
                source_image=str(image),
                artist=data.artist,
                title=data.title,
                identification_confidence=data.identification_confidence,
                recommended_price_eur=data.recommended_price_eur,
                price_confidence=data.price_confidence,
                price_basis=data.price_basis,
                notes=data.notes,
            )
        )
        write_catalog_csv(rows, report_path)
    if not report_path.exists():
        write_catalog_csv(rows, report_path)
    return rows


def choose_provider(input_fn: Callable[[str], str] = input) -> tuple[str, str]:
    print("\nModelo de reconocimiento")
    print(f"1) LM Studio local - {DEFAULT_LM_STUDIO_VISION_MODEL} [por defecto]")
    print(f"2) Gemini - {DEFAULT_GEMINI_MODEL}")
    choice = input_fn("Elige modelo [1-2, Enter=1]: ").strip()
    if choice == "2":
        return "gemini", DEFAULT_GEMINI_MODEL
    return "lm-studio", DEFAULT_LM_STUDIO_VISION_MODEL


def menu_config(input_fn: Callable[[str], str] = input) -> RunConfig:
    print("\nVinyl Quoter")
    print("1) Procesar una imagen concreta")
    print("2) Procesar todas las imágenes de src")
    print("3) Actualizar CSV final por defecto")
    print("4) Machacar/regenerar CSV final por defecto")
    print("5) Salir")
    choice = input_fn("Elige una opción [1-5]: ").strip()

    if choice == "1":
        value = input_fn("Ruta o nombre de imagen: ").strip()
        provider, model = choose_provider(input_fn)
        return RunConfig(image=Path(value), all_images=False, replace=False, provider=provider, model=model)
    if choice in {"2", "3"}:
        provider, model = choose_provider(input_fn)
        return RunConfig(image=None, all_images=True, replace=False, provider=provider, model=model)
    if choice == "4":
        provider, model = choose_provider(input_fn)
        return RunConfig(image=None, all_images=True, replace=True, provider=provider, model=model)
    if choice == "5":
        raise SystemExit(0)
    raise ValueError(f"Invalid menu choice: {choice}")


def parse_args(argv: list[str] | None = None) -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Identify vinyl albums from images and write report/album_catalog.csv.")
    parser.add_argument("--src", type=Path, default=DEFAULT_SRC_DIR, help="Source image directory. Default: src")
    parser.add_argument("--report", type=Path, default=DEFAULT_REPORT, help="CSV report path. Default: report/album_catalog.csv")
    parser.add_argument("--image", type=Path, help="Process one image path or filename from src")
    parser.add_argument("--all", action="store_true", help="Process all supported images from src")
    parser.add_argument("--replace", action="store_true", help="Regenerate the final CSV instead of updating it")
    parser.add_argument("--provider", choices=["lm-studio", "gemini"], default="lm-studio", help="Vision provider. Default: lm-studio")
    parser.add_argument("--model", help=f"Vision model. Default: {DEFAULT_LM_STUDIO_VISION_MODEL} for LM Studio, {DEFAULT_GEMINI_MODEL} for Gemini")
    parser.add_argument("--lm-studio-base-url", default=DEFAULT_LM_STUDIO_BASE_URL, help="LM Studio OpenAI-compatible base URL")
    parser.add_argument("--timeout", type=int, default=60, help="Gemini request timeout in seconds. Default: 60")
    parser.add_argument("--max-retries", type=int, default=3, help="Retries for Gemini 429/503 responses. Default: 3")
    parser.add_argument("--retry-delay", type=float, default=7.0, help="Fallback retry delay. Default: 7")
    return parser.parse_args(argv)


def ensure_api_key() -> str:
    api_key = os.environ.get("GEMINI_API_KEY", "").strip()
    if not api_key:
        raise RuntimeError("GEMINI_API_KEY is required")
    return api_key


def main(argv: list[str] | None = None) -> int:
    args = parse_args(argv)
    try:
        default_model = DEFAULT_GEMINI_MODEL if args.provider == "gemini" else DEFAULT_LM_STUDIO_VISION_MODEL
        config = RunConfig(args.image, args.all, args.replace, args.src, args.report, args.provider, args.model or default_model)
        if config.image is None and not config.all_images:
            config = menu_config()
        images = collect_work_items(config.src_dir, image=config.image, all_images=config.all_images)
        if config.provider == "gemini":
            api_key = ensure_api_key()

            def identify(image: Path) -> AlbumCatalogData:
                return identify_album_with_gemini(
                    image,
                    api_key=api_key,
                    model=config.model,
                    timeout=args.timeout,
                    max_retries=args.max_retries,
                    retry_delay=args.retry_delay,
                )
        else:

            def identify(image: Path) -> AlbumCatalogData:
                return identify_album_with_lm_studio(
                    image,
                    model=config.model,
                    base_url=args.lm_studio_base_url,
                    timeout=args.timeout,
                )

        rows = process_images(
            images,
            config.report,
            replace=config.replace,
            identify=identify,
        )
    except SystemExit:
        raise
    except Exception as error:  # noqa: BLE001 - CLI reports concise failures.
        print(f"error: {error}", file=sys.stderr)
        return 2

    print(f"CSV generado: {config.report} ({len(rows)} filas)")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
