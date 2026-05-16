package imageinput

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

var supportedExtensions = map[string]struct{}{
	".dng": {}, ".heic": {}, ".heif": {}, ".jpg": {}, ".jpeg": {},
	".png": {}, ".tif": {}, ".tiff": {}, ".webp": {},
}

func IsSupportedImage(path string) bool {
	_, ok := supportedExtensions[strings.ToLower(filepath.Ext(path))]
	return ok
}

func Collect(sourceDir string, image string, allImages bool) ([]string, error) {
	if image == "" && !allImages {
		return nil, errors.New("choose a specific image or all images")
	}
	if image != "" && allImages {
		return nil, errors.New("choose only one image mode")
	}
	if image != "" {
		resolved := image
		if _, err := os.Stat(resolved); err != nil {
			resolved = filepath.Join(sourceDir, image)
		}
		info, err := os.Stat(resolved)
		if err != nil {
			return nil, fmt.Errorf("image not found: %s", image)
		}
		if info.IsDir() || !IsSupportedImage(resolved) {
			return nil, fmt.Errorf("unsupported image: %s", resolved)
		}
		return []string{resolved}, nil
	}

	entries, err := os.ReadDir(sourceDir)
	if err != nil {
		return nil, fmt.Errorf("source directory not found: %s", sourceDir)
	}
	images := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		path := filepath.Join(sourceDir, entry.Name())
		if IsSupportedImage(path) {
			images = append(images, path)
		}
	}
	sort.Strings(images)
	return images, nil
}
