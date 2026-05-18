package catalog

import (
	"net/url"
	"strings"
)

func ReferenceURLs(artist string, title string) ReferenceLinks {
	query := referenceQuery(artist, title)
	encoded := url.QueryEscape(query)
	ebayQuery := strings.TrimSpace(strings.Join([]string{query, "vinyl lp"}, " "))
	return ReferenceLinks{
		Discogs: "https://www.discogs.com/search/?q=" + encoded + "&type=release",
		EBay:    "https://www.ebay.es/sch/i.html?_nkw=" + url.QueryEscape(ebayQuery),
		Popsike: "https://www.popsike.com/php/quicksearch.php?searchtext=" + encoded,
	}
}

func referenceQuery(artist string, title string) string {
	parts := make([]string, 0, 2)
	for _, value := range []string{artist, title} {
		cleaned := strings.TrimSpace(value)
		if cleaned == "" || strings.EqualFold(cleaned, "Unknown") {
			continue
		}
		parts = append(parts, cleaned)
	}
	return strings.Join(parts, " ")
}
