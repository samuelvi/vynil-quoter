package catalog

import (
	"encoding/csv"
	"errors"
	"os"
	"path/filepath"
)

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

func Pending(images []string, reportPath string, replace bool) ([]string, error) {
	if replace {
		return images, nil
	}
	rows, err := Read(reportPath)
	if err != nil {
		return nil, err
	}
	existing := map[string]struct{}{}
	for _, row := range rows {
		existing[row.SourceImage] = struct{}{}
		existing[filepath.Base(row.SourceImage)] = struct{}{}
	}
	pending := make([]string, 0, len(images))
	for _, image := range images {
		if _, ok := existing[image]; ok {
			continue
		}
		if _, ok := existing[filepath.Base(image)]; ok {
			continue
		}
		pending = append(pending, image)
	}
	return pending, nil
}
