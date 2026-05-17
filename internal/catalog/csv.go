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
		rows = append(rows, Row{record[0], record[1], record[2], record[3], record[4], record[5], record[6], record[7], record[8], record[9], record[10]})
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
		if err := writer.Write([]string{row.SourceImage, row.Artist, row.Title, row.IdentificationConfidence, row.RecommendedPriceEUR, row.PriceConfidence, row.PriceBasis, row.Notes, row.DiscogsReferenceURL, row.EBayReferenceURL, row.PopsikeReferenceURL}); err != nil {
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
