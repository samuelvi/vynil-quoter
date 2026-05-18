package lmstudio_test

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"image"
	"image/color"
	"image/jpeg"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"vinylquoter/internal/provider"
	"vinylquoter/internal/provider/lmstudio"
)

func TestParseResponseExtractsIdentification(t *testing.T) {
	body := []byte(`{"choices":[{"message":{"content":"{\"artist\":\"The Cure\",\"title\":\"Disintegration\",\"identification_confidence\":\"high\",\"recommended_price_eur\":\"22\",\"price_confidence\":\"medium\",\"price_basis\":\"EU\",\"notes\":\"ok\"}"}}]}`)
	got, err := lmstudio.ParseResponse(body)
	if err != nil {
		t.Fatal(err)
	}
	if got.Artist != "The Cure" || got.Title != "Disintegration" {
		t.Fatalf("got %#v", got)
	}
}

func TestParseResponseRejectsMissingChoice(t *testing.T) {
	if _, err := lmstudio.ParseResponse([]byte(`{"choices":[]}`)); err == nil {
		t.Fatal("expected error")
	}
}

func TestIdentifySendsModelFriendlyPromptAndPreviewImage(t *testing.T) {
	imagePath := writeLargeJPEG(t, 2200, 1600)
	var prompt string
	var previewMaxSide int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload struct {
			Messages []struct {
				Content []struct {
					Type     string `json:"type"`
					Text     string `json:"text"`
					ImageURL struct {
						URL string `json:"url"`
					} `json:"image_url"`
				} `json:"content"`
			} `json:"messages"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatal(err)
		}
		for _, part := range payload.Messages[0].Content {
			if part.Type == "text" {
				prompt = part.Text
			}
			if part.Type == "image_url" {
				previewMaxSide = decodedImageMaxSide(t, part.ImageURL.URL)
			}
		}
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"{\"artist\":\"A\",\"title\":\"T\",\"identification_confidence\":\"high\",\"recommended_price_eur\":\"12\",\"price_confidence\":\"medium\",\"price_basis\":\"Spain/EU; VG+/VG\",\"notes\":\"ok\"}"}}]}`))
	}))
	defer server.Close()

	_, err := lmstudio.Client{BaseURL: server.URL, Model: "test", HTTPClient: server.Client()}.Identify(context.Background(), provider.RecognitionRequest{ImagePath: imagePath, MediaCondition: "VG+", SleeveCondition: "G+"})
	if err != nil {
		t.Fatal(err)
	}

	for _, want := range []string{"exact shape", "Use Unknown", "Spain/EU market", "media VG+", "sleeve G+", "Do not include markdown or commentary"} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("prompt missing %q: %s", want, prompt)
		}
	}
	if previewMaxSide > 1400 {
		t.Fatalf("expected preview max side <= 1400, got %d", previewMaxSide)
	}
}

func writeLargeJPEG(t *testing.T, width int, height int) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "large.jpg")
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.RGBA{R: 120, G: 80, B: 60, A: 255})
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

func decodedImageMaxSide(t *testing.T, dataURL string) int {
	t.Helper()
	const prefix = "data:image/jpeg;base64,"
	if !strings.HasPrefix(dataURL, prefix) {
		t.Fatalf("expected JPEG data URL, got %s", dataURL[:min(len(dataURL), 32)])
	}
	data, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(dataURL, prefix))
	if err != nil {
		t.Fatal(err)
	}
	config, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	return max(config.Width, config.Height)
}
