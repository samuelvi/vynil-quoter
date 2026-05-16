# VinylQuoter Development

Use this guide for tests, Docker runtime, and quality checks. User-facing CLI usage lives in `docs/QUICKSTART.md`.

## Docker app/test runtime

Most Make targets auto-build Docker image dependencies before running. You can still build it explicitly:

```bash
make docker-build
```

Run the app inside Docker:

```bash
make run
make run IMAGE=data/src/DSC01.jpg
make run-all
make run-all-replace
make run-gemini
```

No manual rebuild is needed for `make run`: it uses `go run` inside Docker, so Go recompiles changed code automatically. Use `make build` only when you need a binary at `bin/vinyl-quoter`.

Open a shell:

```bash
make docker-shell
```

If you start the long-running runtime container, stop it with:

```bash
make docker-up
make docker-down
```

Backward-compatible aliases remain available: `make test-build`, `make test-up`, `make test-down`, and `make test-shell`.

## Tests and quality

Run Go tests inside the container. `make test` auto-build Docker image dependencies before running:

```bash
make test
```

Run tests plus the strict quick quality gate:

```bash
make quality
```

## Build

Compile the shell executable inside Docker:

```bash
make build
```
