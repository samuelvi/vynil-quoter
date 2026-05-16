# VinylQuoter Quickstart

VinylQuoter runs from Docker so the host computer does not need Go installed. It identifies vinyl records from cover images and updates `data/report/album_catalog.csv`.

## 1. Add images

Put cover images in:

```text
data/src/
```

Supported extensions: `.jpg`, `.jpeg`, `.png`, `.webp`, `.dng`, `.heic`, `.heif`, `.tif`, `.tiff`.

## 2. Build the Docker runtime

```bash
make docker-build
```

## 3. Run without parameters: interactive menu

```bash
make run
```

The interactive menu asks whether to process one image, all images in `data/src/`, update the CSV, regenerate the CSV, or exit. It also asks which recognition model to use:

- LM Studio local `qwen2.5-vl-7b-instruct` — default.
- LM Studio local `gemma-3-4b-it` — alternate local vision model if loaded in LM Studio.
- Gemini `gemini-2.5-flash-lite` — requires `GEMINI_API_KEY`.

When running in Docker, Make uses `http://host.docker.internal:1234/v1` for LM Studio by default. Override it if needed:

```bash
make run LM_STUDIO_BASE_URL=http://host.docker.internal:1234/v1
```

## 4. Run with parameters through Make

Process one image. This only creates or updates the CSV; it never replaces it:

```bash
make run IMAGE=DSC01.jpg
make run IMAGE=data/src/DSC01.jpg
```

Process every supported image in `data/src/` and append missing rows:

```bash
make run-all
```

Regenerate the CSV from all images:

```bash
make run-all-replace
```

Use Gemini:

```bash
make run-gemini
```

Use a specific LM Studio model:

```bash
make run-all MODEL=qwen2.5-vl-7b-instruct
make run-all MODEL=gemma-3-4b-it
```

## 5. Raw CLI flags through Make

If you need a flag not wrapped by the common Make commands, pass raw CLI flags through Docker with `ARGS`:

```bash
make run-cli ARGS="--all --provider lm-studio --model qwen2.5-vl-7b-instruct"
```

Useful CLI flags:

- `--src`: source image directory, default `data/src`.
- `--report`: CSV path, default `data/report/album_catalog.csv`.
- `--image`: one image path or file name from `data/src`.
- `--all`: process all supported images in `data/src`.
- `--replace`: regenerate the CSV, only meaningful with `--all`.
- `--provider`: `lm-studio` or `gemini`.
- `--model`: model name override.
- `--lm-studio-base-url`: LM Studio OpenAI-compatible base URL.

## 6. Review the CSV

Output path:

```text
data/report/album_catalog.csv
```

The `source_image` column is the image ID and stores the file name, for example `DSC01.jpg`.
