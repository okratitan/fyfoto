// +build linux,openbsd,freebsd,netbsd

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
	return userHome()
}

func cacheDirectory() (string, error) {
	xdgCache := os.Getenv("XDG_CACHE_HOME")
	if xdgCache == "" {
		return userCache()
	}
	return xdgCache, nil
}
