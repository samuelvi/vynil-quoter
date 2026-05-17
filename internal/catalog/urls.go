package catalog

import (
	"net/url"
	"strings"
)

func ReferenceURLs(artist string, title string) ReferenceLinks {
	query := referenceQuery(artist, title)
	encoded := url.QueryEscape(query)
	return ReferenceLinks{
		Discogs: "https://www.discogs.com/search/?q=" + encoded + "&type=all",
		EBay:    "https://www.ebay.es/sch/i.html?_nkw=" + encoded,
		Popsike: "https://www.popsike.com/php/quicksearch.php?searchtext=" + encoded,
	}
}

func referenceQuery(artist string, title string) string {
	parts := make([]string, 0, 5)
	for _, value := range []string{artist, title} {
		cleaned := strings.TrimSpace(value)
		if cleaned == "" || strings.EqualFold(cleaned, "Unknown") {
			continue
		}
		parts = append(parts, cleaned)
	}
	parts = append(parts, "vinyl", "VG+", "sleeve", "VG+")
	return strings.Join(parts, " ")
}
