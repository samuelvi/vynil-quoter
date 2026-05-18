package visionpayload_test

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/color"
	"image/jpeg"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"vinylquoter/internal/provider/visionpayload"
)

func TestPromptDefinesStrictPricingJSONContract(t *testing.T) {
	prompt := visionpayload.Prompt()
	for _, want := range []string{"exact shape", "identification_confidence", "recommended_price_eur", "Spain/EU market", "media VG", "Do not include markdown or commentary"} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("prompt missing %q: %s", want, prompt)
		}
	}
}

func TestPromptIncludesSelectedConditionAndNumericPriceContract(t *testing.T) {
	prompt := visionpayload.PromptForCondition("VG+", "G+")

	for _, want := range []string{"media VG+", "sleeve G+", "numbers only", "without currency", "12-18"} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("prompt missing %q: %s", want, prompt)
		}
	}
}

func TestInlineImageReturnsDownscaledJPEGPreview(t *testing.T) {
	path := writeJPEG(t, 2000, 1600)
	mimeType, encoded, err := visionpayload.InlineImage(path)
	if err != nil {
		t.Fatal(err)
	}
	if mimeType != "image/jpeg" {
		t.Fatalf("expected image/jpeg, got %s", mimeType)
	}
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		t.Fatal(err)
	}
	config, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	if max(config.Width, config.Height) > 1400 {
		t.Fatalf("expected max side <= 1400, got %dx%d", config.Width, config.Height)
	}
}

func writeJPEG(t *testing.T, width int, height int) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "input.jpg")
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.RGBA{R: 80, G: 90, B: 100, A: 255})
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
