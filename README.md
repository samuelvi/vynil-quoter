# VinylQuoter

VinylQuoter identifies vinyl records from cover images in `data/src/` and writes a valuation catalog to `data/report/album_catalog.csv`.

For the fastest path, read [docs/QUICKSTART.md](docs/QUICKSTART.md).

## What it does

- Scans one image or every supported image in `data/src/`.
- Identifies album title and artist/band using LM Studio or Gemini.
- Estimates a conservative EUR resale value.
- Writes a CSV catalog with confidence and review notes.

Default provider: LM Studio local model `qwen2.5-vl-7b-instruct`.

## Project layout

```text
cmd/vinyl-quoter/        CLI entrypoint
internal/app/            orchestration and flags
internal/catalog/        CSV rows, read/write, update/replace policy
internal/imageinput/     image discovery and supported extensions
internal/provider/       recognizer interface and model clients
internal/ui/             interactive menu
data/src/                input images (ignored except .gitkeep)
data/dst/                intermediate output placeholder
data/report/             CSV output (ignored except .gitkeep)
docker/test/             Go test/runtime container
docs/QUICKSTART.md       short usage guide
```

Supported image extensions: `.jpg`, `.jpeg`, `.png`, `.webp`, `.dng`, `.heic`, `.heif`, `.tif`, `.tiff`.

## Requirements

- Docker with Compose.
- For local recognition: LM Studio running at `http://localhost:1234/v1` with `qwen2.5-vl-7b-instruct` loaded.
- For Gemini recognition: `GEMINI_API_KEY` in the environment.

## Makefile commands

Start the test/runtime container before running app commands:

```bash
make test-build
make test-up
```

Common commands:

```bash
make run IMAGE=data/src/DSC01.jpg
make run-all
make run-all-replace
make run-gemini
make test
make quality
make test-shell
make test-down
```

All app and Go test commands run inside the `docker/test` container. Go caches are bind-mounted under the ignored project-local `.cache/` directory.

## Direct CLI usage inside the container

```bash
go run ./cmd/vinyl-quoter --all
go run ./cmd/vinyl-quoter --image data/src/DSC01.jpg
go run ./cmd/vinyl-quoter --all --replace
go run ./cmd/vinyl-quoter --all --provider gemini
```

Important flags:

- `--src`: source image directory, default `data/src`
- `--report`: CSV path, default `data/report/album_catalog.csv`
- `--provider`: `lm-studio` or `gemini`
- `--model`: model override
- `--lm-studio-base-url`: default `http://localhost:1234/v1`

## CSV output

Default path: `data/report/album_catalog.csv`.

Columns:

- `source_image`
- `artist`
- `title`
- `identification_confidence`
- `recommended_price_eur`
- `price_confidence`
- `price_basis`
- `notes`

Update mode keeps existing rows and skips images already present in the CSV. Replace mode regenerates the CSV from scratch.

## AI demo video

Generate a local one-minute demo video from `prompt.txt`:

```bash
make video-build
make video-generate
```

Output:

```text
data/video/vinylquoter-ai-demo.mp4
```

The video is generated locally with `ffmpeg` in the `docker/video` container. Generated video files are ignored by Git.

## Data and Git

`data/` is ignored except for `.gitkeep` placeholders in `data/src/`, `data/dst/`, `data/report/`, and `data/video/`. Do not commit source images, generated reports, or generated videos.

## Troubleshooting

- `GEMINI_API_KEY is required`: choose LM Studio or export a Gemini API key.
- `LM Studio request failed`: start LM Studio, load the vision model, and enable the local server.
- Missing rows in update mode: run `make run-all-replace`.
