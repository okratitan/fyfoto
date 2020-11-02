// +build windows

package cache

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
)

func ImageCache() (string, error) {
	cacheDir, err := userCache()
	if err != nil {
		return "", err
	}
	imageDir := filepath.Join(cacheDir, "images")
	if err := os.MkdirAll(imageDir, 0700); err != nil {
		return "", err
	}
	return imageDir, nil
}

func ThumbnailCache() (string, error) {
	cacheDir, err := userCache()
	if err != nil {
		return "", err
	}
	thumbDir := filepath.Join(cacheDir, "thumbnails", "normal")
	if err := os.MkdirAll(thumbDir, 0700); err != nil {
		return "", err
	}
	return thumbDir, nil
}
