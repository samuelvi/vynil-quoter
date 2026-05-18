# VinylQuoter Usage Guide

Use this guide for runtime behavior, CLI flags, providers, and interactive menu details.

## Runtime model

All app commands run in Docker through the repository `Makefile`. The runtime image is built automatically when needed.

```bash
make run
make run-all
make run-cli ARGS="--all --provider lm-studio"
```

`go run` executes inside Docker, so changed Go code is recompiled automatically during development.

## Image pipeline

```text
data/src → data/dst → model → data/report/album_catalog.csv
```

1. `internal/imageinput` collects one image or all supported images from `data/src/`.
2. `internal/crop` prepares each image into `data/dst/`.
3. The selected provider analyzes the prepared path from `data/dst/`.
4. `internal/catalog` writes or updates the CSV.

Decodable images are cropped and written as `.jpg`. Supported images that cannot be decoded locally are copied unchanged to `data/dst/` so the recognizer can still inspect them.

## Interactive menu

Run:

```bash
make run
```

Main menu:

1. `Procesar una imagen concreta`
2. `Procesar todas las imágenes de data/src`
3. `Guardado csv (<current CSV path>)`
4. `Modelo (<current provider: model>)`
5. `Calidad carátula (<current sleeve condition>)`
6. `Calidad vinilo (<current media condition>)`
7. `Salir`

The menu persists after each action. Provider, model, CSV path, media condition, and sleeve condition remain selected for the session.

### CSV submenu

`Guardado csv` lets you:

1. Change the current CSV path.
2. Update the current CSV from all images.
3. Regenerate the current CSV from all images.
4. Return to the main menu.

This submenu does not ask for model settings.

### Model submenu

Available choices:

- LM Studio local `qwen2.5-vl-7b-instruct` — default.
- LM Studio local `gemma-3-4b-it` — alternate local vision model when loaded in LM Studio.
- Gemini `gemini-2.5-flash-lite` — requires `GEMINI_API_KEY`.

When running in Docker, Make sets the LM Studio base URL to `http://host.docker.internal:1234/v1` by default. Override it when needed:

```bash
make run LM_STUDIO_BASE_URL=http://host.docker.internal:1234/v1
```

### Condition submenus

Media condition values:

```text
M, NM/M-, VG+, VG, G+, G, F, P
```

Sleeve condition values:

```text
M, NM/M-, VG+, VG, G+, G, F, P, Generic
```

Both default to `VG`. The selected conditions are sent to the provider prompt and written to the CSV as `media: <grade>; sleeve: <grade>`.

## Raw CLI flags

Use `make run-cli` for flags not wrapped by a dedicated Make target:

```bash
make run-cli ARGS="--all --provider lm-studio --model qwen2.5-vl-7b-instruct"
```

Supported flags:

- `--src`: source image directory, default `data/src`.
- `--dst`: prepared image directory, default `data/dst`.
- `--report`: CSV report path, default `data/report/album_catalog.csv`.
- `--image`: one image path or file name from `data/src`.
- `--all`: process every supported image from `data/src`.
- `--replace`: regenerate the CSV instead of updating it; single-image runs never replace.
- `--provider`: `lm-studio` or `gemini`.
- `--model`: provider model override.
- `--lm-studio-base-url`: LM Studio OpenAI-compatible base URL.
- `--timeout`: request timeout seconds.
- `--max-retries`: Gemini retry count.
- `--retry-delay`: fallback retry delay in seconds.
- `--media-condition`: media grade.
- `--sleeve-condition`: sleeve grade.

## Provider requirements

### LM Studio

Start LM Studio, load a supported vision model, and enable the local server.

The project defaults to `qwen2.5-vl-7b-instruct`. The alternate local model is `gemma-3-4b-it`.

### Gemini

Export `GEMINI_API_KEY` before running Gemini commands. Do not commit API keys or print secret values.
