package crop

import (
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	_ "image/png"
	"math"
	"os"
	"path/filepath"
	"strings"
)

type Result struct {
	SourcePath  string
	CroppedPath string
	Fallback    bool
}

func Process(sourcePath string, destinationDir string) (Result, error) {
	if err := os.MkdirAll(destinationDir, 0o755); err != nil {
		return Result{}, err
	}
	destinationPath := destinationPathFor(sourcePath, destinationDir)
	imageFile, err := os.Open(sourcePath)
	if err != nil {
		return Result{}, err
	}
	defer imageFile.Close()

	img, _, err := image.Decode(imageFile)
	if err != nil {
		if copyErr := copyFile(sourcePath, destinationPath); copyErr != nil {
			return Result{}, copyErr
		}
		return Result{SourcePath: sourcePath, CroppedPath: destinationPath, Fallback: true}, nil
	}

	cropped := cropDominantObject(img)
	output, err := os.Create(destinationPath)
	if err != nil {
		return Result{}, err
	}
	defer output.Close()
	if err := jpeg.Encode(output, cropped, &jpeg.Options{Quality: 92}); err != nil {
		return Result{}, err
	}
	return Result{SourcePath: sourcePath, CroppedPath: destinationPath}, nil
}

func destinationPathFor(sourcePath string, destinationDir string) string {
	base := filepath.Base(sourcePath)
	ext := filepath.Ext(base)
	if strings.EqualFold(ext, ".jpg") || strings.EqualFold(ext, ".jpeg") || strings.EqualFold(ext, ".png") {
		base = strings.TrimSuffix(base, ext) + ".jpg"
	}
	return filepath.Join(destinationDir, base)
}

func cropDominantObject(img image.Image) image.Image {
	bounds := img.Bounds()
	if bounds.Dx() < 8 || bounds.Dy() < 8 {
		return img
	}
	background := estimateBackground(img)
	threshold := 45.0
	minX, minY := bounds.Max.X, bounds.Max.Y
	maxX, maxY := bounds.Min.X, bounds.Min.Y
	foreground := 0
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			if colorDistance(img.At(x, y), background) > threshold {
				foreground++
				if x < minX {
					minX = x
				}
				if y < minY {
					minY = y
				}
				if x > maxX {
					maxX = x
				}
				if y > maxY {
					maxY = y
				}
			}
		}
	}
	area := bounds.Dx() * bounds.Dy()
	if foreground < area/50 || minX >= maxX || minY >= maxY {
		return centerCrop(img)
	}
	marginX := int(math.Round(float64(maxX-minX) * 0.08))
	marginY := int(math.Round(float64(maxY-minY) * 0.08))
	cropRect := image.Rect(
		max(bounds.Min.X, minX-marginX),
		max(bounds.Min.Y, minY-marginY),
		min(bounds.Max.X, maxX+marginX+1),
		min(bounds.Max.Y, maxY+marginY+1),
	)
	if cropRect.Dx() <= 0 || cropRect.Dy() <= 0 {
		return centerCrop(img)
	}
	return cloneCrop(img, cropRect)
}

func estimateBackground(img image.Image) color.Color {
	b := img.Bounds()
	points := []image.Point{
		{X: b.Min.X, Y: b.Min.Y},
		{X: b.Max.X - 1, Y: b.Min.Y},
		{X: b.Min.X, Y: b.Max.Y - 1},
		{X: b.Max.X - 1, Y: b.Max.Y - 1},
	}
	var r, g, bl uint32
	for _, point := range points {
		pr, pg, pb, _ := img.At(point.X, point.Y).RGBA()
		r += pr >> 8
		g += pg >> 8
		bl += pb >> 8
	}
	return color.RGBA{R: uint8(r / 4), G: uint8(g / 4), B: uint8(bl / 4), A: 255}
}

func colorDistance(a color.Color, b color.Color) float64 {
	ar, ag, ab, _ := a.RGBA()
	br, bg, bb, _ := b.RGBA()
	dr := float64(int(ar>>8) - int(br>>8))
	dg := float64(int(ag>>8) - int(bg>>8))
	db := float64(int(ab>>8) - int(bb>>8))
	return math.Sqrt(dr*dr + dg*dg + db*db)
}

func centerCrop(img image.Image) image.Image {
	b := img.Bounds()
	size := min(b.Dx(), b.Dy())
	x0 := b.Min.X + (b.Dx()-size)/2
	y0 := b.Min.Y + (b.Dy()-size)/2
	return cloneCrop(img, image.Rect(x0, y0, x0+size, y0+size))
}

func cloneCrop(img image.Image, rect image.Rectangle) image.Image {
	dst := image.NewRGBA(image.Rect(0, 0, rect.Dx(), rect.Dy()))
	draw.Draw(dst, dst.Bounds(), img, rect.Min, draw.Src)
	return dst
}

func copyFile(sourcePath string, destinationPath string) error {
	data, err := os.ReadFile(sourcePath)
	if err != nil {
		return err
	}
	return os.WriteFile(destinationPath, data, 0o644)
}

func min(a int, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a int, b int) int {
	if a > b {
		return a
	}
	return b
}
