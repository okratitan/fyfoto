package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"image"
	"image/png"
	"math"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	"fyne.io/fyne"
	"fyne.io/fyne/canvas"

	"golang.org/x/image/draw"
)

func thumbnail(ff *FyFoto, thumbQueue <-chan gridImage, quitQueue <-chan string) {
	for {
		select {
		case <-quitQueue:
			return
		case gi := <-thumbQueue:
			file := strings.TrimPrefix(gi.image.String(), "file://")
			base := ""
			xdgcache := os.Getenv("XDG_CACHE_HOME")
			if xdgcache == "" {
				usr, err := user.Current()
				if err != nil {
					fmt.Println("Could not find the current user")
					return
				}
				base = usr.HomeDir + "/.cache"
			} else {
				base = xdgcache
			}
			os.Mkdir(base, 0700)
			thumbDir := base + "/thumbnails"
			os.Mkdir(thumbDir, 0700)
			thumbDir += "/normal"
			os.Mkdir(thumbDir, 0700)

			uri := []byte(gi.image.String())
			newfileMD5 := md5.New()
			newfileMD5.Write(uri)
			newfile := hex.EncodeToString(newfileMD5.Sum(nil))
			destfile := thumbDir + "/" + newfile + ".png"

			needThumb := 1
			th, err := os.Stat(destfile)
			if !os.IsNotExist(err) {
				orig, _ := os.Stat(file)
				modThumb := th.ModTime()
				modOrig := orig.ModTime()
				diff := modThumb.Sub(modOrig)
				if diff > (time.Duration(0) * time.Second) {
					needThumb = 0
				}
			}
			if needThumb == 1 {
				img, err := os.Open(file)
				if err != nil {
					fmt.Println("Could not open image to thumbnail")
					img.Close()
					break
				}
				src, _, err := image.Decode(img)
				if err != nil {
					fmt.Println("Could not decode source image for thumbnail")
					img.Close()
					break
				}
				img.Close()
				img, err = os.Open(file)
				if err != nil {
					fmt.Println("Could not open image to thumbnail")
					img.Close()
					break
				}
				cfg, _, err := image.DecodeConfig(img)
				if err != nil {
					fmt.Println("Could not get original image size")
					img.Close()
					break
				}
				img.Close()
				newWidth, newHeight := 128, 128
				if cfg.Width > cfg.Height {
					scale := float64(cfg.Width) / float64(cfg.Height)
					newHeight = int(math.Round(float64(newWidth) / float64(scale)))
				} else if cfg.Height > cfg.Width {
					scale := float64(cfg.Height) / float64(cfg.Width)
					newWidth = int(math.Round(float64(newHeight) / float64(scale)))
				}
				dest := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))
				draw.NearestNeighbor.Scale(dest, dest.Bounds(), src, src.Bounds(), draw.Src, nil)

				out, err := os.Create(destfile)
				if err != nil {
					fmt.Println("Could not create thumbnail destination file")
				}
				err = png.Encode(out, dest)
				if err != nil {
					fmt.Println("Could not encode png for thumbnail")
				}
			}
			if &gi != nil && "file://" + filepath.Dir(file) == ff.currentDir.String() {
				gi.imageObject = canvas.NewImageFromFile(destfile)
				gi.imageObject.FillMode = canvas.ImageFillContain
				size := int(math.Floor(float64(128 * ff.window.Canvas().Scale())))
				gi.imageObject.SetMinSize(fyne.NewSize(size, size))
				gi.Append(gi.imageObject)

				ff.images.AddObject(&gi)
				canvas.Refresh(ff.images)
			}
		}
	}
}
