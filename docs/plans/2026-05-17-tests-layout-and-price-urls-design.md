# Tests Layout and Price Reference URLs Design

## Context

The project currently keeps Go unit tests next to the implementation under `internal/**`. The new requirement is to keep unit tests under `tests/` while preserving the implementation path shape. The CSV currently contains pricing fields but no reference URLs for external price checks.

## Goals

- Move unit tests to `tests/` with mirrored paths, for example `internal/app/app.go` maps to `tests/internal/app/app_test.go`.
- Keep tests runnable with `go test ./...` through the existing Docker-backed `make test` target.
- Add three CSV columns with deterministic price-reference search URLs for a medium-high condition assumption.
- Use generic platform search URLs, not model-generated URLs.

## Test Layout Design

Because Go packages under `tests/internal/...` are outside the implementation directories, tests will use external package names such as `app_test` and import implementation packages like `vinylquoter/internal/app`. This is valid because the `tests/` tree is still inside the same parent module and can import `internal` packages.

Tests that currently depend on unexported functions must be handled by testing public behavior instead. `internal/crop` already exposes `Process`, `internal/catalog` exposes CSV helpers, providers expose parse or request behavior, and `visionpayload` exposes `Prompt`, `InlineImage`, and `DataURL`.

## CSV URL Design

Add three columns:

- `discogs_reference_url`
- `ebay_reference_url`
- `popsike_reference_url`

The app will fill them after recognition using artist and title. The query will include a condition hint for medium-high vinyl and sleeve quality:

```text
<artist> <title> vinyl VG+ sleeve VG+
```

The generated URLs will be deterministic and URL-encoded:

- Discogs: `https://www.discogs.com/search/?q=<query>&type=all`
- eBay: `https://www.ebay.es/sch/i.html?_nkw=<query>`
- Popsike: `https://www.popsike.com/php/quicksearch.php?searchtext=<query>`

If artist or title is missing or `Unknown`, the URL builder will still use the available non-empty terms plus the condition hint. This keeps CSV output useful for manual review without requiring provider-generated links.

## Compatibility

CSV reads remain backward compatible: rows with the old eight-column header are padded with empty URL fields. New writes always emit the expanded header.

The model JSON contract does not need to return URLs. URL generation belongs in deterministic Go code so both LM Studio and Gemini behave the same.

## Verification

- `make test` must continue to run all packages, including the new `tests/...` packages.
- CSV tests must cover old CSV rows, new header fields, and URL upsert/write behavior.
- App tests must prove processed rows include the three URL fields.
- Documentation guard tests must reflect the new `tests/` layout and CSV columns.
