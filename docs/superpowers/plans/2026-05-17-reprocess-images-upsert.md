# Reprocess Images and Upsert CSV Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make every menu and CLI processing entrypoint reprocess selected `src` images, overwrite prepared `dst` images, and upsert CSV rows instead of skipping existing entries.

**Architecture:** Keep one internal pipeline in `internal/app.Process()`. Replace pending-skip behavior with catalog-level row upsert by normalized image ID, preserving non-selected CSV rows when `replace=false` and preserving full regeneration when `replace=true`.

**Tech Stack:** Go 1.23, standard library CSV/image packages, package-local Go unit tests, Docker-backed Make targets (`make test`, `make quality`).

---

## File Structure

- Modify `internal/catalog/csv.go`: add `Upsert(rows []Row, row Row) []Row`; remove `Pending()` if no code still uses it.
- Modify `internal/catalog/csv_test.go`: replace pending-skip tests with read/write and upsert tests.
- Modify `internal/app/app.go`: process every collected image and upsert each generated row.
- Modify `internal/app/app_test.go`: update existing skip tests and interactive expectations; add coverage for destination overwrite.
- Modify `README.md`: document refresh/upsert behavior for normal processing.
- Modify `docs/QUICKSTART.md`: document that `make run-all` refreshes all source images and upserts the CSV.

---

### Task 1: Add catalog row upsert behavior

**Files:**
- Modify: `internal/catalog/csv_test.go`
- Modify: `internal/catalog/csv.go`

- [ ] **Step 1: Replace pending tests with read/write and upsert tests**

Edit `internal/catalog/csv_test.go` to this complete content:

```go
package catalog

import (
	"path/filepath"
	"testing"
)

func TestWriteAndReadRows(t *testing.T) {
	tmp := t.TempDir()
	report := filepath.Join(tmp, "data", "report", "album_catalog.csv")
	existing := Row{SourceImage: "data/src/a.jpg", Artist: "Artist A", Title: "Title A", IdentificationConfidence: "high", RecommendedPriceEUR: "12", PriceConfidence: "medium", PriceBasis: "existing"}
	if err := Write(report, []Row{existing}); err != nil {
		t.Fatal(err)
	}
	rows, err := Read(report)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].Artist != "Artist A" {
		t.Fatalf("got %#v", rows)
	}
}

func TestUpsertReplacesExistingRowByImageID(t *testing.T) {
	rows := []Row{{SourceImage: "data/src/DSC01.jpg", Artist: "Old", Title: "Old Title"}}
	fresh := Row{SourceImage: "DSC01.jpg", Artist: "Fresh", Title: "Fresh Title"}

	updated := Upsert(rows, fresh)

	if len(updated) != 1 {
		t.Fatalf("upsert should replace instead of append, got %#v", updated)
	}
	if updated[0].SourceImage != "DSC01.jpg" || updated[0].Artist != "Fresh" || updated[0].Title != "Fresh Title" {
		t.Fatalf("got %#v", updated[0])
	}
}

func TestUpsertAppendsMissingRowAndPreservesExistingOrder(t *testing.T) {
	rows := []Row{{SourceImage: "DSC01.jpg", Artist: "Artist 1"}}
	fresh := Row{SourceImage: "DSC02.jpg", Artist: "Artist 2"}

	updated := Upsert(rows, fresh)

	if len(updated) != 2 {
		t.Fatalf("got %#v", updated)
	}
	if updated[0].SourceImage != "DSC01.jpg" || updated[1].SourceImage != "DSC02.jpg" {
		t.Fatalf("upsert should preserve existing order and append new rows, got %#v", updated)
	}
}
```

- [ ] **Step 2: Run catalog tests to verify failure**

Run:

```bash
docker compose -f docker/test/docker-compose.yml run --rm go-scripts go test ./internal/catalog
```

Expected result before implementation: FAIL with `undefined: Upsert`.

- [ ] **Step 3: Implement `Upsert` and remove `Pending`**

Edit `internal/catalog/csv.go` so it contains this complete content:

```go
package catalog

import (
	"encoding/csv"
	"errors"
	"os"
	"path/filepath"
)

func ImageID(imagePath string) string {
	return filepath.Base(imagePath)
}

func Read(path string) ([]Row, error) {
	file, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer file.Close()
	records, err := csv.NewReader(file).ReadAll()
	if err != nil {
		return nil, err
	}
	rows := make([]Row, 0, len(records))
	for i, record := range records {
		if i == 0 {
			continue
		}
		for len(record) < len(Header) {
			record = append(record, "")
		}
		rows = append(rows, Row{record[0], record[1], record[2], record[3], record[4], record[5], record[6], record[7]})
	}
	return rows, nil
}

func Write(path string, rows []Row) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	writer := csv.NewWriter(file)
	if err := writer.Write(Header); err != nil {
		return err
	}
	for _, row := range rows {
		if err := writer.Write([]string{row.SourceImage, row.Artist, row.Title, row.IdentificationConfidence, row.RecommendedPriceEUR, row.PriceConfidence, row.PriceBasis, row.Notes}); err != nil {
			return err
		}
	}
	writer.Flush()
	return writer.Error()
}

func Upsert(rows []Row, row Row) []Row {
	rowID := ImageID(row.SourceImage)
	for index := range rows {
		if ImageID(rows[index].SourceImage) == rowID {
			rows[index] = row
			return rows
		}
	}
	return append(rows, row)
}
```

- [ ] **Step 4: Run catalog tests to verify pass**

Run:

```bash
docker compose -f docker/test/docker-compose.yml run --rm go-scripts go test ./internal/catalog
```

Expected: PASS.

- [ ] **Step 5: Commit Task 1 if commits are requested**

```bash
git add internal/catalog/csv.go internal/catalog/csv_test.go
git commit -m "fix(catalog): upsert rows by source image"
```

Skip this step unless the user has explicitly requested commits in this session.

---

### Task 2: Make `Process` reprocess every requested image

**Files:**
- Modify: `internal/app/app_test.go`
- Modify: `internal/app/app.go`

- [ ] **Step 1: Replace the old skip test with a reprocess/upsert test**

In `internal/app/app_test.go`, replace `TestProcessSkipsExistingBasenameIdentifier` with this test:

```go
func TestProcessReprocessesExistingBasenameIdentifierAndUpdatesCSV(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "data", "src", "DSC01.jpg")
	writeTinyJPEG(t, src)
	report := filepath.Join(tmp, "data", "report", "album_catalog.csv")
	dstDir := filepath.Join(tmp, "data", "dst")
	if err := catalog.Write(report, []catalog.Row{{SourceImage: "DSC01.jpg", Artist: "Existing", Title: "Old"}}); err != nil {
		t.Fatal(err)
	}
	identifyCalls := 0

	rows, err := Process(context.Background(), []string{src}, report, false, dstDir, recognizerFunc(func(ctx context.Context, imagePath string) (catalog.Identification, error) {
		identifyCalls++
		return catalog.Identification{Artist: "Fresh", Title: "New", IdentificationConfidence: "high"}, nil
	}))

	if err != nil {
		t.Fatal(err)
	}
	if identifyCalls != 1 {
		t.Fatalf("expected existing image to be recognized again, calls=%d", identifyCalls)
	}
	if len(rows) != 1 || rows[0].Artist != "Fresh" || rows[0].Title != "New" {
		t.Fatalf("got %#v", rows)
	}
	written, err := catalog.Read(report)
	if err != nil {
		t.Fatal(err)
	}
	if len(written) != 1 || written[0].Artist != "Fresh" || written[0].Title != "New" {
		t.Fatalf("CSV row should be replaced, got %#v", written)
	}
}
```

- [ ] **Step 2: Add a destination overwrite regression test**

In `internal/app/app_test.go`, add this test after `TestProcessReprocessesExistingBasenameIdentifierAndUpdatesCSV`:

```go
func TestProcessOverwritesDstImageWhenReprocessingExistingCSVRow(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "data", "src", "DSC01.jpg")
	writeTinyJPEG(t, src)
	report := filepath.Join(tmp, "data", "report", "album_catalog.csv")
	dstDir := filepath.Join(tmp, "data", "dst")
	dst := filepath.Join(dstDir, "DSC01.jpg")
	if err := os.MkdirAll(dstDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dst, []byte("stale dst image"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := catalog.Write(report, []catalog.Row{{SourceImage: "DSC01.jpg", Artist: "Existing"}}); err != nil {
		t.Fatal(err)
	}

	_, err := Process(context.Background(), []string{src}, report, false, dstDir, fakeRecognizer{})

	if err != nil {
		t.Fatal(err)
	}
	content, err := os.ReadFile(dst)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) == "stale dst image" {
		t.Fatal("dst image was not overwritten during reprocessing")
	}
}
```

- [ ] **Step 3: Update interactive repeated-action expectation**

In `TestRunInteractiveResetsActionModeAfterOptionTwo`, change this assertion:

```go
if strings.Join(seen, ",") != "DSC01.jpg,DSC02.jpg" {
	t.Fatalf("option 1 should not fail after option 2, got %#v", seen)
}
```

to:

```go
if strings.Join(seen, ",") != "DSC01.jpg,DSC02.jpg,DSC01.jpg" {
	t.Fatalf("option 1 should reprocess the selected image after option 2, got %#v", seen)
}
```

- [ ] **Step 4: Run app tests to verify failure**

Run:

```bash
docker compose -f docker/test/docker-compose.yml run --rm go-scripts go test ./internal/app
```

Expected before implementation: FAIL because existing CSV rows are still skipped and `identifyCalls` remains `0`.

- [ ] **Step 5: Update `Process` implementation**

In `internal/app/app.go`, replace the full `Process` function with:

```go
func Process(ctx context.Context, images []string, reportPath string, replace bool, destinationDir string, recognizer provider.Recognizer) ([]catalog.Row, error) {
	rows := []catalog.Row{}
	if !replace {
		existing, err := catalog.Read(reportPath)
		if err != nil {
			return nil, err
		}
		rows = existing
	}
	for _, image := range images {
		cropResult, cropErr := crop.Process(image, destinationDir)
		imageForRecognition := cropResult.CroppedPath
		if cropErr != nil {
			imageForRecognition = image
		}
		identification, err := recognizer.Identify(ctx, imageForRecognition)
		if err != nil {
			identification = catalog.Identification{Artist: "Unknown", Title: "Unknown", IdentificationConfidence: "manual-review", PriceConfidence: "manual-review", Notes: "identification failed: " + err.Error()}
		}
		if cropErr != nil {
			identification.Notes = strings.TrimSpace(identification.Notes + " crop failed: " + cropErr.Error())
		}
		row := catalog.Row{SourceImage: catalog.ImageID(image), Artist: identification.Artist, Title: identification.Title, IdentificationConfidence: identification.IdentificationConfidence, RecommendedPriceEUR: identification.RecommendedPriceEUR, PriceConfidence: identification.PriceConfidence, PriceBasis: identification.PriceBasis, Notes: identification.Notes}
		rows = catalog.Upsert(rows, row)
		if err := catalog.Write(reportPath, rows); err != nil {
			return nil, err
		}
	}
	if len(images) == 0 {
		if err := catalog.Write(reportPath, rows); err != nil {
			return nil, err
		}
	}
	return rows, nil
}
```

- [ ] **Step 6: Run app tests to verify pass**

Run:

```bash
docker compose -f docker/test/docker-compose.yml run --rm go-scripts go test ./internal/app
```

Expected: PASS.

- [ ] **Step 7: Run all Go tests**

Run:

```bash
make test
```

Expected: PASS for `go test ./...`.

- [ ] **Step 8: Commit Task 2 if commits are requested**

```bash
git add internal/app/app.go internal/app/app_test.go
git commit -m "fix(app): reprocess selected images"
```

Skip this step unless the user has explicitly requested commits in this session.

---

### Task 3: Update user-facing documentation

**Files:**
- Modify: `README.md`
- Modify: `docs/QUICKSTART.md`

- [ ] **Step 1: Update README behavior text**

In `README.md`, replace line text that currently says:

```markdown
Update mode keeps existing rows and skips images already present in the CSV. Replace mode regenerates the CSV from scratch only for all-images runs; single-image CLI runs always update/create the CSV and never replace it.
```

with:

```markdown
Normal processing reprocesses the selected source images every time, overwrites their prepared files in `data/dst`, and upserts CSV rows by `source_image`. Rows for images not selected in the current run are preserved. Replace mode regenerates the CSV from scratch only for all-images runs; single-image CLI runs always update/create the CSV and never replace it.
```

Also replace the troubleshooting bullet:

```markdown
- Missing rows in update mode: run `make run-all-replace`.
```

with:

```markdown
- Stale or missing rows: run `make run-all` to refresh every supported image in `data/src`, or `make run-all-replace` to discard rows for images no longer present and regenerate the CSV from current inputs.
```

- [ ] **Step 2: Update Quickstart `make run-all` wording**

In `docs/QUICKSTART.md`, replace:

```markdown
Process every supported image in `data/src` and append missing rows:
```

with:

```markdown
Process every supported image in `data/src`, overwrite prepared images in `data/dst`, and upsert CSV rows:
```

After the `make run-all` command block, add:

```markdown
Existing CSV rows with the same `source_image` are replaced with fresh recognition results. Rows for images not selected in the current run are kept unless you use replace mode.
```

- [ ] **Step 3: Run documentation guard tests**

Run:

```bash
docker compose -f docker/test/docker-compose.yml run --rm go-scripts go test ./internal/projectfiles
```

Expected: PASS.

- [ ] **Step 4: Commit Task 3 if commits are requested**

```bash
git add README.md docs/QUICKSTART.md
git commit -m "docs: describe image refresh behavior"
```

Skip this step unless the user has explicitly requested commits in this session.

---

### Task 4: Final verification and closure

**Files:**
- No source changes expected beyond Tasks 1-3.

- [ ] **Step 1: Run full test suite inside Docker**

Run:

```bash
make test
```

Expected: PASS.

- [ ] **Step 2: Run strict quality gate**

Run:

```bash
python3 .opencode/meta/hooks/quality_gate.py --workspace . --mode quick --strict
```

Expected: JSON output with `"ok": true` and no `blocking_findings`.

- [ ] **Step 3: Inspect working tree**

Run:

```bash
git status --short
```

Expected changed files after implementation:

```text
 M README.md
 M docs/QUICKSTART.md
 M internal/app/app.go
 M internal/app/app_test.go
 M internal/catalog/csv.go
 M internal/catalog/csv_test.go
```

The two plan files may also be present if this planning artifact remains uncommitted:

```text
?? docs/plans/2026-05-17-reprocess-images-upsert-design.md
?? docs/superpowers/plans/2026-05-17-reprocess-images-upsert.md
```

- [ ] **Step 4: Final commit if commits are requested**

If the user explicitly asks to commit the full change as one commit, run:

```bash
git add README.md docs/QUICKSTART.md internal/app/app.go internal/app/app_test.go internal/catalog/csv.go internal/catalog/csv_test.go docs/plans/2026-05-17-reprocess-images-upsert-design.md docs/superpowers/plans/2026-05-17-reprocess-images-upsert.md
git commit -m "fix(app): reprocess images and upsert csv"
```

Do not commit unless explicitly requested.

---

## Self-Review

- Spec coverage: The plan covers menu option `1`, menu option `2`, CLI single-image runs, CLI all-image runs, `src → dst` overwrite, CSV update, and duplicate-row prevention through upsert.
- Placeholder scan: The plan contains exact file paths, concrete test code, concrete implementation code, and exact verification commands.
- Type consistency: The new helper is consistently named `catalog.Upsert(rows []Row, row Row) []Row`; `app.Process()` calls that function with a `catalog.Row` built from the recognizer output.
