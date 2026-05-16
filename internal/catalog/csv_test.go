package catalog

import (
	"path/filepath"
	"testing"
)

func TestWriteReadAndPendingUpdate(t *testing.T) {
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
	pending, err := Pending([]string{"data/src/a.jpg", "data/src/b.jpg"}, report, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(pending) != 1 || pending[0] != "data/src/b.jpg" {
		t.Fatalf("got %#v", pending)
	}
}

func TestPendingReplaceProcessesEveryImage(t *testing.T) {
	pending, err := Pending([]string{"data/src/a.jpg", "data/src/b.jpg"}, "missing.csv", true)
	if err != nil {
		t.Fatal(err)
	}
	if len(pending) != 2 {
		t.Fatalf("got %#v", pending)
	}
}
