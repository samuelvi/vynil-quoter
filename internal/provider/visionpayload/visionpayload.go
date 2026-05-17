package visionpayload

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/jpeg"
	_ "image/png"
	"mime"
	"os"
	"path/filepath"
)

const maxPreviewSide = 1400

func Prompt() string {
	return "Identify the vinyl album from this cropped front cover image and estimate a conservative second-hand sale price. " +
		"Return only JSON with this exact shape: " +
		`{"artist":"string","title":"string",` +
		`"identification_confidence":"high|medium|low|manual-review",` +
		`"recommended_price_eur":"string",` +
		`"price_confidence":"high|medium|low|manual-review",` +
		`"price_basis":"string","notes":"string"}. ` +
		"Use Unknown for artist/title if the cover is unreadable or ambiguous. " +
		"Price assumptions: Spain/EU market, EUR, media VG+, sleeve VG, normal second-hand sale. " +
		"If the album or price is uncertain, use low or manual-review confidence and explain in notes. " +
		"Do not include markdown or commentary."
}

func InlineImage(imagePath string) (string, string, error) {
	data, err := os.ReadFile(imagePath)
	if err != nil {
		return "", "", err
	}
	mimeType := mime.TypeByExtension(filepath.Ext(imagePath))
	if mimeType == "" {
		mimeType = "image/jpeg"
	}
	preview, err := jpegPreview(data)
	if err != nil {
		return mimeType, base64.StdEncoding.EncodeToString(data), nil
	}
	return "image/jpeg", base64.StdEncoding.EncodeToString(preview), nil
}

func DataURL(imagePath string) (string, error) {
	mimeType, encoded, err := InlineImage(imagePath)
	if err != nil {
		return "", err
	}
	return "data:" + mimeType + ";base64," + encoded, nil
}

func jpegPreview(data []byte) ([]byte, error) {
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	img = scaleToMaxSide(img, maxPreviewSide)
	var output bytes.Buffer
	if err := jpeg.Encode(&output, img, &jpeg.Options{Quality: 90}); err != nil {
		return nil, err
	}
	return output.Bytes(), nil
}

func scaleToMaxSide(img image.Image, maxSide int) image.Image {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	if width <= maxSide && height <= maxSide {
		return img
	}
	if width <= 0 || height <= 0 {
		return img
	}
	scale := float64(maxSide) / float64(max(width, height))
	dstWidth := max(1, int(float64(width)*scale))
	dstHeight := max(1, int(float64(height)*scale))
	dst := image.NewRGBA(image.Rect(0, 0, dstWidth, dstHeight))
	for y := 0; y < dstHeight; y++ {
		sourceY := bounds.Min.Y + y*height/dstHeight
		for x := 0; x < dstWidth; x++ {
			sourceX := bounds.Min.X + x*width/dstWidth
			dst.Set(x, y, img.At(sourceX, sourceY))
		}
	}
	return dst
}
