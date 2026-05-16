package crop

import (
	"image"
	"image/color"
	"image/jpeg"
	"os"
	"path/filepath"
	"testing"
)

func TestCropVinylWritesDominantObjectToDst(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "src", "DSC01.jpg")
	dstDir := filepath.Join(tmp, "dst")
	writeSyntheticVinylPhoto(t, src)

	result, err := Process(src, dstDir)
	if err != nil {
		t.Fatal(err)
	}
	if result.SourcePath != src {
		t.Fatalf("got %#v", result)
	}
	if result.CroppedPath != filepath.Join(dstDir, "DSC01.jpg") {
		t.Fatalf("got %#v", result)
	}
	info, err := os.Stat(result.CroppedPath)
	if err != nil {
		t.Fatal(err)
	}
	if info.Size() == 0 {
		t.Fatal("cropped image is empty")
	}
	cropped := decodeJPEG(t, result.CroppedPath)
	bounds := cropped.Bounds()
	if bounds.Dx() >= 300 || bounds.Dy() >= 300 {
		t.Fatalf("expected cropped image smaller than source, got %dx%d", bounds.Dx(), bounds.Dy())
	}
	if bounds.Dx() < 120 || bounds.Dy() < 120 {
		t.Fatalf("expected crop to retain vinyl object, got %dx%d", bounds.Dx(), bounds.Dy())
	}
}

func writeSyntheticVinylPhoto(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	img := image.NewRGBA(image.Rect(0, 0, 300, 300))
	for y := 0; y < 300; y++ {
		for x := 0; x < 300; x++ {
			img.Set(x, y, color.RGBA{R: 230, G: 230, B: 220, A: 255})
		}
	}
	for y := 82; y < 222; y++ {
		for x := 92; x < 232; x++ {
			img.Set(x, y, color.RGBA{R: 20, G: 30, B: 45, A: 255})
		}
	}
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	if err := jpeg.Encode(file, img, &jpeg.Options{Quality: 95}); err != nil {
		t.Fatal(err)
	}
}

func decodeJPEG(t *testing.T, path string) image.Image {
	t.Helper()
	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	img, err := jpeg.Decode(file)
	if err != nil {
		t.Fatal(err)
	}
	return img
}
