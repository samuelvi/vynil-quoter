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
	if len(records) == 0 {
		return nil, nil
	}
	headerIndexes := headerIndex(records[0])
	rows := make([]Row, 0, len(records))
	for i, record := range records {
		if i == 0 {
			continue
		}
		rows = append(rows, Row{
			SourceImage:              field(record, headerIndexes, "source_image"),
			Artist:                   field(record, headerIndexes, "artist"),
			Title:                    field(record, headerIndexes, "title"),
			IdentificationConfidence: field(record, headerIndexes, "identification_confidence"),
			RecommendedPriceEUR:      field(record, headerIndexes, "recommended_price_eur"),
			Condition:                field(record, headerIndexes, "condition"),
			PriceConfidence:          field(record, headerIndexes, "price_confidence"),
			PriceBasis:               field(record, headerIndexes, "price_basis"),
			Notes:                    field(record, headerIndexes, "notes"),
			DiscogsReferenceURL:      field(record, headerIndexes, "discogs_reference_url"),
			EBayReferenceURL:         field(record, headerIndexes, "ebay_reference_url"),
			PopsikeReferenceURL:      field(record, headerIndexes, "popsike_reference_url"),
		})
	}
	return rows, nil
}

func headerIndex(header []string) map[string]int {
	indexes := make(map[string]int, len(header))
	for index, name := range header {
		indexes[name] = index
	}
	return indexes
}

func field(record []string, headerIndexes map[string]int, name string) string {
	index, ok := headerIndexes[name]
	if !ok || index >= len(record) {
		return ""
	}
	return record[index]
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
		if err := writer.Write([]string{row.SourceImage, row.Artist, row.Title, row.IdentificationConfidence, row.RecommendedPriceEUR, row.Condition, row.PriceConfidence, row.PriceBasis, row.Notes, row.DiscogsReferenceURL, row.EBayReferenceURL, row.PopsikeReferenceURL}); err != nil {
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
