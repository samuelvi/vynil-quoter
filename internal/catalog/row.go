package catalog

type Identification struct {
	Artist                   string `json:"artist"`
	Title                    string `json:"title"`
	IdentificationConfidence string `json:"identification_confidence"`
	RecommendedPriceEUR      string `json:"recommended_price_eur"`
	PriceConfidence          string `json:"price_confidence"`
	PriceBasis               string `json:"price_basis"`
	Notes                    string `json:"notes"`
}

type Row struct {
	SourceImage              string
	Artist                   string
	Title                    string
	IdentificationConfidence string
	RecommendedPriceEUR      string
	PriceConfidence          string
	PriceBasis               string
	Notes                    string
}

var Header = []string{"source_image", "artist", "title", "identification_confidence", "recommended_price_eur", "price_confidence", "price_basis", "notes"}
