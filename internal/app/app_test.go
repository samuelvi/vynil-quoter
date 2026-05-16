package app

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"vinylquoter/internal/catalog"
	"vinylquoter/internal/config"
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
