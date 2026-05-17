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
	"sort"
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
	imageFile, err := os.Open(sourcePath)
	if err != nil {
		return Result{}, err
	}
	defer imageFile.Close()

	img, _, err := image.Decode(imageFile)
	if err != nil {
		destinationPath := fallbackDestinationPathFor(sourcePath, destinationDir)
		if copyErr := copyFile(sourcePath, destinationPath); copyErr != nil {
			return Result{}, copyErr
		}
		return Result{SourcePath: sourcePath, CroppedPath: destinationPath, Fallback: true}, nil
	}

	destinationPath := jpgDestinationPathFor(sourcePath, destinationDir)
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

func jpgDestinationPathFor(sourcePath string, destinationDir string) string {
	base := filepath.Base(sourcePath)
	ext := filepath.Ext(base)
	name := strings.TrimSuffix(base, ext)
	if name == "" {
		name = base
	}
	return filepath.Join(destinationDir, name+".jpg")
}

func fallbackDestinationPathFor(sourcePath string, destinationDir string) string {
	base := filepath.Base(sourcePath)
	return filepath.Join(destinationDir, base)
}

func cropDominantObject(img image.Image) image.Image {
	bounds := img.Bounds()
	if bounds.Dx() < 8 || bounds.Dy() < 8 {
		return img
	}
	background := estimateBackground(img)
	threshold := 45.0
	foregroundRect, ok := dominantForegroundRect(img, background, threshold)
	if !ok {
		return centerCrop(img)
	}
	marginX := int(math.Round(float64(foregroundRect.Dx()) * 0.08))
	marginY := int(math.Round(float64(foregroundRect.Dy()) * 0.08))
	cropRect := image.Rect(
		max(bounds.Min.X, foregroundRect.Min.X-marginX),
		max(bounds.Min.Y, foregroundRect.Min.Y-marginY),
		min(bounds.Max.X, foregroundRect.Max.X+marginX),
		min(bounds.Max.Y, foregroundRect.Max.Y+marginY),
	)
	if cropRect.Dx() <= 0 || cropRect.Dy() <= 0 {
		return centerCrop(img)
	}
	if cropRect.Dx()*cropRect.Dy() >= int(float64(bounds.Dx()*bounds.Dy())*0.98) {
		return centerCrop(img)
	}
	return cloneCrop(img, cropRect)
}

func estimateBackground(img image.Image) color.Color {
	b := img.Bounds()
	if b.Empty() {
		return color.RGBA{A: 255}
	}
	step := max(1, min(b.Dx(), b.Dy())/200)
	reds := []int{}
	greens := []int{}
	blues := []int{}
	add := func(x int, y int) {
		r, g, bl, _ := img.At(x, y).RGBA()
		reds = append(reds, int(r>>8))
		greens = append(greens, int(g>>8))
		blues = append(blues, int(bl>>8))
	}
	minSide := min(b.Dx(), b.Dy())
	for _, offset := range []int{0, max(1, minSide/50), max(1, minSide/20), max(1, minSide/10)} {
		if offset*2 >= b.Dx() || offset*2 >= b.Dy() {
			continue
		}
		minX := b.Min.X + offset
		maxX := b.Max.X - offset
		minY := b.Min.Y + offset
		maxY := b.Max.Y - offset
		for x := minX; x < maxX; x += step {
			add(x, minY)
			add(x, maxY-1)
		}
		for y := minY + step; y < maxY-step; y += step {
			add(minX, y)
			add(maxX-1, y)
		}
	}
	sort.Ints(reds)
	sort.Ints(greens)
	sort.Ints(blues)
	mid := len(reds) / 2
	return color.RGBA{R: uint8(reds[mid]), G: uint8(greens[mid]), B: uint8(blues[mid]), A: 255}
}

type foregroundRange struct {
	start int
	end   int
}

func dominantForegroundRect(img image.Image, background color.Color, threshold float64) (image.Rectangle, bool) {
	b := img.Bounds()
	width := b.Dx()
	height := b.Dy()
	area := width * height
	columnCounts := make([]int, width)
	rowCounts := make([]int, height)
	foregroundCount := 0
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if colorDistance(img.At(b.Min.X+x, b.Min.Y+y), background) > threshold {
				columnCounts[x]++
				rowCounts[y]++
				foregroundCount++
			}
		}
	}
	if foregroundCount < max(16, area/100) {
		return image.Rectangle{}, false
	}
	xRange, ok := dominantCountRange(columnCounts, max(3, int(math.Round(float64(height)*0.08))))
	if !ok {
		return image.Rectangle{}, false
	}
	yRange, ok := dominantCountRange(rowCounts, max(3, int(math.Round(float64(width)*0.08))))
	if !ok {
		return image.Rectangle{}, false
	}
	cropRect := image.Rect(b.Min.X+xRange.start, b.Min.Y+yRange.start, b.Min.X+xRange.end, b.Min.Y+yRange.end)
	if cropRect.Dx()*cropRect.Dy() < max(16, area/100) || cropRect.Empty() {
		return image.Rectangle{}, false
	}
	return cropRect, true
}

func dominantCountRange(counts []int, minimumCount int) (foregroundRange, bool) {
	best := foregroundRange{}
	bestScore := 0
	bestLength := 0
	start := -1
	score := 0
	for index, count := range counts {
		if count >= minimumCount {
			if start == -1 {
				start = index
				score = 0
			}
			score += count
			continue
		}
		if start != -1 {
			best, bestScore, bestLength = betterRange(best, bestScore, bestLength, foregroundRange{start: start, end: index}, score)
			start = -1
		}
	}
	if start != -1 {
		best, bestScore, bestLength = betterRange(best, bestScore, bestLength, foregroundRange{start: start, end: len(counts)}, score)
	}
	return best, bestScore > 0
}

func betterRange(best foregroundRange, bestScore int, bestLength int, candidate foregroundRange, candidateScore int) (foregroundRange, int, int) {
	candidateLength := candidate.end - candidate.start
	if candidateScore > bestScore || (candidateScore == bestScore && candidateLength > bestLength) {
		return candidate, candidateScore, candidateLength
	}
	return best, bestScore, bestLength
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
