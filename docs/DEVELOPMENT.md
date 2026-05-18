# VinylQuoter Development Guide

Use this guide for project structure, Docker runtime, tests, quality checks, and development commands. User-facing behavior lives in [Usage Guide](USAGE.md).

## Stack

- Language: Go `1.23`.
- Runtime: Docker Compose under `docker/test`.
- Entry point: `cmd/vinyl-quoter/main.go`.
- Tests: Go tests under `tests/`; the layout mirrors implementation paths.
- Command interface: `Makefile`.

## Project layout

```text
cmd/vinyl-quoter/        CLI entrypoint
internal/app/            argument parsing, menu loop, processing orchestration
internal/catalog/        CSV format, read/write, row upsert, reference URLs
internal/config/         defaults for paths, providers, models, and conditions
internal/crop/           image preparation into data/dst
internal/imageinput/     supported image discovery
internal/provider/       recognizer interface and provider clients
internal/provider/gemini/ Gemini implementation
internal/provider/lmstudio/ LM Studio implementation
internal/provider/jsonparse/ robust JSON extraction from model responses
internal/provider/visionpayload/ shared prompt and image payload creation
internal/ui/             interactive menu
tests/                   package tests mirroring internal paths
docker/test/             Docker runtime used for app and tests
data/                    ignored runtime input/output directories
```

## Docker runtime

The Docker service is `go-scripts` in `docker/test/docker-compose.yml`.

It bind-mounts:

- repository root to `/workspace`
- `.cache/go` to `/go/pkg/mod`
- `.cache/go-build` to `/root/.cache/go-build`

Build explicitly when needed:

```bash
make docker-build
```

Open a shell:

```bash
make docker-shell
```

Start or stop the long-running container:

```bash
make docker-up
make docker-down
```

Backward-compatible aliases remain available: `make test-build`, `make test-up`, `make test-down`, and `make test-shell`.

## App commands

Most app targets depend on `docker-build`, so they prepare Docker automatically and auto-build Docker image dependencies when needed.

```bash
make build
make run
make run IMAGE=data/src/DSC01.jpg
make run-all
make run-all-replace
make run-gemini
make run-cli ARGS="--all --provider lm-studio --media-condition VG+ --sleeve-condition VG"
```

No manual rebuild is needed for `make run`. The app uses `go run` inside Docker, and Go recompiles changed code automatically.

Use `make build` only when you need a binary at `bin/vinyl-quoter`.

## Tests and quality

Run Go tests inside Docker:

```bash
make test
```

Run tests plus the strict quick quality gate:

```bash
make quality
```

The quality target runs the repository quality gate configured by the local automation tooling. That tooling is not product code.

## Data hygiene

`data/` is ignored except `.gitkeep` placeholders in:

- `data/src/`
- `data/dst/`
- `data/report/`
- `data/video/`

Do not commit source images, cropped outputs, generated CSV reports, or generated videos.

## Environment files

`.env` and `.env.*` are ignored because they may contain local machine paths or secrets. Commit only `.env.example`. Runtime defaults are loaded from built-in values, then `.env`, then process environment variables, then CLI flags.

## Documentation maintenance

Keep the root `README.md` lightweight. Add detail to focused files under `docs/` and link every maintained doc from `docs/index.md`.
