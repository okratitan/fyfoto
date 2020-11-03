// +build android

package filesystem

import (
	"os"
	"path/filepath"
)

func ImageCache() (string, error) {
	cacheDir, err := cacheDirectory()
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
	cacheDir, err := cacheDirectory()
	if err != nil {
		return "", err
	}
	thumbDir := filepath.Join(cacheDir, "thumbnails", "normal")
	if err := os.MkdirAll(thumbDir, 0700); err != nil {
		return "", err
	}
	return thumbDir, nil
}

func RootDirectory() (string, error) {
	filesDir := os.Getenv("FILESDIR")
	if filesDir == "" {
		return userHome()
	}
	return filesDir, nil
}

func cacheDirectory() (string, error) {
	tempDir := os.Getenv("TMPDIR")
	if tempDir == "" {
		return userCache()
	}
	return tempDir, nil
}
