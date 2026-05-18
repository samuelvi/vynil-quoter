package catalog_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"vinylquoter/internal/catalog"
)

func TestWriteAndReadRows(t *testing.T) {
	tmp := t.TempDir()
	report := filepath.Join(tmp, "data", "report", "album_catalog.csv")
	existing := catalog.Row{SourceImage: "data/src/a.jpg", Artist: "Artist A", Title: "Title A", IdentificationConfidence: "high", RecommendedPriceEUR: "12", PriceConfidence: "medium", PriceBasis: "existing"}
	if err := catalog.Write(report, []catalog.Row{existing}); err != nil {
		t.Fatal(err)
	}
	rows, err := catalog.Read(report)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].Artist != "Artist A" {
		t.Fatalf("got %#v", rows)
	}
}

func TestUpsertReplacesExistingRowByImageID(t *testing.T) {
	rows := []catalog.Row{{SourceImage: "data/src/DSC01.jpg", Artist: "Old", Title: "Old Title"}}
	fresh := catalog.Row{SourceImage: "DSC01.jpg", Artist: "Fresh", Title: "Fresh Title"}

	updated := catalog.Upsert(rows, fresh)

	if len(updated) != 1 {
		t.Fatalf("upsert should replace instead of append, got %#v", updated)
	}
	if updated[0].SourceImage != "DSC01.jpg" || updated[0].Artist != "Fresh" || updated[0].Title != "Fresh Title" {
		t.Fatalf("got %#v", updated[0])
	}
}

func TestUpsertAppendsMissingRowAndPreservesExistingOrder(t *testing.T) {
	rows := []catalog.Row{{SourceImage: "DSC01.jpg", Artist: "Artist 1"}}
	fresh := catalog.Row{SourceImage: "DSC02.jpg", Artist: "Artist 2"}

	updated := catalog.Upsert(rows, fresh)

	if len(updated) != 2 {
		t.Fatalf("got %#v", updated)
	}
	if updated[0].SourceImage != "DSC01.jpg" || updated[1].SourceImage != "DSC02.jpg" {
		t.Fatalf("upsert should preserve existing order and append new rows, got %#v", updated)
	}
}

func TestReferenceURLsUseBroadMarketplaceQueries(t *testing.T) {
	refs := catalog.ReferenceURLs("The Cure", "Disintegration")

	for name, got := range map[string]string{
		"discogs": refs.Discogs,
		"ebay":    refs.EBay,
		"popsike": refs.Popsike,
	} {
		if strings.Contains(got, "VG") || strings.Contains(got, "sleeve") {
			t.Fatalf("%s URL should not contain condition hints: %s", name, got)
		}
	}
	if !strings.Contains(refs.Discogs, "q=The+Cure+Disintegration") || !strings.Contains(refs.Discogs, "type=release") {
		t.Fatalf("Discogs URL missing broad release query: %s", refs.Discogs)
	}
	if !strings.Contains(refs.EBay, "_nkw=The+Cure+Disintegration+vinyl+lp") {
		t.Fatalf("eBay URL missing broad marketplace query: %s", refs.EBay)
	}
	if !strings.Contains(refs.Popsike, "searchtext=The+Cure+Disintegration") {
		t.Fatalf("Popsike URL missing broad artist/title query: %s", refs.Popsike)
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

func TestReferenceURLsSkipUnknownAndEmptyValues(t *testing.T) {
	refs := catalog.ReferenceURLs("Unknown", "")

	for name, got := range map[string]string{
		"discogs": refs.Discogs,
		"ebay":    refs.EBay,
		"popsike": refs.Popsike,
	} {
		if strings.Contains(got, "Unknown") || strings.Contains(got, "VG") || strings.Contains(got, "sleeve") {
			t.Fatalf("%s URL should not contain unknown or condition hints: %s", name, got)
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

func TestWriteReadRowsIncludesConditionColumn(t *testing.T) {
	tmp := t.TempDir()
	report := filepath.Join(tmp, "data", "report", "album_catalog.csv")
	row := catalog.Row{SourceImage: "DSC01.jpg", Artist: "The Cure", Title: "Disintegration", RecommendedPriceEUR: "22", Condition: "media: VG; sleeve: G+", PriceConfidence: "medium"}

	if err := catalog.Write(report, []catalog.Row{row}); err != nil {
		t.Fatal(err)
	}
	content, err := os.ReadFile(report)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(content), "recommended_price_eur,condition,price_confidence") {
		t.Fatalf("CSV header missing condition between price and confidence: %s", string(content))
	}
	rows, err := catalog.Read(report)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].Condition != row.Condition {
		t.Fatalf("got %#v", rows)
	}
}

func TestReadOldElevenColumnCSVKeepsEmptyCondition(t *testing.T) {
	tmp := t.TempDir()
	report := filepath.Join(tmp, "album_catalog.csv")
	oldCSV := "source_image,artist,title,identification_confidence,recommended_price_eur,price_confidence,price_basis,notes,discogs_reference_url,ebay_reference_url,popsike_reference_url\nDSC01.jpg,The Cure,Disintegration,high,22,medium,basis,notes,https://discogs.example,https://ebay.example,https://popsike.example\n"
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
	if rows[0].Condition != "" {
		t.Fatalf("old CSV rows should have empty condition, got %#v", rows[0])
	}
	if rows[0].DiscogsReferenceURL != "https://discogs.example" || rows[0].EBayReferenceURL != "https://ebay.example" || rows[0].PopsikeReferenceURL != "https://popsike.example" {
		t.Fatalf("old CSV URLs should stay mapped, got %#v", rows[0])
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
	if rows[0].Condition != "" {
		t.Fatalf("old CSV rows should have empty condition, got %#v", rows[0])
	}
}
