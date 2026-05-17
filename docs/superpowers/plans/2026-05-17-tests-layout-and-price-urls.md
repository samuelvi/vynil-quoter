# Tests Layout and Price Reference URLs Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Move Go unit tests into mirrored `tests/` paths and add three deterministic price-reference URL columns to the CSV.

**Architecture:** Tests become external packages under `tests/internal/...` and import public APIs from `vinylquoter/internal/...`. CSV rows gain three URL fields generated in Go from `artist + title + vinyl VG+ sleeve VG+`, so menu and CLI outputs stay deterministic across providers.

**Tech Stack:** Go 1.23, standard library `net/url`, standard CSV package, Docker-backed Make targets (`make test`, `make quality`).

---

## File Structure

- Move: `internal/app/app_test.go` → `tests/internal/app/app_test.go`
- Move: `internal/catalog/csv_test.go` → `tests/internal/catalog/csv_test.go`
- Move: `internal/crop/crop_test.go` → `tests/internal/crop/crop_test.go`
- Move: `internal/imageinput/images_test.go` → `tests/internal/imageinput/images_test.go`
- Move: `internal/projectfiles/project_test.go` → `tests/internal/projectfiles/project_test.go`
- Move: `internal/provider/gemini/gemini_test.go` → `tests/internal/provider/gemini/gemini_test.go`
- Move: `internal/provider/lmstudio/lmstudio_test.go` → `tests/internal/provider/lmstudio/lmstudio_test.go`
- Move: `internal/provider/visionpayload/visionpayload_test.go` → `tests/internal/provider/visionpayload/visionpayload_test.go`
- Move: `internal/ui/menu_test.go` → `tests/internal/ui/menu_test.go`
- Modify: `internal/catalog/row.go` to add URL fields to `Identification`, `Row`, and `Header`.
- Modify: `internal/catalog/csv.go` to read/write eleven columns.
- Create: `internal/catalog/urls.go` for deterministic reference URL generation.
- Modify: `internal/app/app.go` to fill URL fields before upsert/write.
- Modify: `internal/provider/visionpayload/visionpayload.go` only if prompt tests require removing URL expectations; current design does not require provider URLs.
- Modify: `README.md`, `docs/QUICKSTART.md`, `docs/DEVELOPMENT.md`, and `docs/index.md` to document `tests/` and URL columns.

---

### Task 1: Move tests to mirrored `tests/` packages

**Files:**
- Create under `tests/internal/**`: moved test files.
- Delete old `internal/**/*_test.go` files.

- [ ] **Step 1: Move test files without changing assertions**

Run these filesystem moves:

```bash
mkdir -p tests/internal/app tests/internal/catalog tests/internal/crop tests/internal/imageinput tests/internal/projectfiles tests/internal/provider/gemini tests/internal/provider/lmstudio tests/internal/provider/visionpayload tests/internal/ui
mv internal/app/app_test.go tests/internal/app/app_test.go
mv internal/catalog/csv_test.go tests/internal/catalog/csv_test.go
mv internal/crop/crop_test.go tests/internal/crop/crop_test.go
mv internal/imageinput/images_test.go tests/internal/imageinput/images_test.go
mv internal/projectfiles/project_test.go tests/internal/projectfiles/project_test.go
mv internal/provider/gemini/gemini_test.go tests/internal/provider/gemini/gemini_test.go
mv internal/provider/lmstudio/lmstudio_test.go tests/internal/provider/lmstudio/lmstudio_test.go
mv internal/provider/visionpayload/visionpayload_test.go tests/internal/provider/visionpayload/visionpayload_test.go
mv internal/ui/menu_test.go tests/internal/ui/menu_test.go
```

- [ ] **Step 2: Convert package names and imports**

For each moved file, change package and direct calls:

- `tests/internal/app/app_test.go`: `package app_test`; import `vinylquoter/internal/app`; call `app.ParseArgs`, `app.Process`. Keep local fakes in the test file.
- `tests/internal/catalog/csv_test.go`: `package catalog_test`; import `vinylquoter/internal/catalog`; prefix `catalog.Row`, `catalog.Write`, `catalog.Read`, `catalog.Upsert`.
- `tests/internal/crop/crop_test.go`: `package crop_test`; import `vinylquoter/internal/crop`; call `crop.Process`.
- `tests/internal/imageinput/images_test.go`: `package imageinput_test`; import `vinylquoter/internal/imageinput`; prefix public calls.
- `tests/internal/projectfiles/project_test.go`: `package projectfiles_test`; keep helper functions local; no implementation import needed.
- `tests/internal/provider/gemini/gemini_test.go`: `package gemini_test`; import `vinylquoter/internal/provider/gemini`; call `gemini.ParseResponse` and `gemini.Client`.
- `tests/internal/provider/lmstudio/lmstudio_test.go`: `package lmstudio_test`; import `vinylquoter/internal/provider/lmstudio`; call `lmstudio.ParseResponse` and `lmstudio.Client`.
- `tests/internal/provider/visionpayload/visionpayload_test.go`: `package visionpayload_test`; import `vinylquoter/internal/provider/visionpayload`; call `visionpayload.Prompt`, `visionpayload.InlineImage`.
- `tests/internal/ui/menu_test.go`: `package ui_test`; import `vinylquoter/internal/ui`; call `ui.ReadMenu`, `ui.ReadMenuWithState`, compare `ui.ErrNoAction`.

- [ ] **Step 3: Run tests to expose migration misses**

Run:

```bash
make test
```

Expected initial result: FAIL if any moved test still references unqualified implementation symbols. Fix only package/import issues until test behavior is unchanged.

- [ ] **Step 4: Verify test layout is clean**

Run:

```bash
python3 - <<'PY'
from pathlib import Path
old = sorted(Path('internal').rglob('*_test.go'))
if old:
    raise SystemExit('tests still under internal/: ' + ', '.join(str(p) for p in old))
expected = [
    Path('tests/internal/app/app_test.go'),
    Path('tests/internal/catalog/csv_test.go'),
    Path('tests/internal/crop/crop_test.go'),
    Path('tests/internal/imageinput/images_test.go'),
    Path('tests/internal/projectfiles/project_test.go'),
    Path('tests/internal/provider/gemini/gemini_test.go'),
    Path('tests/internal/provider/lmstudio/lmstudio_test.go'),
    Path('tests/internal/provider/visionpayload/visionpayload_test.go'),
    Path('tests/internal/ui/menu_test.go'),
]
missing = [str(p) for p in expected if not p.exists()]
if missing:
    raise SystemExit('missing moved tests: ' + ', '.join(missing))
print('test layout ok')
PY
```

Expected: `test layout ok`.

---

### Task 2: Add CSV URL fields and deterministic URL builder

**Files:**
- Modify: `internal/catalog/row.go`
- Modify: `internal/catalog/csv.go`
- Create: `internal/catalog/urls.go`
- Modify: `tests/internal/catalog/csv_test.go`

- [ ] **Step 1: Write failing catalog URL tests**

Add these tests to `tests/internal/catalog/csv_test.go`:

```go
func TestReferenceURLsUseArtistTitleAndConditionHint(t *testing.T) {
	refs := catalog.ReferenceURLs("The Cure", "Disintegration")

	for name, got := range map[string]string{
		"discogs": refs.Discogs,
		"ebay":    refs.EBay,
		"popsike": refs.Popsike,
	} {
		if !strings.Contains(got, "The+Cure+Disintegration+vinyl+VG%2B+sleeve+VG%2B") {
			t.Fatalf("%s URL missing encoded pricing query: %s", name, got)
		}
	}
	if !strings.HasPrefix(refs.Discogs, "https://www.discogs.com/search/") {
		t.Fatalf("unexpected Discogs URL: %s", refs.Discogs)
	}
	if !strings.HasPrefix(refs.EBay, "https://www.ebay.es/sch/i.html") {
		t.Fatalf("unexpected eBay URL: %s", refs.EBay)
	}
	if !strings.HasPrefix(refs.Popsike, "https://www.popsike.com/php/quicksearch.php") {
		t.Fatalf("unexpected Popsike URL: %s", refs.Popsike)
	}
}

func TestWriteReadRowsIncludesReferenceURLColumns(t *testing.T) {
	tmp := t.TempDir()
	report := filepath.Join(tmp, "data", "report", "album_catalog.csv")
	row := catalog.Row{SourceImage: "DSC01.jpg", Artist: "The Cure", Title: "Disintegration", DiscogsReferenceURL: "https://discogs.example", EBayReferenceURL: "https://ebay.example", PopsikeReferenceURL: "https://popsike.example"}

	if err := catalog.Write(report, []catalog.Row{row}); err != nil {
		t.Fatal(err)
	}
	content, err := os.ReadFile(report)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"discogs_reference_url", "ebay_reference_url", "popsike_reference_url"} {
		if !strings.Contains(string(content), want) {
			t.Fatalf("CSV header missing %s: %s", want, string(content))
		}
	}
	rows, err := catalog.Read(report)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].DiscogsReferenceURL != row.DiscogsReferenceURL || rows[0].EBayReferenceURL != row.EBayReferenceURL || rows[0].PopsikeReferenceURL != row.PopsikeReferenceURL {
		t.Fatalf("got %#v", rows)
	}
}

func TestReadOldEightColumnCSVKeepsEmptyReferenceURLs(t *testing.T) {
	tmp := t.TempDir()
	report := filepath.Join(tmp, "album_catalog.csv")
	oldCSV := "source_image,artist,title,identification_confidence,recommended_price_eur,price_confidence,price_basis,notes\nDSC01.jpg,The Cure,Disintegration,high,22,medium,basis,notes\n"
	if err := os.WriteFile(report, []byte(oldCSV), 0o644); err != nil {
		t.Fatal(err)
	}

	rows, err := catalog.Read(report)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 {
		t.Fatalf("got %#v", rows)
	}
	if rows[0].DiscogsReferenceURL != "" || rows[0].EBayReferenceURL != "" || rows[0].PopsikeReferenceURL != "" {
		t.Fatalf("old CSV rows should have empty reference URLs, got %#v", rows[0])
	}
}
```

Add imports `os` and `strings` to the catalog test file.

- [ ] **Step 2: Run catalog tests to verify RED**

Run:

```bash
docker compose -f docker/test/docker-compose.yml run --rm go-scripts go test ./tests/internal/catalog
```

Expected: FAIL with undefined `catalog.ReferenceURLs` and missing row fields.

- [ ] **Step 3: Implement catalog URL fields**

Update `internal/catalog/row.go`:

```go
package catalog

type Identification struct {
	Artist                   string `json:"artist"`
	Title                    string `json:"title"`
	IdentificationConfidence string `json:"identification_confidence"`
	RecommendedPriceEUR      string `json:"recommended_price_eur"`
	PriceConfidence          string `json:"price_confidence"`
	PriceBasis               string `json:"price_basis"`
	Notes                    string `json:"notes"`
}

type Row struct {
	SourceImage              string
	Artist                   string
	Title                    string
	IdentificationConfidence string
	RecommendedPriceEUR      string
	PriceConfidence          string
	PriceBasis               string
	Notes                    string
	DiscogsReferenceURL      string
	EBayReferenceURL         string
	PopsikeReferenceURL      string
}

type ReferenceLinks struct {
	Discogs string
	EBay    string
	Popsike string
}

var Header = []string{"source_image", "artist", "title", "identification_confidence", "recommended_price_eur", "price_confidence", "price_basis", "notes", "discogs_reference_url", "ebay_reference_url", "popsike_reference_url"}
```

Create `internal/catalog/urls.go`:

```go
package catalog

import (
	"net/url"
	"strings"
)

func ReferenceURLs(artist string, title string) ReferenceLinks {
	query := referenceQuery(artist, title)
	encoded := url.QueryEscape(query)
	return ReferenceLinks{
		Discogs: "https://www.discogs.com/search/?q=" + encoded + "&type=all",
		EBay:    "https://www.ebay.es/sch/i.html?_nkw=" + encoded,
		Popsike: "https://www.popsike.com/php/quicksearch.php?searchtext=" + encoded,
	}
}

func referenceQuery(artist string, title string) string {
	parts := make([]string, 0, 5)
	for _, value := range []string{artist, title} {
		cleaned := strings.TrimSpace(value)
		if cleaned == "" || strings.EqualFold(cleaned, "Unknown") {
			continue
		}
		parts = append(parts, cleaned)
	}
	parts = append(parts, "vinyl", "VG+", "sleeve", "VG+")
	return strings.Join(parts, " ")
}
```

- [ ] **Step 4: Update CSV read/write for eleven columns**

In `internal/catalog/csv.go`, change row construction to include the three URL fields:

```go
rows = append(rows, Row{record[0], record[1], record[2], record[3], record[4], record[5], record[6], record[7], record[8], record[9], record[10]})
```

Change writer row output to include all eleven fields:

```go
if err := writer.Write([]string{row.SourceImage, row.Artist, row.Title, row.IdentificationConfidence, row.RecommendedPriceEUR, row.PriceConfidence, row.PriceBasis, row.Notes, row.DiscogsReferenceURL, row.EBayReferenceURL, row.PopsikeReferenceURL}); err != nil {
	return err
}
```

Keep the existing padding loop so old eight-column CSV files remain readable.

- [ ] **Step 5: Run catalog tests to verify GREEN**

Run:

```bash
docker compose -f docker/test/docker-compose.yml run --rm go-scripts go test ./tests/internal/catalog
```

Expected: PASS.

---

### Task 3: Fill URL fields from app processing

**Files:**
- Modify: `internal/app/app.go`
- Modify: `tests/internal/app/app_test.go`

- [ ] **Step 1: Add failing app test for URL fields**

Add this test to `tests/internal/app/app_test.go`:

```go
func TestProcessAddsPriceReferenceURLs(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "data", "src", "DSC01.jpg")
	writeTinyJPEG(t, src)
	report := filepath.Join(tmp, "data", "report", "album_catalog.csv")
	dstDir := filepath.Join(tmp, "data", "dst")

	rows, err := app.Process(context.Background(), []string{src}, report, false, dstDir, recognizerFunc(func(ctx context.Context, imagePath string) (catalog.Identification, error) {
		return catalog.Identification{Artist: "The Cure", Title: "Disintegration"}, nil
	}))

	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 {
		t.Fatalf("got %#v", rows)
	}
	for name, got := range map[string]string{
		"discogs": rows[0].DiscogsReferenceURL,
		"ebay":    rows[0].EBayReferenceURL,
		"popsike": rows[0].PopsikeReferenceURL,
	} {
		if !strings.Contains(got, "The+Cure+Disintegration+vinyl+VG%2B+sleeve+VG%2B") {
			t.Fatalf("%s URL missing expected query: %s", name, got)
		}
	}
}
```

- [ ] **Step 2: Run app tests to verify RED**

Run:

```bash
docker compose -f docker/test/docker-compose.yml run --rm go-scripts go test ./tests/internal/app
```

Expected: FAIL because URL fields are empty.

- [ ] **Step 3: Fill URLs in `Process`**

In `internal/app/app.go`, after crop/provider notes and before creating `catalog.Row`, add:

```go
referenceURLs := catalog.ReferenceURLs(identification.Artist, identification.Title)
```

Then include URL fields in the row literal:

```go
row := catalog.Row{
	SourceImage:              catalog.ImageID(image),
	Artist:                   identification.Artist,
	Title:                    identification.Title,
	IdentificationConfidence: identification.IdentificationConfidence,
	RecommendedPriceEUR:      identification.RecommendedPriceEUR,
	PriceConfidence:          identification.PriceConfidence,
	PriceBasis:               identification.PriceBasis,
	Notes:                    identification.Notes,
	DiscogsReferenceURL:      referenceURLs.Discogs,
	EBayReferenceURL:         referenceURLs.EBay,
	PopsikeReferenceURL:      referenceURLs.Popsike,
}
```

- [ ] **Step 4: Run app tests to verify GREEN**

Run:

```bash
docker compose -f docker/test/docker-compose.yml run --rm go-scripts go test ./tests/internal/app
```

Expected: PASS.

---

### Task 4: Update project documentation and guards

**Files:**
- Modify: `README.md`
- Modify: `docs/QUICKSTART.md`
- Modify: `docs/DEVELOPMENT.md`
- Modify: `docs/index.md`
- Modify: `tests/internal/projectfiles/project_test.go`

- [ ] **Step 1: Update documentation guard test**

In `tests/internal/projectfiles/project_test.go`, update documentation assertions to require:

```go
"tests/"
"discogs_reference_url"
"ebay_reference_url"
"popsike_reference_url"
```

Add those strings to `TestReadmeDocumentsCLIAndDevelopmentDocs`, `TestQuickstartDocumentsInteractiveAndFlagUsage`, and `TestDevelopmentDocDocumentsTestsAndDocker` where appropriate.

- [ ] **Step 2: Run projectfiles tests to verify RED**

Run:

```bash
docker compose -f docker/test/docker-compose.yml run --rm go-scripts go test ./tests/internal/projectfiles
```

Expected: FAIL until docs mention the new layout and columns.

- [ ] **Step 3: Update docs**

Update `README.md`:

- In project layout, add `tests/` as mirrored unit tests.
- In CSV columns, add the three URL columns.
- Add one sentence that URL references are generic searches for Discogs, eBay, and Popsike using medium-high quality assumptions.

Update `docs/QUICKSTART.md`:

- Under CSV review, list the three new URL columns.
- Explain they are search links for external price comparison, generated from artist/title with `vinyl VG+ sleeve VG+`.

Update `docs/DEVELOPMENT.md`:

- Mention tests live under `tests/` mirroring implementation paths.
- Keep `make test` as the verification command.

Update `docs/index.md`:

- Mention the implementation plan/design doc for this change.

- [ ] **Step 4: Run projectfiles tests to verify GREEN**

Run:

```bash
docker compose -f docker/test/docker-compose.yml run --rm go-scripts go test ./tests/internal/projectfiles
```

Expected: PASS.

---

### Task 5: Final verification and commit

**Files:**
- All changed files from Tasks 1-4.

- [ ] **Step 1: Format Go files**

Run:

```bash
docker compose -f docker/test/docker-compose.yml run --rm go-scripts gofmt -w internal/catalog/row.go internal/catalog/csv.go internal/catalog/urls.go internal/app/app.go tests/internal/app/app_test.go tests/internal/catalog/csv_test.go tests/internal/crop/crop_test.go tests/internal/imageinput/images_test.go tests/internal/projectfiles/project_test.go tests/internal/provider/gemini/gemini_test.go tests/internal/provider/lmstudio/lmstudio_test.go tests/internal/provider/visionpayload/visionpayload_test.go tests/internal/ui/menu_test.go
```

Expected: command exits 0.

- [ ] **Step 2: Run full tests**

Run:

```bash
make test
```

Expected: PASS for all packages, including `vinylquoter/tests/internal/...` packages.

- [ ] **Step 3: Run strict quality gate**

Run:

```bash
python3 .opencode/meta/hooks/quality_gate.py --workspace . --mode quick --strict
```

Expected: `"ok": true` and no `blocking_findings`.

- [ ] **Step 4: Inspect final diff**

Run:

```bash
git status --short --branch
git diff --stat
```

Expected: old `internal/**/*_test.go` files are deleted, mirrored `tests/internal/**/*_test.go` files are added, CSV/app/docs files are modified, and plan/design docs are added.

- [ ] **Step 5: Commit if requested**

If the user requests commit, run:

```bash
git add README.md docs/QUICKSTART.md docs/DEVELOPMENT.md docs/index.md docs/plans/2026-05-17-tests-layout-and-price-urls-design.md docs/superpowers/plans/2026-05-17-tests-layout-and-price-urls.md internal/app/app.go internal/catalog/row.go internal/catalog/csv.go internal/catalog/urls.go tests/internal internal/app/app_test.go internal/catalog/csv_test.go internal/crop/crop_test.go internal/imageinput/images_test.go internal/projectfiles/project_test.go internal/provider/gemini/gemini_test.go internal/provider/lmstudio/lmstudio_test.go internal/provider/visionpayload/visionpayload_test.go internal/ui/menu_test.go
git commit -m "feat(csv): add price reference urls"
```

Use the local git identity in a `Co-Authored-By` trailer when committing generated code.

---

## Self-Review

- Spec coverage: The plan covers mirrored `tests/` layout and the three generic platform URL columns.
- Placeholder scan: The plan names exact files, commands, expected failures, and code snippets.
- Type consistency: URL fields are consistently named `DiscogsReferenceURL`, `EBayReferenceURL`, and `PopsikeReferenceURL`; CSV headers use snake_case equivalents.
