# VinylQuoter Quickstart

VinylQuoter runs through Docker, so normal use does not require Go on the host.
Make prepares Docker automatically for user commands. No manual rebuild is needed for normal use: Go recompiles changed code automatically inside Docker.

## 1. Add images

Put cover photos in:

```text
data/src/
```

Supported extensions: `.jpg`, `.jpeg`, `.png`, `.webp`, `.dng`, `.heic`, `.heif`, `.tif`, `.tiff`.

The pipeline is always:

```text
data/src → data/dst → model → CSV
```

Decodable inputs are cropped locally and saved as JPG in `data/dst/`. Supported inputs that cannot be decoded locally are copied to `data/dst/` as fallback.

## 2. Start the interactive menu

```bash
make run
```

The menu stays open after each action and asks for confirmation when you choose `Salir`.

Main menu:

0. `Salir`
1. `Procesar una imagen concreta`
2. `Procesar todas las imágenes de data/src`
3. `Guardado csv (<current CSV path>)`
4. `Modelo (<current provider: model>)`
5. `Calidad carátula (<current sleeve condition>)`
6. `Calidad vinilo (<current media condition>)`

Defaults:

- Provider: LM Studio.
- Model: `qwen2.5-vl-7b-instruct`.
- Media condition: `VG`.
- Sleeve condition: `VG`.
- CSV path: `data/report/album_catalog.csv`.

To override these defaults locally, copy `.env.example` to `.env` and change the `VINYLQUOTER_*` values you need. CLI flags still take precedence over `.env` and process environment values.

## 3. Run common actions directly

Process one image:

```bash
make run IMAGE=DSC01.jpg
make run IMAGE=data/src/DSC01.jpg
```

Process every supported image and update existing CSV rows:

```bash
make run-all
```

Regenerate the CSV from the current images in `data/src/`:

```bash
make run-all-replace
```

Use Gemini instead of LM Studio:

```bash
make run-gemini
```

Use a specific LM Studio model:

```bash
make run-all MODEL=gemma-3-4b-it
```

Pass raw flags through Docker when needed:

```bash
make run-cli ARGS="--all --provider lm-studio --model qwen2.5-vl-7b-instruct"
```

## 4. Review output

Default CSV:

```text
data/report/album_catalog.csv
```

The row identifier is `source_image`, which stores the original image file name such as `DSC01.jpg`.

Reference URL columns are `discogs_reference_url`, `ebay_reference_url`, and `popsike_reference_url`.

See [CSV Output](CSV.md) for columns and update rules.

## 5. Troubleshooting

- `GEMINI_API_KEY is required`: choose LM Studio or export `GEMINI_API_KEY`.
- `LM Studio request failed`: start LM Studio, load a vision model, and enable the local server.
- Missing or stale rows: run `make run-all` to refresh all current images, or `make run-all-replace` to rebuild the CSV from current inputs.
