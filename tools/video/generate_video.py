#!/usr/bin/env python3
"""Generate a local MP4 demo video from prompt.txt."""

from __future__ import annotations

import shutil
import subprocess
from pathlib import Path

from PIL import Image, ImageDraw, ImageFont


ROOT = Path(__file__).resolve().parents[2]
PROMPT = ROOT / "prompt.txt"
FRAME_DIR = ROOT / ".cache" / "video" / "frames"
OUTPUT = ROOT / "data" / "video" / "vinylquoter-ai-demo.mp4"
WIDTH = 1280
HEIGHT = 720
FPS = 24
SLIDES = [
    {
        "title": "VinylQuoter: generado con IA por fases",
        "subtitle": "Modelo: GPT-5.5 · Agente: Builder Workflow Agent",
        "terminal": ["$ git status --short", "$ go test ./...", "$ make test"],
        "badge": "0-5s · contexto inicial",
        "duration": 5,
    },
    {
        "title": "Contexto antes de actuar",
        "subtitle": "Repo + README + Makefile + historial Git + Memory Bank",
        "terminal": ["$ git branch --show-current", "main", "$ ls cmd internal docker data docs", "contexto cargado ✓"],
        "badge": "5-12s · context loading",
        "duration": 7,
    },
    {
        "title": "Plan primero, ejecución después",
        "subtitle": "Skills: brainstorming · writing-plans · executing-plans",
        "terminal": ["Fase 1: CLI y CSV", "Fase 2: proveedores IA", "Fase 3: data/ + Git hygiene", "Fase 4: Docker test", "Fase 5: migración Go"],
        "badge": "12-20s · planificación",
        "duration": 8,
    },
    {
        "title": "TDD: rojo, verde, refactor",
        "subtitle": "Primero prueba fallida; después implementación mínima",
        "terminal": ["FAIL TestCollectAllImages", "implement imageinput.Collect", "PASS internal/imageinput", "PASS internal/catalog"],
        "badge": "20-30s · test-driven-development",
        "duration": 10,
    },
    {
        "title": "Go con responsabilidades separadas",
        "subtitle": "cmd + internal packages",
        "terminal": ["cmd/vinyl-quoter/main.go", "internal/app", "internal/catalog", "internal/imageinput", "internal/provider/gemini", "internal/provider/lmstudio", "internal/ui"],
        "badge": "30-42s · arquitectura",
        "duration": 12,
    },
    {
        "title": "Docker test reproducible",
        "subtitle": "Volúmenes bind explícitos y caché local del proyecto",
        "terminal": ["docker/test/Dockerfile", "docker/test/docker-compose.yml", "type: bind -> /workspace", ".cache/go -> /go/pkg/mod"],
        "badge": "42-50s · entorno reproducible",
        "duration": 8,
    },
    {
        "title": "Evidencia antes de cerrar",
        "subtitle": "Tests, Docker y quality gate en verde",
        "terminal": ["$ go test ./...", "ok vinylquoter/internal/app", "ok vinylquoter/internal/catalog", "$ quality_gate.py --strict", "ok: true"],
        "badge": "50-57s · verificación",
        "duration": 7,
    },
    {
        "title": "Flujo final",
        "subtitle": "data/src → VinylQuoter → data/report/album_catalog.csv",
        "terminal": ["Modelo local: qwen2.5-vl-7b-instruct", "Fallback: Gemini", "CSV: artista, álbum, precio EUR, confianza", "IA útil = contexto + plan + fases + tests"],
        "badge": "57-60s · resultado",
        "duration": 3,
    },
]


def font(size: int, bold: bool = False) -> ImageFont.FreeTypeFont:
    candidates = [
        "/usr/share/fonts/truetype/dejavu/DejaVuSans-Bold.ttf" if bold else "/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf",
        "/Library/Fonts/Arial.ttf",
    ]
    for candidate in candidates:
        try:
            return ImageFont.truetype(candidate, size)
        except OSError:
            continue
    return ImageFont.load_default()


def draw_slide(slide: dict[str, object], path: Path) -> None:
    img = Image.new("RGB", (WIDTH, HEIGHT), "#0b1020")
    draw = ImageDraw.Draw(img)
    title_font = font(46, bold=True)
    subtitle_font = font(26)
    mono_font = font(24)
    small_font = font(20, bold=True)

    draw.rectangle((0, 0, WIDTH, HEIGHT), fill="#0b1020")
    draw.rectangle((48, 42, 1232, 106), fill="#111a33", outline="#2c3d74", width=2)
    draw.text((72, 56), str(slide["badge"]), fill="#9fb7ff", font=small_font)

    draw.text((72, 144), str(slide["title"]), fill="#ffffff", font=title_font)
    draw.text((72, 210), str(slide["subtitle"]), fill="#b7c4e8", font=subtitle_font)

    panel = (72, 292, 1208, 636)
    draw.rounded_rectangle(panel, radius=18, fill="#050812", outline="#27365f", width=2)
    draw.rectangle((72, 292, 1208, 334), fill="#10172b")
    draw.ellipse((94, 306, 110, 322), fill="#ff5f57")
    draw.ellipse((120, 306, 136, 322), fill="#ffbd2e")
    draw.ellipse((146, 306, 162, 322), fill="#28c840")

    y = 362
    for line in slide["terminal"]:
        color = "#7ee787" if "PASS" in line or "ok" in line or "✓" in line or "true" in line else "#d6deff"
        if "FAIL" in line:
            color = "#ff7b72"
        draw.text((104, y), str(line), fill=color, font=mono_font)
        y += 36

    draw.text((72, 668), "VinylQuoter AI demo · generado localmente con ffmpeg", fill="#7280a7", font=small_font)
    img.save(path)


def main() -> None:
    if not PROMPT.exists():
        raise SystemExit("prompt.txt not found")
    shutil.rmtree(FRAME_DIR, ignore_errors=True)
    FRAME_DIR.mkdir(parents=True, exist_ok=True)
    OUTPUT.parent.mkdir(parents=True, exist_ok=True)

    frame_index = 0
    for slide in SLIDES:
        slide_path = FRAME_DIR / f"slide-{frame_index:03d}.png"
        draw_slide(slide, slide_path)
        repeat = FPS * int(slide["duration"])
        for _ in range(repeat):
            frame_path = FRAME_DIR / f"frame-{frame_index:05d}.png"
            if not frame_path.exists():
                shutil.copy(slide_path, frame_path)
            frame_index += 1

    subprocess.run(
        [
            "ffmpeg",
            "-y",
            "-framerate",
            str(FPS),
            "-i",
            str(FRAME_DIR / "frame-%05d.png"),
            "-c:v",
            "libx264",
            "-pix_fmt",
            "yuv420p",
            "-movflags",
            "+faststart",
            str(OUTPUT),
        ],
        check=True,
    )
    print(f"Video generated: {OUTPUT}")


if __name__ == "__main__":
    main()
