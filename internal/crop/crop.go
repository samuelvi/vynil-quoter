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

const (
	backgroundDistanceThreshold = 45.0
	dominantObjectMarginRatio   = 0.08
	darkCoverMarginRatio        = 0.04
	darkCoverMinimumSide        = 32
	darkCoverMinimumAreaRatio   = 0.08
	darkCoverMaximumAreaRatio   = 0.95
	darkCoverMinimumAspect      = 0.65
	darkCoverMaximumAspect      = 1.45
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
	if coverRect, ok := dominantDarkCoverRect(img); ok {
		return cloneCrop(img, coverRect)
	}
	background := estimateBackground(img)
	foregroundRect, ok := dominantForegroundRect(img, background, backgroundDistanceThreshold)
	if !ok {
		return centerCrop(img)
	}
	cropRect := expandRect(foregroundRect, bounds, dominantObjectMarginRatio)
	if cropRect.Dx() <= 0 || cropRect.Dy() <= 0 {
		return centerCrop(img)
	}
	if cropRect.Dx()*cropRect.Dy() >= int(float64(bounds.Dx()*bounds.Dy())*0.98) {
		return centerCrop(img)
	}
	return cloneCrop(img, cropRect)
}

func dominantDarkCoverRect(img image.Image) (image.Rectangle, bool) {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	if width < darkCoverMinimumSide || height < darkCoverMinimumSide {
		return image.Rectangle{}, false
	}
	threshold := darkPixelThreshold(img)
	rowCounts := countDarkRows(img, threshold)
	yRange, ok := dominantCountRange(rowCounts, max(3, int(math.Round(float64(width)*0.08))))
	if !ok {
		return image.Rectangle{}, false
	}
	columnCounts := countDarkColumns(img, threshold, yRange)
	xRange, ok := dominantCountRange(columnCounts, max(3, int(math.Round(float64(yRange.end-yRange.start)*0.08))))
	if !ok {
		return image.Rectangle{}, false
	}
	coverRect := image.Rect(bounds.Min.X+xRange.start, bounds.Min.Y+yRange.start, bounds.Min.X+xRange.end, bounds.Min.Y+yRange.end)
	coverRect = expandRect(coverRect, bounds, darkCoverMarginRatio)
	if !isUsableCoverRect(coverRect, bounds) {
		return image.Rectangle{}, false
	}
	return coverRect, true
}

func darkPixelThreshold(img image.Image) int {
	histogram := [256]int{}
	bounds := img.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			histogram[luminance(img.At(x, y))]++
		}
	}
	return otsuThreshold(histogram, bounds.Dx()*bounds.Dy())
}

func countDarkRows(img image.Image, threshold int) []int {
	bounds := img.Bounds()
	counts := make([]int, bounds.Dy())
	for y := 0; y < bounds.Dy(); y++ {
		for x := 0; x < bounds.Dx(); x++ {
			if luminance(img.At(bounds.Min.X+x, bounds.Min.Y+y)) <= threshold {
				counts[y]++
			}
		}
	}
	return counts
}

func countDarkColumns(img image.Image, threshold int, yRange foregroundRange) []int {
	bounds := img.Bounds()
	counts := make([]int, bounds.Dx())
	for y := yRange.start; y < yRange.end; y++ {
		for x := 0; x < bounds.Dx(); x++ {
			if luminance(img.At(bounds.Min.X+x, bounds.Min.Y+y)) <= threshold {
				counts[x]++
			}
		}
	}
	return counts
}

func isUsableCoverRect(rect image.Rectangle, bounds image.Rectangle) bool {
	if rect.Empty() || rect.Dx() < darkCoverMinimumSide || rect.Dy() < darkCoverMinimumSide {
		return false
	}
	area := bounds.Dx() * bounds.Dy()
	coverArea := rect.Dx() * rect.Dy()
	if coverArea < int(float64(area)*darkCoverMinimumAreaRatio) {
		return false
	}
	if coverArea > int(float64(area)*darkCoverMaximumAreaRatio) {
		return false
	}
	aspect := float64(rect.Dx()) / float64(rect.Dy())
	return aspect >= darkCoverMinimumAspect && aspect <= darkCoverMaximumAspect
}

func otsuThreshold(histogram [256]int, total int) int {
	sum := 0
	for value, count := range histogram {
		sum += value * count
	}
	backgroundWeight := 0
	backgroundSum := 0
	bestThreshold := 0
	bestVariance := -1.0
	for threshold, count := range histogram {
		backgroundWeight += count
		if backgroundWeight == 0 {
			continue
		}
		foregroundWeight := total - backgroundWeight
		if foregroundWeight == 0 {
			break
		}
		backgroundSum += threshold * count
		backgroundMean := float64(backgroundSum) / float64(backgroundWeight)
		foregroundMean := float64(sum-backgroundSum) / float64(foregroundWeight)
		variance := float64(backgroundWeight) * float64(foregroundWeight) * math.Pow(backgroundMean-foregroundMean, 2)
		if variance > bestVariance {
			bestVariance = variance
			bestThreshold = threshold
		}
	}
	return bestThreshold
}

func luminance(c color.Color) int {
	r, g, b, _ := c.RGBA()
	return (299*int(r>>8) + 587*int(g>>8) + 114*int(b>>8)) / 1000
}

func expandRect(rect image.Rectangle, bounds image.Rectangle, marginRatio float64) image.Rectangle {
	marginX := int(math.Round(float64(rect.Dx()) * marginRatio))
	marginY := int(math.Round(float64(rect.Dy()) * marginRatio))
	return image.Rect(
		max(bounds.Min.X, rect.Min.X-marginX),
		max(bounds.Min.Y, rect.Min.Y-marginY),
		min(bounds.Max.X, rect.Max.X+marginX),
		min(bounds.Max.Y, rect.Max.Y+marginY),
	)
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
