package jsonparse

import (
	"encoding/json"
	"errors"
	"strings"
	"vinylquoter/internal/catalog"
)

func IdentificationFromText(text string) (catalog.Identification, error) {
	cleaned := strings.TrimSpace(text)
	if strings.HasPrefix(cleaned, "```") {
		cleaned = strings.TrimPrefix(cleaned, "```json")
		cleaned = strings.TrimPrefix(cleaned, "```")
		cleaned = strings.TrimSuffix(cleaned, "```")
		cleaned = strings.TrimSpace(cleaned)
	}
	start := strings.Index(cleaned, "{")
	end := strings.LastIndex(cleaned, "}")
	if start < 0 || end < start {
		return catalog.Identification{}, errors.New("response did not contain JSON object")
	}
	var result catalog.Identification
	if err := json.Unmarshal([]byte(cleaned[start:end+1]), &result); err != nil {
		return catalog.Identification{}, err
	}
	if strings.TrimSpace(result.Artist) == "" {
		result.Artist = "Unknown"
	}
	if strings.TrimSpace(result.Title) == "" {
		result.Title = "Unknown"
	}
	if strings.TrimSpace(result.IdentificationConfidence) == "" {
		result.IdentificationConfidence = "manual-review"
	}
	if strings.TrimSpace(result.PriceConfidence) == "" {
		result.PriceConfidence = "manual-review"
	}
	return result, nil
}
