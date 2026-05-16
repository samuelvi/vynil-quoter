# VinylQuoter Development

Use this guide for tests, Docker runtime, and quality checks. User-facing CLI usage lives in `docs/QUICKSTART.md`.

## Docker app/test runtime

Build the Docker image:

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

Run Go tests inside the container:

```bash
make test
```

Run tests plus the strict quick quality gate:

```bash
make quality
```

Run Go tests directly on the host:

```bash
go test ./...
```

## Build

Compile the shell executable:

```bash
make build
```

Equivalent direct command:

```bash
go build -o bin/vinyl-quoter ./cmd/vinyl-quoter
```
