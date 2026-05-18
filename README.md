# VinylQuoter

VinylQuoter is a Docker-first Go CLI that identifies vinyl records from cover images and writes a resale catalog to CSV.

The app prepares images locally before recognition:

```text
data/src → data/dst → vision model → data/report/album_catalog.csv
```

## Quick start

1. Put images in `data/src/`.
2. Start the interactive menu:

```bash
make run
```

3. Review the CSV at `data/report/album_catalog.csv`.

Docker is the supported runtime for app and test commands. You do not need Go installed on the host for normal use.
No manual rebuild is needed for normal use: Go recompiles changed code automatically inside Docker.

## Common commands

```bash
make build                       # compile bin/vinyl-quoter when needed
make run                         # interactive menu
make run IMAGE=data/src/DSC01.jpg # process one image
make run-all                     # update CSV from all images
make run-all-replace             # regenerate CSV from all images
make run-gemini                  # use Gemini provider
make run-cli ARGS="--all --provider lm-studio"
make test                        # Go tests inside Docker
make quality                     # tests + quality gate
```

## What it does

- Discovers supported images under `data/src/`.
- Crops decodable images into JPG files under `data/dst/`.
- The crop pipeline crops source images into `data/dst` and analyzes the cropped image.
- Copies supported but non-decodable inputs to `data/dst/` as fallback.
- Sends the prepared `data/dst/` image to LM Studio or Gemini.
- Captures artist, title, confidence, conservative EUR price, selected media/sleeve condition, notes, and broad reference URLs.
- Updates existing CSV rows by `source_image` without duplicating records.

Default local provider: LM Studio with `qwen2.5-vl-7b-instruct`.

Interactive state is changed through `Guardado csv` and `Modelo (<current provider: model>)`. CSV reference columns include `discogs_reference_url`, `ebay_reference_url`, and `popsike_reference_url`.

## Documentation

- [Quickstart](docs/QUICKSTART.md) — fastest path to run the app.
- [Usage Guide](docs/USAGE.md) — menu, flags, providers, conditions, and data flow.
- [CSV Output](docs/CSV.md) — columns, update/replace policy, and reference URL behavior.
- [Development Guide](docs/DEVELOPMENT.md) — Docker runtime, tests, quality checks, and project layout.
- [Documentation Index](docs/index.md) — all maintained docs and active plans.

## Project layout

```text
cmd/vinyl-quoter/        CLI entrypoint
internal/app/            orchestration, flags, processing pipeline
internal/catalog/        CSV rows, read/write, upsert, reference URLs
internal/config/         defaults, providers, conditions
internal/crop/           local image preparation into data/dst
internal/imageinput/     image discovery and supported extensions
internal/provider/       recognizer interface and model clients
internal/ui/             interactive menu
tests/                   Go tests mirroring implementation paths
docker/test/             Docker runtime for app and tests
data/src/                input images, ignored except .gitkeep
data/dst/                prepared images, ignored except .gitkeep
data/report/             CSV reports, ignored except .gitkeep
```

## Data and Git

`data/` contents are ignored except `.gitkeep` placeholders. Do not commit source images, cropped images, generated reports, or generated videos.
