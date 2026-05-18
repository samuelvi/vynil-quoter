# VinylQuoter

VinylQuoter identifies vinyl records from cover images in `data/src/` and writes a valuation catalog to `data/report/album_catalog.csv`.

For the fastest path, read [docs/QUICKSTART.md](docs/QUICKSTART.md). For all project documentation, including active implementation plans, use [docs/index.md](docs/index.md).

## What it does

- Scans one image or every supported image in `data/src/`.
- crops source images into `data/dst` as JPG files to isolate the vinyl/cover from noisy backgrounds.
- Identifies album title and artist/band using LM Studio or Gemini; it analyzes the cropped image from `data/dst`.
- Estimates a conservative EUR resale value.
- Writes a CSV catalog with the image file name (for example `DSC01.jpg`) as the row identifier, plus confidence and review notes.

Default provider: LM Studio local model `qwen2.5-vl-7b-instruct`. An alternate local vision option, `gemma-3-4b-it`, is available from the interactive menu or with `--model` when loaded in LM Studio.

## Project layout

```text
cmd/vinyl-quoter/        CLI entrypoint
internal/app/            orchestration and flags
internal/catalog/        CSV rows, read/write, update/replace policy
internal/crop/           local crop pipeline from source photos to data/dst
internal/imageinput/     image discovery and supported extensions
internal/provider/       recognizer interface and model clients
internal/ui/             interactive menu
data/src/                input images (ignored except .gitkeep)
data/dst/                cropped images generated from data/src
data/report/             CSV output (ignored except .gitkeep)
docker/test/             Go test/runtime container
tests/                   unit tests mirroring implementation paths
docs/QUICKSTART.md       user usage guide
docs/DEVELOPMENT.md      development, Docker, and test guide
```

Supported image extensions: `.jpg`, `.jpeg`, `.png`, `.webp`, `.dng`, `.heic`, `.heif`, `.tif`, `.tiff`.

## Requirements

- Docker with Compose.
- For local recognition: LM Studio running on the host with `qwen2.5-vl-7b-instruct` loaded. Docker commands use `http://host.docker.internal:1234/v1` by default.
- For Gemini recognition: `GEMINI_API_KEY` in the environment.

## Makefile commands

App commands prepare Docker automatically. You can start with:

```bash
make run
```

Common commands:

```bash
make build
make run
make run IMAGE=data/src/DSC01.jpg
make run-all
make run-all-replace
make run-gemini
make run-cli ARGS="--all --provider lm-studio"
make test
make quality
make test-shell
make test-down
```

All app and Go test commands run inside Docker and auto-build the runtime image when needed. No manual rebuild is needed for normal use: `go run` runs inside Docker and Go recompiles changed code automatically inside Docker. Go caches are bind-mounted under the ignored project-local `.cache/` directory.

## Interactive menu

`make run` opens a persistent menu. It returns to the main menu after each action and exits only when you choose `Salir`.

Main menu:

1. Process one image.
2. Process every supported image in `data/src`.
3. Open `Guardado csv (<current CSV path>)`.
4. Open `Modelo (<current provider: model>)`.
5. Open `Calidad carátula (<current sleeve condition>)`.
6. Open `Calidad vinilo (<current media condition>)`.
7. Exit.

`Guardado csv` lets you change the current CSV path, update the current CSV, regenerate it from scratch, or go back. It does not ask for the recognition model. `Modelo` changes the model once and keeps that selection for later actions. The two quality menus use Discogs/Goldmine-style grades (`M`, `NM/M-`, `VG+`, `VG`, `G+`, `G`, `F`, `P`; sleeve also supports `Generic`) and default both media and sleeve to `VG`.

## Raw CLI flags through Make

Use `make run-cli` for flags not wrapped by the common Make commands:

```bash
make run-cli ARGS="--all --provider lm-studio --model qwen2.5-vl-7b-instruct"
```

Important flags:

- `--src`: source image directory, default `data/src`
- `--dst`: cropped image directory, default `data/dst`
- `--report`: CSV path, default `data/report/album_catalog.csv`
- `--provider`: `lm-studio` or `gemini`
- `--model`: model override
- `--media-condition`: vinyl/media grade, default `VG`
- `--sleeve-condition`: sleeve/cover grade, default `VG`
- `--lm-studio-base-url`: default `http://localhost:1234/v1`

## CSV output

Processing pipeline: `data/src → data/dst → model → CSV`.

Decodable inputs are locally cropped and written as `.jpg` before recognition. Unsupported decode formats are copied to `data/dst` as a fallback so the model can still inspect them.

Default path: `data/report/album_catalog.csv`.

Columns:

- `source_image` — image file name used as ID, for example `DSC01.jpg`
- `artist`
- `title`
- `identification_confidence`
- `recommended_price_eur`
- `condition`
- `price_confidence`
- `price_basis`
- `notes`
- `discogs_reference_url`
- `ebay_reference_url`
- `popsike_reference_url`

Normal processing reprocesses the selected source images every time, overwrites their prepared files in `data/dst`, and upserts CSV rows by `source_image`. Rows for images not selected in the current run are preserved. Replace mode regenerates the CSV from scratch only for all-images runs; single-image CLI runs always update/create the CSV and never replace it.

`recommended_price_eur` contains only numbers or numeric ranges such as `12` or `12-18`; the currency is expressed by the column name. `condition` stores the selected grades as `media: <grade>; sleeve: <grade>`.

Reference URL columns are broad Discogs, eBay, and Popsike search links generated from artist/title. They intentionally avoid condition terms such as `VG+ sleeve VG+` so searches return more results.

## Data and Git

`data/` is ignored except for `.gitkeep` placeholders in `data/src/`, `data/dst/`, `data/report/`, and `data/video/`. Do not commit source images, generated reports, or generated videos.

## Troubleshooting

- `GEMINI_API_KEY is required`: choose LM Studio or export a Gemini API key.
- `LM Studio request failed`: start LM Studio, load the vision model, and enable the local server.
- Stale or missing rows: run `make run-all` to refresh every supported image in `data/src`, or `make run-all-replace` to discard rows for images no longer present and regenerate the CSV from current inputs.
