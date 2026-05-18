package app_test

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/jpeg"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"vinylquoter/internal/app"
	"vinylquoter/internal/catalog"
	"vinylquoter/internal/config"
	"vinylquoter/internal/provider"
)

type fakeRecognizer struct{}

func (fakeRecognizer) Identify(ctx context.Context, imagePath string) (catalog.Identification, error) {
	return catalog.Identification{Artist: "The Cure", Title: "Disintegration", IdentificationConfidence: "high", RecommendedPriceEUR: "22", PriceConfidence: "medium", PriceBasis: "test", Notes: "ok"}, nil
}

func writeTinyJPEG(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	img := image.NewRGBA(image.Rect(0, 0, 20, 20))
	for y := 0; y < 20; y++ {
		for x := 0; x < 20; x++ {
			img.Set(x, y, color.RGBA{R: 80, G: 40, B: 120, A: 255})
		}
	}
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	if err := jpeg.Encode(file, img, nil); err != nil {
		t.Fatal(err)
	}
}

func TestParseArgsDefaults(t *testing.T) {
	cfg, err := app.ParseArgs([]string{"--all"})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.SourceDir != config.DefaultSourceDir || cfg.ReportPath != config.DefaultReportPath || cfg.Provider != config.DefaultProvider {
		t.Fatalf("got %#v", cfg)
	}
}

func TestParseArgsDefaultsToVGConditions(t *testing.T) {
	cfg, err := app.ParseArgs([]string{"--all"})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.MediaCondition != config.DefaultCondition || cfg.SleeveCondition != config.DefaultCondition {
		t.Fatalf("got %#v", cfg)
	}
}

func TestParseArgsSupportsConditionFlags(t *testing.T) {
	cfg, err := app.ParseArgs([]string{"--all", "--media-condition", "VG+", "--sleeve-condition", "G+"})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.MediaCondition != config.ConditionVeryGoodPlus || cfg.SleeveCondition != config.ConditionGoodPlus {
		t.Fatalf("got %#v", cfg)
	}
}

func TestParseArgsRejectsInvalidCondition(t *testing.T) {
	_, err := app.ParseArgs([]string{"--all", "--media-condition", "BAD"})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "invalid media condition") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestProcessWritesRows(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "data", "src", "a.jpg")
	writeTinyJPEG(t, src)
	report := filepath.Join(tmp, "data", "report", "album_catalog.csv")
	dstDir := filepath.Join(tmp, "data", "dst")
	rows, err := app.Process(context.Background(), []string{src}, report, false, dstDir, fakeRecognizer{})
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].Artist != "The Cure" {
		t.Fatalf("got %#v", rows)
	}
}

func TestProcessAddsPriceReferenceURLs(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "data", "src", "DSC01.jpg")
	writeTinyJPEG(t, src)
	report := filepath.Join(tmp, "data", "report", "album_catalog.csv")
	dstDir := filepath.Join(tmp, "data", "dst")

	rows, err := app.Process(context.Background(), []string{src}, report, false, dstDir, recognizerFunc(func(ctx context.Context, imagePath string) (catalog.Identification, error) {
		return catalog.Identification{Artist: "The Cure", Title: "Disintegration"}, nil
	}))

	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 {
		t.Fatalf("got %#v", rows)
	}
	for name, got := range map[string]string{
		"discogs": rows[0].DiscogsReferenceURL,
		"ebay":    rows[0].EBayReferenceURL,
		"popsike": rows[0].PopsikeReferenceURL,
	} {
		if !strings.Contains(got, "The+Cure+Disintegration+vinyl+VG%2B+sleeve+VG%2B") {
			t.Fatalf("%s URL missing expected query: %s", name, got)
		}
	}
}

func TestProcessCropsSourceImageAndRecognizesDstImage(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "data", "src", "DSC01.jpg")
	writeTinyJPEG(t, src)
	report := filepath.Join(tmp, "data", "report", "album_catalog.csv")
	dstDir := filepath.Join(tmp, "data", "dst")
	var recognizedPath string
	rows, err := app.Process(context.Background(), []string{src}, report, false, dstDir, recognizerFunc(func(ctx context.Context, imagePath string) (catalog.Identification, error) {
		recognizedPath = imagePath
		return catalog.Identification{Artist: "A", Title: "T"}, nil
	}))
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 {
		t.Fatalf("got %#v", rows)
	}
	expectedDst := filepath.Join(dstDir, "DSC01.jpg")
	if recognizedPath != expectedDst {
		t.Fatalf("recognizer should receive cropped dst image, got %s", recognizedPath)
	}
	if _, err := os.Stat(expectedDst); err != nil {
		t.Fatal(err)
	}
}

func TestProcessUsesImageBasenameAsCSVIdentifier(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "data", "src", "DSC01.jpg")
	writeTinyJPEG(t, src)
	report := filepath.Join(tmp, "data", "report", "album_catalog.csv")
	dstDir := filepath.Join(tmp, "data", "dst")
	rows, err := app.Process(context.Background(), []string{src}, report, false, dstDir, fakeRecognizer{})
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].SourceImage != "DSC01.jpg" {
		t.Fatalf("got %#v", rows)
	}
	content, err := os.ReadFile(report)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(content), "data/src/DSC01.jpg") {
		t.Fatalf("CSV should use basename identifier, got %s", string(content))
	}
}

func TestProcessReprocessesExistingBasenameIdentifierAndUpdatesCSV(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "data", "src", "DSC01.jpg")
	writeTinyJPEG(t, src)
	report := filepath.Join(tmp, "data", "report", "album_catalog.csv")
	dstDir := filepath.Join(tmp, "data", "dst")
	if err := catalog.Write(report, []catalog.Row{{SourceImage: "DSC01.jpg", Artist: "Existing", Title: "Old"}}); err != nil {
		t.Fatal(err)
	}
	identifyCalls := 0

	rows, err := app.Process(context.Background(), []string{src}, report, false, dstDir, recognizerFunc(func(ctx context.Context, imagePath string) (catalog.Identification, error) {
		identifyCalls++
		return catalog.Identification{Artist: "Fresh", Title: "New", IdentificationConfidence: "high"}, nil
	}))

	if err != nil {
		t.Fatal(err)
	}
	if identifyCalls != 1 {
		t.Fatalf("expected existing image to be recognized again, calls=%d", identifyCalls)
	}
	if len(rows) != 1 || rows[0].Artist != "Fresh" || rows[0].Title != "New" {
		t.Fatalf("got %#v", rows)
	}
	written, err := catalog.Read(report)
	if err != nil {
		t.Fatal(err)
	}
	if len(written) != 1 || written[0].Artist != "Fresh" || written[0].Title != "New" {
		t.Fatalf("CSV row should be replaced, got %#v", written)
	}
}

func TestProcessOverwritesDstImageWhenReprocessingExistingCSVRow(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "data", "src", "DSC01.jpg")
	writeTinyJPEG(t, src)
	report := filepath.Join(tmp, "data", "report", "album_catalog.csv")
	dstDir := filepath.Join(tmp, "data", "dst")
	dst := filepath.Join(dstDir, "DSC01.jpg")
	if err := os.MkdirAll(dstDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dst, []byte("stale dst image"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := catalog.Write(report, []catalog.Row{{SourceImage: "DSC01.jpg", Artist: "Existing"}}); err != nil {
		t.Fatal(err)
	}

	_, err := app.Process(context.Background(), []string{src}, report, false, dstDir, fakeRecognizer{})

	if err != nil {
		t.Fatal(err)
	}
	content, err := os.ReadFile(dst)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) == "stale dst image" {
		t.Fatal("dst image was not overwritten during reprocessing")
	}
}

func TestParseArgsSingleImageNeverReplacesCSV(t *testing.T) {
	cfg, err := app.ParseArgs([]string{"--image", "DSC01.jpg", "--replace"})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Replace {
		t.Fatalf("single image CLI mode must never replace CSV: %#v", cfg)
	}
}

func TestParseArgsSupportsAlternateLMStudioVisionModel(t *testing.T) {
	cfg, err := app.ParseArgs([]string{"--all", "--model", config.AlternateLMStudioModel})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Provider != config.ProviderLMStudio || cfg.Model != config.AlternateLMStudioModel {
		t.Fatalf("got %#v", cfg)
	}
}

func TestRunInteractiveReturnsToMenuUntilExit(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "data", "src")
	if err := os.MkdirAll(src, 0o755); err != nil {
		t.Fatal(err)
	}
	image := filepath.Join(src, "DSC01.jpg")
	if err := os.WriteFile(image, []byte("jpg"), 0o644); err != nil {
		t.Fatal(err)
	}
	report := filepath.Join(tmp, "data", "report", "album_catalog.csv")
	stdin := bytes.NewBufferString("1\n" + image + "\n7\n")
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cfg := config.DefaultRunConfig()
	cfg.SourceDir = src
	cfg.ReportPath = report
	cfg.DestinationDir = filepath.Join(tmp, "data", "dst")
	code := app.RunWithRecognizerFactory(context.Background(), cfg, stdin, stdout, stderr, func(config.RunConfig) (provider.Recognizer, error) {
		return fakeRecognizer{}, nil
	})
	if code != 0 {
		t.Fatalf("code=%d stderr=%s", code, stderr.String())
	}
	if strings.Count(stdout.String(), "Vinyl Quoter") < 2 {
		t.Fatalf("menu should be shown again before exit, got %s", stdout.String())
	}
}

func TestRunInteractivePersistsSelectedModelAndCSV(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "data", "src")
	if err := os.MkdirAll(src, 0o755); err != nil {
		t.Fatal(err)
	}
	image := filepath.Join(src, "DSC01.jpg")
	if err := os.WriteFile(image, []byte("jpg"), 0o644); err != nil {
		t.Fatal(err)
	}
	report := filepath.Join(tmp, "custom.csv")
	stdin := bytes.NewBufferString("4\n3\n3\n1\n" + report + "\n1\n" + image + "\n7\n")
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cfg := config.DefaultRunConfig()
	cfg.SourceDir = src
	cfg.DestinationDir = filepath.Join(tmp, "data", "dst")
	var seen []config.RunConfig
	code := app.RunWithRecognizerFactory(context.Background(), cfg, stdin, stdout, stderr, func(cfg config.RunConfig) (provider.Recognizer, error) {
		seen = append(seen, cfg)
		return fakeRecognizer{}, nil
	})
	if code != 0 {
		t.Fatalf("code=%d stderr=%s", code, stderr.String())
	}
	if len(seen) != 1 {
		t.Fatalf("got configs %#v", seen)
	}
	if seen[0].ReportPath != report || seen[0].Provider != config.ProviderGemini || seen[0].Model != config.DefaultGeminiModel {
		t.Fatalf("state was not persisted: %#v", seen[0])
	}
}

func TestRunInteractivePersistsSelectedConditions(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "data", "src")
	image := filepath.Join(src, "DSC01.jpg")
	writeTinyJPEG(t, image)
	report := filepath.Join(tmp, "data", "report", "album_catalog.csv")
	stdin := bytes.NewBufferString("5\n3\n6\n5\n1\n" + image + "\n7\n")
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cfg := config.DefaultRunConfig()
	cfg.SourceDir = src
	cfg.ReportPath = report
	cfg.DestinationDir = filepath.Join(tmp, "data", "dst")
	var seen []config.RunConfig
	code := app.RunWithRecognizerFactory(context.Background(), cfg, stdin, stdout, stderr, func(cfg config.RunConfig) (provider.Recognizer, error) {
		seen = append(seen, cfg)
		return fakeRecognizer{}, nil
	})
	if code != 0 {
		t.Fatalf("code=%d stderr=%s", code, stderr.String())
	}
	if len(seen) != 1 {
		t.Fatalf("got configs %#v", seen)
	}
	if seen[0].SleeveCondition != "VG+" || seen[0].MediaCondition != "G+" {
		t.Fatalf("condition state was not persisted: %#v", seen[0])
	}
}

func TestRunInteractiveOptionTwoProcessesAllImages(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "data", "src")
	if err := os.MkdirAll(src, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"DSC01.jpg", "DSC02.jpg"} {
		if err := os.WriteFile(filepath.Join(src, name), []byte("jpg"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	report := filepath.Join(tmp, "data", "report", "album_catalog.csv")
	stdin := bytes.NewBufferString("2\n7\n")
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cfg := config.DefaultRunConfig()
	cfg.SourceDir = src
	cfg.ReportPath = report
	cfg.DestinationDir = filepath.Join(tmp, "data", "dst")
	var seen []string
	code := app.RunWithRecognizerFactory(context.Background(), cfg, stdin, stdout, stderr, func(config.RunConfig) (provider.Recognizer, error) {
		return recognizerFunc(func(ctx context.Context, imagePath string) (catalog.Identification, error) {
			seen = append(seen, filepath.Base(imagePath))
			return catalog.Identification{Artist: "A", Title: "T"}, nil
		}), nil
	})
	if code != 0 {
		t.Fatalf("code=%d stderr=%s", code, stderr.String())
	}
	if strings.Join(seen, ",") != "DSC01.jpg,DSC02.jpg" {
		t.Fatalf("option 2 should process all images, got %#v", seen)
	}
	if !strings.Contains(stdout.String(), "Procesando todas las imágenes") {
		t.Fatalf("option 2 should show visible progress, got %s", stdout.String())
	}
}

func TestRunInteractiveResetsActionModeAfterOptionTwo(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "data", "src")
	if err := os.MkdirAll(src, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"DSC01.jpg", "DSC02.jpg"} {
		if err := os.WriteFile(filepath.Join(src, name), []byte("jpg"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	report := filepath.Join(tmp, "data", "report", "album_catalog.csv")
	stdin := bytes.NewBufferString("2\n1\nDSC01.jpg\n7\n")
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cfg := config.DefaultRunConfig()
	cfg.SourceDir = src
	cfg.ReportPath = report
	cfg.DestinationDir = filepath.Join(tmp, "data", "dst")
	var seen []string
	code := app.RunWithRecognizerFactory(context.Background(), cfg, stdin, stdout, stderr, func(config.RunConfig) (provider.Recognizer, error) {
		return recognizerFunc(func(ctx context.Context, imagePath string) (catalog.Identification, error) {
			seen = append(seen, filepath.Base(imagePath))
			return catalog.Identification{Artist: "A", Title: "T"}, nil
		}), nil
	})
	if code != 0 {
		t.Fatalf("code=%d stderr=%s", code, stderr.String())
	}
	if stderr.String() != "" {
		t.Fatalf("unexpected stderr: %s", stderr.String())
	}
	if strings.Join(seen, ",") != "DSC01.jpg,DSC02.jpg,DSC01.jpg" {
		t.Fatalf("option 1 should reprocess the selected image after option 2, got %#v", seen)
	}
}

type recognizerFunc func(context.Context, string) (catalog.Identification, error)

func (fn recognizerFunc) Identify(ctx context.Context, imagePath string) (catalog.Identification, error) {
	return fn(ctx, imagePath)
}
