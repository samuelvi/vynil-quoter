# VinylQuoter Quickstart

VinylQuoter runs from Docker so the host computer does not need Go installed. It identifies vinyl records from cover images and updates `data/report/album_catalog.csv`.

Make prepares Docker automatically for user commands. You can run `make run` directly; it builds the Docker image if needed.

No manual rebuild is needed for normal use. The app uses `go run` inside Docker, and Go recompiles changed code automatically inside Docker.

## 1. Add images

Put cover images in:

```text
data/src/
```

Supported extensions: `.jpg`, `.jpeg`, `.png`, `.webp`, `.dng`, `.heic`, `.heif`, `.tif`, `.tiff`.

Processing pipeline: `data/src → data/dst → model → CSV`. JPEG and PNG images are cropped locally to isolate the vinyl/cover before model analysis. Other formats are copied to `data/dst` as a fallback.

## 2. Run without parameters: interactive menu

```bash
make run
```

The interactive menu stays open after each action and exits only when you choose `Salir`.

Main menu:

1. Process one image.
2. Process every supported image in `data/src`.
3. Open `Guardado csv (<current CSV path>)`.
4. Open `Modelo (<current provider: model>)`.
5. Exit.

The `Guardado csv` submenu lets you change the current CSV path, update the current CSV, regenerate it from scratch, or go back to the main menu. It does not ask for the recognition model.

The `Modelo` menu lets you choose the recognition model once and keeps that selection for later actions:

- LM Studio local `qwen2.5-vl-7b-instruct` — default.
- LM Studio local `gemma-3-4b-it` — alternate local vision model if loaded in LM Studio.
- Gemini `gemini-2.5-flash-lite` — requires `GEMINI_API_KEY`.

When running in Docker, Make uses `http://host.docker.internal:1234/v1` for LM Studio by default. Override it if needed:

```bash
make run LM_STUDIO_BASE_URL=http://host.docker.internal:1234/v1
```

## 3. Run with parameters through Make

Process one image. This only creates or updates the CSV; it never replaces it:

```bash
make run IMAGE=DSC01.jpg
make run IMAGE=data/src/DSC01.jpg
```

Process every supported image in `data/src`, overwrite prepared images in `data/dst`, and upsert CSV rows:

```bash
make run-all
```

Existing CSV rows with the same `source_image` are replaced with fresh recognition results. Rows for images not selected in the current run are kept unless you use replace mode.

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

## 4. Raw CLI flags through Make

If you need a flag not wrapped by the common Make commands, pass raw CLI flags through Docker with `ARGS`:

```bash
make run-cli ARGS="--all --provider lm-studio --model qwen2.5-vl-7b-instruct"
```

Useful CLI flags:

- `--src`: source image directory, default `data/src`.
- `--dst`: cropped image directory, default `data/dst`.
- `--report`: CSV path, default `data/report/album_catalog.csv`.
- `--image`: one image path or file name from `data/src`.
- `--all`: process all supported images in `data/src`.
- `--replace`: regenerate the CSV, only meaningful with `--all`.
- `--provider`: `lm-studio` or `gemini`.
- `--model`: model name override.
- `--lm-studio-base-url`: LM Studio OpenAI-compatible base URL.

## 5. Review the CSV

Output path:

```text
data/report/album_catalog.csv
```

The `source_image` column is the image ID and stores the file name, for example `DSC01.jpg`.

Price reference columns:

- `discogs_reference_url`
- `ebay_reference_url`
- `popsike_reference_url`

These are generic external search links generated from artist/title plus `vinyl VG+ sleeve VG+`, a medium-high condition assumption for vinyl and sleeve.
