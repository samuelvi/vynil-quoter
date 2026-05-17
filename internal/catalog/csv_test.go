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
