# Condition Menu Report Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add selectable media/sleeve conditions to the menu, prompt, report CSV, and reference URL generation while keeping price cells numeric-only.

**Architecture:** Store media and sleeve condition in `config.RunConfig`, expose condition selection through `internal/ui`, pass the selected condition into `app.Process`, and persist a single formatted `condition` CSV column. Keep provider JSON contract compatible by changing prompt text and sanitize the returned price before row creation.

**Tech Stack:** Go 1.23, standard library only, Docker-backed `make test`, CSV persistence in `internal/catalog`.

---

### Task 1: Add condition defaults and validation

**Files:**
- Modify: `internal/config/config.go`
- Test: `tests/internal/app/app_test.go`

**Step 1: Write the failing test**

Add a test near `TestParseArgsDefaults`:

```go
func TestParseArgsDefaultsToVGConditions(t *testing.T) {
	cfg, err := app.ParseArgs([]string{"--all"})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.MediaCondition != config.DefaultCondition || cfg.SleeveCondition != config.DefaultCondition {
		t.Fatalf("got %#v", cfg)
	}
}

func TestParseArgsSupportsConditionFlags(t *testing.T) {
	cfg, err := app.ParseArgs([]string{"--all", "--media-condition", "VG+", "--sleeve-condition", "G+"})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.MediaCondition != "VG+" || cfg.SleeveCondition != "G+" {
		t.Fatalf("got %#v", cfg)
	}
}

func TestParseArgsRejectsInvalidCondition(t *testing.T) {
	_, err := app.ParseArgs([]string{"--all", "--media-condition", "BAD"})
	if err == nil || !strings.Contains(err.Error(), "invalid media condition") {
		t.Fatalf("expected invalid condition error, got %v", err)
	}
}
```

**Step 2: Run test to verify it fails**

Run:

```bash
docker compose -f docker/test/docker-compose.yml run --rm go-scripts go test ./tests/internal/app -run 'TestParseArgs.*Condition' -count=1 -v
```

Expected: FAIL because `RunConfig.MediaCondition`, `RunConfig.SleeveCondition`, and condition flags do not exist.

**Step 3: Implement minimal config support**

In `internal/config/config.go`, add:

```go
const (
	ConditionMint      = "M"
	ConditionNearMint  = "NM/M-"
	ConditionVeryGoodPlus = "VG+"
	DefaultCondition   = "VG"
	ConditionGoodPlus  = "G+"
	ConditionGood      = "G"
	ConditionFair      = "F"
	ConditionPoor      = "P"
	ConditionGeneric   = "Generic"
)

var MediaConditions = []string{ConditionMint, ConditionNearMint, ConditionVeryGoodPlus, DefaultCondition, ConditionGoodPlus, ConditionGood, ConditionFair, ConditionPoor}
var SleeveConditions = []string{ConditionMint, ConditionNearMint, ConditionVeryGoodPlus, DefaultCondition, ConditionGoodPlus, ConditionGood, ConditionFair, ConditionPoor, ConditionGeneric}

func IsMediaCondition(value string) bool { return containsCondition(MediaConditions, value) }
func IsSleeveCondition(value string) bool { return containsCondition(SleeveConditions, value) }

func containsCondition(values []string, value string) bool {
	for _, candidate := range values {
		if candidate == value {
			return true
		}
	}
	return false
}
```

Add fields to `RunConfig`:

```go
MediaCondition  string
SleeveCondition string
```

Add defaults in `DefaultRunConfig()`:

```go
MediaCondition:  DefaultCondition,
SleeveCondition: DefaultCondition,
```

In `internal/app/app.go`, add flags and validation:

```go
set.StringVar(&cfg.MediaCondition, "media-condition", cfg.MediaCondition, "vinyl media condition grade")
set.StringVar(&cfg.SleeveCondition, "sleeve-condition", cfg.SleeveCondition, "sleeve/cover condition grade")
```

After parsing:

```go
if !config.IsMediaCondition(cfg.MediaCondition) {
	return cfg, fmt.Errorf("invalid media condition: %s", cfg.MediaCondition)
}
if !config.IsSleeveCondition(cfg.SleeveCondition) {
	return cfg, fmt.Errorf("invalid sleeve condition: %s", cfg.SleeveCondition)
}
```

**Step 4: Run test to verify it passes**

Run the same command. Expected: PASS.

**Step 5: Commit**

```bash
git add internal/config/config.go internal/app/app.go tests/internal/app/app_test.go
git commit -m "feat(config): add record condition options"
```

---

### Task 2: Add menu display and condition selection

**Files:**
- Modify: `internal/ui/menu.go`
- Modify: `internal/app/app.go`
- Test: `tests/internal/ui/menu_test.go`
- Test: `tests/internal/app/app_test.go`

**Step 1: Write the failing tests**

Add UI tests:

```go
func TestMenuShowsCurrentConditionState(t *testing.T) {
	stdout := &bytes.Buffer{}
	state := config.DefaultRunConfig()
	state.MediaCondition = "VG+"
	state.SleeveCondition = "G+"
	_, _ = ui.ReadMenuWithState(bytes.NewBufferString("7\n"), stdout, state)
	output := stdout.String()
	if !strings.Contains(output, "Calidad carátula (G+)") {
		t.Fatalf("menu should show sleeve condition, got %s", output)
	}
	if !strings.Contains(output, "Calidad vinilo (VG+)") {
		t.Fatalf("menu should show media condition, got %s", output)
	}
}

func TestMenuCanSelectSleeveCondition(t *testing.T) {
	cfg, err := ui.ReadMenuWithState(bytes.NewBufferString("5\n3\n"), &bytes.Buffer{}, config.DefaultRunConfig())
	if err != ui.ErrNoAction {
		t.Fatalf("got %v", err)
	}
	if cfg.SleeveCondition != "VG+" {
		t.Fatalf("got %#v", cfg)
	}
}

func TestMenuCanSelectMediaCondition(t *testing.T) {
	cfg, err := ui.ReadMenuWithState(bytes.NewBufferString("6\n5\n"), &bytes.Buffer{}, config.DefaultRunConfig())
	if err != ui.ErrNoAction {
		t.Fatalf("got %v", err)
	}
	if cfg.MediaCondition != "G+" {
		t.Fatalf("got %#v", cfg)
	}
}
```

Update existing tests that use option `5` to exit so they use `7` instead.

Add app persistence test adjustment to `TestRunInteractivePersistsSelectedModelAndCSV` or create a new test:

```go
func TestRunInteractivePersistsSelectedConditions(t *testing.T) {
	// create one tiny image, input: choose sleeve VG+ (5,3), media G+ (6,5), process image, exit.
	// assert factory receives cfg.SleeveCondition == "VG+" and cfg.MediaCondition == "G+".
}
```

**Step 2: Run tests to verify they fail**

Run:

```bash
docker compose -f docker/test/docker-compose.yml run --rm go-scripts go test ./tests/internal/ui ./tests/internal/app -run 'Condition|Interactive' -count=1 -v
```

Expected: FAIL because menu options do not exist and exit option is still `5`.

**Step 3: Implement menu changes**

In `internal/ui/menu.go`, change main menu lines:

```go
fmt.Fprintf(out, "5) Calidad carátula (%s)\n", cfg.SleeveCondition)
fmt.Fprintf(out, "6) Calidad vinilo (%s)\n", cfg.MediaCondition)
fmt.Fprintln(out, "7) Salir")
fmt.Fprint(out, "Elige una opción [1-7]: ")
```

Add switch cases:

```go
case "5":
	condition, err := ReadSleeveCondition(reader, out, cfg.SleeveCondition)
	if err != nil { return cfg, err }
	cfg.SleeveCondition = condition
	return cfg, ErrNoAction
case "6":
	condition, err := ReadMediaCondition(reader, out, cfg.MediaCondition)
	if err != nil { return cfg, err }
	cfg.MediaCondition = condition
	return cfg, ErrNoAction
case "7":
	return cfg, io.EOF
```

Add helpers:

```go
func ReadMediaCondition(in *bufio.Reader, out io.Writer, current string) (string, error) {
	return readCondition(in, out, "Calidad vinilo", current, config.MediaConditions)
}

func ReadSleeveCondition(in *bufio.Reader, out io.Writer, current string) (string, error) {
	return readCondition(in, out, "Calidad carátula", current, config.SleeveConditions)
}

func readCondition(in *bufio.Reader, out io.Writer, title string, current string, values []string) (string, error) {
	fmt.Fprintf(out, "\n%s (%s)\n", title, current)
	for index, value := range values {
		defaultLabel := ""
		if value == config.DefaultCondition {
			defaultLabel = " [por defecto]"
		}
		fmt.Fprintf(out, "%d) %s%s\n", index+1, value, defaultLabel)
	}
	fmt.Fprintf(out, "Elige calidad [1-%d, Enter=%s]: ", len(values), current)
	choice, _ := in.ReadString('\n')
	trimmed := strings.TrimSpace(choice)
	if trimmed == "" {
		return current, nil
	}
	index, err := strconv.Atoi(trimmed)
	if err != nil || index < 1 || index > len(values) {
		return current, fmt.Errorf("invalid condition choice")
	}
	return values[index-1], nil
}
```

Import `strconv`.

In `internal/app/app.go`, persist fields after menu selection:

```go
menuCfg.MediaCondition = state.MediaCondition
menuCfg.SleeveCondition = state.SleeveCondition
```

**Step 4: Run tests to verify they pass**

Run the same command. Expected: PASS.

**Step 5: Commit**

```bash
git add internal/ui/menu.go internal/app/app.go tests/internal/ui/menu_test.go tests/internal/app/app_test.go
git commit -m "feat(ui): select record conditions"
```

---

### Task 3: Add CSV condition column and backward compatibility

**Files:**
- Modify: `internal/catalog/row.go`
- Modify: `internal/catalog/csv.go`
- Test: `tests/internal/catalog/csv_test.go`

**Step 1: Write failing tests**

Add/adjust tests:

```go
func TestWriteReadRowsIncludesConditionColumn(t *testing.T) {
	tmp := t.TempDir()
	report := filepath.Join(tmp, "album_catalog.csv")
	row := catalog.Row{SourceImage: "DSC01.jpg", Artist: "The Cure", Title: "Disintegration", Condition: "media: VG; sleeve: G+"}

	if err := catalog.Write(report, []catalog.Row{row}); err != nil {
		t.Fatal(err)
	}
	content, err := os.ReadFile(report)
	if err != nil { t.Fatal(err) }
	if !strings.Contains(string(content), "recommended_price_eur,condition,price_confidence") {
		t.Fatalf("CSV header missing condition after price: %s", string(content))
	}
	rows, err := catalog.Read(report)
	if err != nil { t.Fatal(err) }
	if len(rows) != 1 || rows[0].Condition != row.Condition {
		t.Fatalf("got %#v", rows)
	}
}

func TestReadOldElevenColumnCSVKeepsEmptyCondition(t *testing.T) {
	// old header without condition but with reference URL columns
	// assert rows[0].Condition == "" and URLs remain mapped correctly.
}
```

**Step 2: Run tests to verify they fail**

Run:

```bash
docker compose -f docker/test/docker-compose.yml run --rm go-scripts go test ./tests/internal/catalog -run 'Condition|Old' -count=1 -v
```

Expected: FAIL because `Row.Condition` and CSV mapping do not exist.

**Step 3: Implement CSV mapping**

In `internal/catalog/row.go`, add field:

```go
Condition string
```

Update header:

```go
var Header = []string{"source_image", "artist", "title", "identification_confidence", "recommended_price_eur", "condition", "price_confidence", "price_basis", "notes", "discogs_reference_url", "ebay_reference_url", "popsike_reference_url"}
```

In `internal/catalog/csv.go`, avoid positional breakage by mapping header names:

```go
func rowFromRecord(header []string, record []string) Row { ... }
func valueFor(header []string, record []string, name string) string { ... }
```

Write rows in `Header` order with the new `Condition` field after price.

**Step 4: Run tests to verify they pass**

Run same command. Expected: PASS.

**Step 5: Commit**

```bash
git add internal/catalog/row.go internal/catalog/csv.go tests/internal/catalog/csv_test.go
git commit -m "feat(csv): report selected condition"
```

---

### Task 4: Pass condition into processing and sanitize prices

**Files:**
- Modify: `internal/app/app.go`
- Test: `tests/internal/app/app_test.go`

**Step 1: Write failing tests**

Add tests:

```go
func TestProcessWritesConditionAndSanitizesPrice(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "data", "src", "DSC01.jpg")
	writeTinyJPEG(t, src)
	report := filepath.Join(tmp, "data", "report", "album_catalog.csv")
	dstDir := filepath.Join(tmp, "data", "dst")
	cfg := config.DefaultRunConfig()
	cfg.MediaCondition = "VG+"
	cfg.SleeveCondition = "G+"

	rows, err := app.ProcessWithConfig(context.Background(), []string{src}, report, false, dstDir, cfg, recognizerFunc(func(ctx context.Context, imagePath string) (catalog.Identification, error) {
		return catalog.Identification{Artist: "The Cure", Title: "Disintegration", RecommendedPriceEUR: "15-20 EUR"}, nil
	}))
	if err != nil { t.Fatal(err) }
	if rows[0].Condition != "media: VG+; sleeve: G+" {
		t.Fatalf("got %#v", rows[0])
	}
	if rows[0].RecommendedPriceEUR != "15-20" {
		t.Fatalf("got price %q", rows[0].RecommendedPriceEUR)
	}
}
```

**Step 2: Run test to verify it fails**

Run:

```bash
docker compose -f docker/test/docker-compose.yml run --rm go-scripts go test ./tests/internal/app -run 'Condition|Sanitize' -count=1 -v
```

Expected: FAIL because `ProcessWithConfig`, condition formatting, and price sanitization do not exist.

**Step 3: Implement processing changes**

Keep existing `Process` for compatibility:

```go
func Process(ctx context.Context, images []string, reportPath string, replace bool, destinationDir string, recognizer provider.Recognizer) ([]catalog.Row, error) {
	return ProcessWithConfig(ctx, images, reportPath, replace, destinationDir, config.DefaultRunConfig(), recognizer)
}
```

Create:

```go
func ProcessWithConfig(ctx context.Context, images []string, reportPath string, replace bool, destinationDir string, cfg config.RunConfig, recognizer provider.Recognizer) ([]catalog.Row, error) {
	// moved existing body here
}
```

Update `runOnce` to call `ProcessWithConfig(..., cfg, recognizer)`.

Add helpers:

```go
func conditionLabel(cfg config.RunConfig) string {
	return "media: " + cfg.MediaCondition + "; sleeve: " + cfg.SleeveCondition
}

func numericPrice(value string) string {
	// Trim spaces, remove €, EUR, eur, euros, commas; keep digits, dots, and hyphen ranges.
}
```

Set row fields:

```go
RecommendedPriceEUR: numericPrice(identification.RecommendedPriceEUR),
Condition: conditionLabel(cfg),
```

**Step 4: Run test to verify it passes**

Run same command. Expected: PASS.

**Step 5: Commit**

```bash
git add internal/app/app.go tests/internal/app/app_test.go
git commit -m "fix(report): normalize condition and price"
```

---

### Task 5: Improve reference URLs

**Files:**
- Modify: `internal/catalog/urls.go`
- Test: `tests/internal/catalog/csv_test.go`

**Step 1: Write failing test**

Update `TestReferenceURLsUseArtistTitleAndConditionHint` to expect no condition hints:

```go
func TestReferenceURLsUseBroadMarketplaceQueries(t *testing.T) {
	refs := catalog.ReferenceURLs("The Cure", "Disintegration")

	if !strings.Contains(refs.Discogs, "q=The+Cure+Disintegration") || !strings.Contains(refs.Discogs, "type=release") {
		t.Fatalf("unexpected Discogs URL: %s", refs.Discogs)
	}
	if !strings.Contains(refs.EBay, "_nkw=The+Cure+Disintegration+vinyl+lp") {
		t.Fatalf("unexpected eBay URL: %s", refs.EBay)
	}
	if !strings.Contains(refs.Popsike, "searchtext=The+Cure+Disintegration") {
		t.Fatalf("unexpected Popsike URL: %s", refs.Popsike)
	}
	for _, got := range []string{refs.Discogs, refs.EBay, refs.Popsike} {
		if strings.Contains(got, "VG") || strings.Contains(got, "sleeve") {
			t.Fatalf("URL should not include condition hints: %s", got)
		}
	}
}
```

**Step 2: Run test to verify it fails**

Run:

```bash
docker compose -f docker/test/docker-compose.yml run --rm go-scripts go test ./tests/internal/catalog -run ReferenceURLs -count=1 -v
```

Expected: FAIL because URLs still include `VG+ sleeve VG+`.

**Step 3: Implement cleaner URL queries**

In `internal/catalog/urls.go`:

```go
func ReferenceURLs(artist string, title string) ReferenceLinks {
	baseQuery := referenceQuery(artist, title)
	ebayQuery := strings.TrimSpace(baseQuery + " vinyl lp")
	return ReferenceLinks{
		Discogs: "https://www.discogs.com/search/?q=" + url.QueryEscape(baseQuery) + "&type=release",
		EBay:    "https://www.ebay.es/sch/i.html?_nkw=" + url.QueryEscape(ebayQuery),
		Popsike: "https://www.popsike.com/php/quicksearch.php?searchtext=" + url.QueryEscape(baseQuery),
	}
}

func referenceQuery(artist string, title string) string {
	// only artist/title; no vinyl/condition/sleeve hints here
}
```

**Step 4: Run test to verify it passes**

Run same command. Expected: PASS.

**Step 5: Commit**

```bash
git add internal/catalog/urls.go tests/internal/catalog/csv_test.go
git commit -m "fix(catalog): broaden reference searches"
```

---

### Task 6: Update prompt for dynamic condition and numeric-only price

**Files:**
- Modify: `internal/provider/visionpayload/visionpayload.go`
- Modify: `internal/provider/lmstudio/lmstudio.go`
- Modify: `internal/provider/gemini/gemini.go`
- Modify: `internal/provider/provider.go`
- Test: provider tests under `tests/internal/provider/**`

**Step 1: Write failing tests**

In `tests/internal/provider/visionpayload/visionpayload_test.go`, add:

```go
func TestPromptIncludesSelectedConditionAndNumericPriceContract(t *testing.T) {
	prompt := visionpayload.PromptForCondition("VG+", "G+")
	for _, want := range []string{"media VG+", "sleeve G+", "numbers only", "without currency", "12-18"} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("prompt missing %q: %s", want, prompt)
		}
	}
}
```

Provider client tests should assert request prompts include the selected conditions after interface changes.

**Step 2: Run tests to verify they fail**

Run:

```bash
docker compose -f docker/test/docker-compose.yml run --rm go-scripts go test ./tests/internal/provider/... -run Prompt -count=1 -v
```

Expected: FAIL because prompt is static.

**Step 3: Implement prompt changes**

In `internal/provider/provider.go`, replace the interface with:

```go
type RecognitionRequest struct {
	ImagePath       string
	MediaCondition string
	SleeveCondition string
}

type Recognizer interface {
	Identify(ctx context.Context, request RecognitionRequest) (catalog.Identification, error)
}
```

Update app calls:

```go
recognizer.Identify(ctx, provider.RecognitionRequest{ImagePath: imageForRecognition, MediaCondition: cfg.MediaCondition, SleeveCondition: cfg.SleeveCondition})
```

In `visionpayload`, keep `Prompt()` as default wrapper and add:

```go
func Prompt() string { return PromptForCondition("VG", "VG") }

func PromptForCondition(mediaCondition string, sleeveCondition string) string {
	return "Identify ..." +
		`"recommended_price_eur":"number-or-range",` +
		"Price assumptions: Spain/EU market, EUR currency, media " + mediaCondition + ", sleeve " + sleeveCondition + ", normal second-hand sale. " +
		"Return recommended_price_eur as numbers only without currency symbols or text, for example 12 or 12-18. " +
		...
}
```

Update LM Studio and Gemini clients to use `request.ImagePath` and `visionpayload.PromptForCondition(request.MediaCondition, request.SleeveCondition)`.

Update fake recognizers in tests to accept `provider.RecognitionRequest`.

**Step 4: Run tests to verify they pass**

Run:

```bash
docker compose -f docker/test/docker-compose.yml run --rm go-scripts go test ./tests/internal/provider/... ./tests/internal/app -count=1 -v
```

Expected: PASS.

**Step 5: Commit**

```bash
git add internal/provider internal/app tests/internal/provider tests/internal/app
git commit -m "feat(provider): price by selected condition"
```

---

### Task 7: Full verification and quality gate

**Files:**
- All changed files

**Step 1: Format Go files**

Run:

```bash
docker compose -f docker/test/docker-compose.yml run --rm go-scripts gofmt -w internal tests
```

Expected: no output or formatting changes only.

**Step 2: Run full tests**

Run:

```bash
make test
```

Expected: all packages PASS.

**Step 3: Run diff check**

Run:

```bash
git diff --check
```

Expected: no output.

**Step 4: Run quality gate**

Run:

```bash
python3 /Users/work/Projects/Personal/VinylQuoter/opencode-bundle/.opencode/meta/hooks/quality_gate.py --workspace . --mode quick --strict
```

Expected: JSON with `"ok": true` and no blocking findings.

**Step 5: Commit final docs if needed**

If any documentation/test fixture changes remain uncommitted:

```bash
git status --short
git add <intended-files>
git commit -m "docs: update condition workflow"
```

**Step 6: Report handoff**

Report:

- Branch name and worktree path.
- Commit list.
- Verification command outputs.
- Remaining risks, especially provider model compliance with numeric-only prices.
