package app

import (
	"context"
	"path/filepath"
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
