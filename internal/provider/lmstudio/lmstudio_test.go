package lmstudio

import "testing"

func TestParseResponseExtractsIdentification(t *testing.T) {
	body := []byte(`{"choices":[{"message":{"content":"{\"artist\":\"The Cure\",\"title\":\"Disintegration\",\"identification_confidence\":\"high\",\"recommended_price_eur\":\"22\",\"price_confidence\":\"medium\",\"price_basis\":\"EU\",\"notes\":\"ok\"}"}}]}`)
	got, err := ParseResponse(body)
	if err != nil {
		t.Fatal(err)
	}
	if got.Artist != "The Cure" || got.Title != "Disintegration" {
		t.Fatalf("got %#v", got)
	}
}

func TestParseResponseRejectsMissingChoice(t *testing.T) {
	if _, err := ParseResponse([]byte(`{"choices":[]}`)); err == nil {
		t.Fatal("expected error")
	}
}
