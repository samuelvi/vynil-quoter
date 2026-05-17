package gemini_test

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"image"
	"image/color"
	"image/jpeg"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"vinylquoter/internal/provider/gemini"
)

func TestParseResponseExtractsIdentification(t *testing.T) {
	body := []byte(`{"candidates":[{"content":{"parts":[{"text":"{\"artist\":\"The Cure\",\"title\":\"Disintegration\",\"identification_confidence\":\"high\",\"recommended_price_eur\":\"22\",\"price_confidence\":\"medium\",\"price_basis\":\"EU\",\"notes\":\"ok\"}"}]}}]}`)
	got, err := gemini.ParseResponse(body)
	if err != nil {
		t.Fatal(err)
	}
	if got.Artist != "The Cure" || got.Title != "Disintegration" {
		t.Fatalf("got %#v", got)
	}
}

func TestParseResponseRejectsMissingText(t *testing.T) {
	if _, err := gemini.ParseResponse([]byte(`{"candidates":[]}`)); err == nil {
		t.Fatal("expected error")
	}
}

func TestIdentifySendsStrictPromptAndPreviewImage(t *testing.T) {
	imagePath := writeGeminiJPEG(t, 2200, 1600)
	transport := &captureTransport{t: t}
	client := gemini.Client{APIKey: "test-key", Model: "test-model", HTTPClient: &http.Client{Transport: transport}}

	_, err := client.Identify(context.Background(), imagePath)
	if err != nil {
		t.Fatal(err)
	}

	for _, want := range []string{"exact shape", "Use Unknown", "Spain/EU market", "media VG+", "Do not include markdown or commentary"} {
		if !strings.Contains(transport.prompt, want) {
			t.Fatalf("prompt missing %q: %s", want, transport.prompt)
		}
	}
	if transport.mimeType != "image/jpeg" {
		t.Fatalf("expected image/jpeg, got %s", transport.mimeType)
	}
	if transport.previewMaxSide > 1400 {
		t.Fatalf("expected preview max side <= 1400, got %d", transport.previewMaxSide)
	}
}

type captureTransport struct {
	t              *testing.T
	prompt         string
	mimeType       string
	previewMaxSide int
}

func (t *captureTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	var payload struct {
		Contents []struct {
			Parts []struct {
				Text       string `json:"text"`
				InlineData struct {
					MimeType string `json:"mime_type"`
					Data     string `json:"data"`
				} `json:"inline_data"`
			} `json:"parts"`
		} `json:"contents"`
	}
	if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
		t.t.Fatal(err)
	}
	for _, part := range payload.Contents[0].Parts {
		if part.Text != "" {
			t.prompt = part.Text
		}
		if part.InlineData.Data != "" {
			t.mimeType = part.InlineData.MimeType
			t.previewMaxSide = maxDecodedSide(t.t, part.InlineData.Data)
		}
	}
	return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(`{"candidates":[{"content":{"parts":[{"text":"{\"artist\":\"A\",\"title\":\"T\",\"identification_confidence\":\"high\",\"recommended_price_eur\":\"12\",\"price_confidence\":\"medium\",\"price_basis\":\"Spain/EU; VG+/VG\",\"notes\":\"ok\"}"}]}}]}`))}, nil
}

func writeGeminiJPEG(t *testing.T, width int, height int) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "large.jpg")
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.RGBA{R: 110, G: 70, B: 40, A: 255})
		}
	}
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	if err := jpeg.Encode(file, img, &jpeg.Options{Quality: 90}); err != nil {
		t.Fatal(err)
	}
	return path
}

func maxDecodedSide(t *testing.T, encoded string) int {
	t.Helper()
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		t.Fatal(err)
	}
	config, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	return max(config.Width, config.Height)
}
