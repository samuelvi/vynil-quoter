package imageinput_test

import (
	"os"
	"path/filepath"
	"testing"
	"vinylquoter/internal/imageinput"
)

func TestCollectSingleImageUsesRequestedFile(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "data", "src")
	if err := os.MkdirAll(src, 0o755); err != nil {
		t.Fatal(err)
	}
	image := filepath.Join(src, "DSC01.jpg")
	if err := os.WriteFile(image, []byte("jpg"), 0o644); err != nil {
		t.Fatal(err)
	}

	items, err := imageinput.Collect(src, image, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0] != image {
		t.Fatalf("got %#v", items)
	}
}

func TestCollectAllImagesUsesSupportedFilesSorted(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "data", "src")
	if err := os.MkdirAll(src, 0o755); err != nil {
		t.Fatal(err)
	}
	for name, body := range map[string]string{"b.png": "png", "a.jpg": "jpg", "notes.txt": "ignore"} {
		if err := os.WriteFile(filepath.Join(src, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	items, err := imageinput.Collect(src, "", true)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{filepath.Join(src, "a.jpg"), filepath.Join(src, "b.png")}
	if len(items) != len(want) || items[0] != want[0] || items[1] != want[1] {
		t.Fatalf("got %#v want %#v", items, want)
	}
}

func TestIsSupportedImage(t *testing.T) {
	for _, path := range []string{"a.jpg", "a.jpeg", "a.png", "a.webp", "a.dng", "a.heic", "a.heif", "a.tif", "a.tiff"} {
		if !imageinput.IsSupportedImage(path) {
			t.Fatalf("expected supported: %s", path)
		}
	}
	if imageinput.IsSupportedImage("notes.txt") {
		t.Fatal("txt should not be supported")
	}
}
