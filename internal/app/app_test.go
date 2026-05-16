package app

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"vinylquoter/internal/catalog"
	"vinylquoter/internal/config"
	"vinylquoter/internal/provider"
)

type fakeRecognizer struct{}

func (fakeRecognizer) Identify(ctx context.Context, imagePath string) (catalog.Identification, error) {
	return catalog.Identification{Artist: "The Cure", Title: "Disintegration", IdentificationConfidence: "high", RecommendedPriceEUR: "22", PriceConfidence: "medium", PriceBasis: "test", Notes: "ok"}, nil
}

func TestParseArgsDefaults(t *testing.T) {
	cfg, err := ParseArgs([]string{"--all"})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.SourceDir != config.DefaultSourceDir || cfg.ReportPath != config.DefaultReportPath || cfg.Provider != config.DefaultProvider {
		t.Fatalf("got %#v", cfg)
	}
}

func TestProcessWritesRows(t *testing.T) {
	tmp := t.TempDir()
	report := filepath.Join(tmp, "data", "report", "album_catalog.csv")
	rows, err := Process(context.Background(), []string{"data/src/a.jpg"}, report, false, fakeRecognizer{})
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].Artist != "The Cure" {
		t.Fatalf("got %#v", rows)
	}
}

func TestProcessUsesImageBasenameAsCSVIdentifier(t *testing.T) {
	tmp := t.TempDir()
	report := filepath.Join(tmp, "data", "report", "album_catalog.csv")
	rows, err := Process(context.Background(), []string{"data/src/DSC01.jpg"}, report, false, fakeRecognizer{})
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].SourceImage != "DSC01.jpg" {
		t.Fatalf("got %#v", rows)
	}
	content, err := os.ReadFile(report)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(content), "data/src/DSC01.jpg") {
		t.Fatalf("CSV should use basename identifier, got %s", string(content))
	}
}

func TestProcessSkipsExistingBasenameIdentifier(t *testing.T) {
	tmp := t.TempDir()
	report := filepath.Join(tmp, "data", "report", "album_catalog.csv")
	if err := catalog.Write(report, []catalog.Row{{SourceImage: "DSC01.jpg", Artist: "Existing"}}); err != nil {
		t.Fatal(err)
	}
	rows, err := Process(context.Background(), []string{"data/src/DSC01.jpg"}, report, false, fakeRecognizer{})
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].Artist != "Existing" {
		t.Fatalf("got %#v", rows)
	}
}

func TestParseArgsSingleImageNeverReplacesCSV(t *testing.T) {
	cfg, err := ParseArgs([]string{"--image", "DSC01.jpg", "--replace"})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Replace {
		t.Fatalf("single image CLI mode must never replace CSV: %#v", cfg)
	}
}

func TestParseArgsSupportsAlternateLMStudioVisionModel(t *testing.T) {
	cfg, err := ParseArgs([]string{"--all", "--model", config.AlternateLMStudioModel})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Provider != config.ProviderLMStudio || cfg.Model != config.AlternateLMStudioModel {
		t.Fatalf("got %#v", cfg)
	}
}

func TestRunInteractiveReturnsToMenuUntilExit(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "data", "src")
	if err := os.MkdirAll(src, 0o755); err != nil {
		t.Fatal(err)
	}
	image := filepath.Join(src, "DSC01.jpg")
	if err := os.WriteFile(image, []byte("jpg"), 0o644); err != nil {
		t.Fatal(err)
	}
	report := filepath.Join(tmp, "data", "report", "album_catalog.csv")
	stdin := bytes.NewBufferString("1\n" + image + "\n5\n")
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cfg := config.DefaultRunConfig()
	cfg.SourceDir = src
	cfg.ReportPath = report
	code := runWithRecognizerFactory(context.Background(), cfg, stdin, stdout, stderr, func(config.RunConfig) (provider.Recognizer, error) {
		return fakeRecognizer{}, nil
	})
	if code != 0 {
		t.Fatalf("code=%d stderr=%s", code, stderr.String())
	}
	if strings.Count(stdout.String(), "Vinyl Quoter") < 2 {
		t.Fatalf("menu should be shown again before exit, got %s", stdout.String())
	}
}

func TestRunInteractivePersistsSelectedModelAndCSV(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "data", "src")
	if err := os.MkdirAll(src, 0o755); err != nil {
		t.Fatal(err)
	}
	image := filepath.Join(src, "DSC01.jpg")
	if err := os.WriteFile(image, []byte("jpg"), 0o644); err != nil {
		t.Fatal(err)
	}
	report := filepath.Join(tmp, "custom.csv")
	stdin := bytes.NewBufferString("4\n3\n3\n1\n" + report + "\n1\n" + image + "\n5\n")
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cfg := config.DefaultRunConfig()
	cfg.SourceDir = src
	var seen []config.RunConfig
	code := runWithRecognizerFactory(context.Background(), cfg, stdin, stdout, stderr, func(cfg config.RunConfig) (provider.Recognizer, error) {
		seen = append(seen, cfg)
		return fakeRecognizer{}, nil
	})
	if code != 0 {
		t.Fatalf("code=%d stderr=%s", code, stderr.String())
	}
	if len(seen) != 1 {
		t.Fatalf("got configs %#v", seen)
	}
	if seen[0].ReportPath != report || seen[0].Provider != config.ProviderGemini || seen[0].Model != config.DefaultGeminiModel {
		t.Fatalf("state was not persisted: %#v", seen[0])
	}
}

func TestRunInteractiveOptionTwoProcessesAllImages(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "data", "src")
	if err := os.MkdirAll(src, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"DSC01.jpg", "DSC02.jpg"} {
		if err := os.WriteFile(filepath.Join(src, name), []byte("jpg"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	report := filepath.Join(tmp, "data", "report", "album_catalog.csv")
	stdin := bytes.NewBufferString("2\n5\n")
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cfg := config.DefaultRunConfig()
	cfg.SourceDir = src
	cfg.ReportPath = report
	var seen []string
	code := runWithRecognizerFactory(context.Background(), cfg, stdin, stdout, stderr, func(config.RunConfig) (provider.Recognizer, error) {
		return recognizerFunc(func(ctx context.Context, imagePath string) (catalog.Identification, error) {
			seen = append(seen, filepath.Base(imagePath))
			return catalog.Identification{Artist: "A", Title: "T"}, nil
		}), nil
	})
	if code != 0 {
		t.Fatalf("code=%d stderr=%s", code, stderr.String())
	}
	if strings.Join(seen, ",") != "DSC01.jpg,DSC02.jpg" {
		t.Fatalf("option 2 should process all images, got %#v", seen)
	}
}

func TestRunInteractiveResetsActionModeAfterOptionTwo(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "data", "src")
	if err := os.MkdirAll(src, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"DSC01.jpg", "DSC02.jpg"} {
		if err := os.WriteFile(filepath.Join(src, name), []byte("jpg"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	report := filepath.Join(tmp, "data", "report", "album_catalog.csv")
	stdin := bytes.NewBufferString("2\n1\nDSC01.jpg\n5\n")
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cfg := config.DefaultRunConfig()
	cfg.SourceDir = src
	cfg.ReportPath = report
	var seen []string
	code := runWithRecognizerFactory(context.Background(), cfg, stdin, stdout, stderr, func(config.RunConfig) (provider.Recognizer, error) {
		return recognizerFunc(func(ctx context.Context, imagePath string) (catalog.Identification, error) {
			seen = append(seen, filepath.Base(imagePath))
			return catalog.Identification{Artist: "A", Title: "T"}, nil
		}), nil
	})
	if code != 0 {
		t.Fatalf("code=%d stderr=%s", code, stderr.String())
	}
	if stderr.String() != "" {
		t.Fatalf("unexpected stderr: %s", stderr.String())
	}
	if strings.Join(seen, ",") != "DSC01.jpg,DSC02.jpg" {
		t.Fatalf("option 1 should not fail after option 2, got %#v", seen)
	}
}

type recognizerFunc func(context.Context, string) (catalog.Identification, error)

func (fn recognizerFunc) Identify(ctx context.Context, imagePath string) (catalog.Identification, error) {
	return fn(ctx, imagePath)
}
