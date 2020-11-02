package ui

import (
	"fmt"
	"fyne.io/fyne"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/container"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"
	"golang.org/x/image/draw"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	"math"
	"os"
	"sync"
)

const ThumbnailSize = 128

type thumbnailTable struct {
	widget.Table
	sync.RWMutex
	ids        []string
	thumbnails map[string]*canvas.Image // TODO create tappable image, handle selection with OnTapped instead of table.OnSelected
}

func newThumbnailTable() *thumbnailTable {
	t := &thumbnailTable{
		Table:      widget.Table{},
		thumbnails: make(map[string]*canvas.Image),
	}
	t.CreateCell = func() fyne.CanvasObject {
		image := canvas.NewImageFromResource(theme.FileImageIcon())
		image.FillMode = canvas.ImageFillContain
		size := fyne.NewSize(ThumbnailSize, ThumbnailSize)
		image.SetMinSize(size)
		image.Resize(size)
		return container.NewMax(image)
	}
	t.Length = t.countToRowCol
	return t
}

func (t *thumbnailTable) Count() int {
	return len(t.ids)
}

func (t *thumbnailTable) countToRowCol() (int, int) {
	rows := 0
	cols := 0
	if count := t.Count(); count > 0 {
		cols = fyne.Min(t.Size().Width/ThumbnailSize, count)
		rows = count / cols
		if count%cols > 0 {
			rows++
		}
	}
	return rows, cols
}

func (t *thumbnailTable) idToIndex(id widget.TableCellID) int {
	cols := t.Size().Width / ThumbnailSize
	index := id.Row*cols + id.Col
	return index
}

func writeThumbnail(img image.Image, output string) {
	bounds := img.Bounds()
	newWidth, newHeight := ThumbnailSize, ThumbnailSize
	if bounds.Dx() > bounds.Dy() {
		scale := float64(bounds.Dx()) / float64(bounds.Dy())
		newHeight = int(math.Round(float64(ThumbnailSize) / float64(scale)))
	} else if bounds.Dy() > bounds.Dx() {
		scale := float64(bounds.Dy()) / float64(bounds.Dx())
		newWidth = int(math.Round(float64(ThumbnailSize) / float64(scale)))
	}
	dest := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))
	draw.NearestNeighbor.Scale(dest, dest.Bounds(), img, bounds, draw.Src, nil)

	out, err := os.Create(output)
	if err != nil {
		fmt.Println("Could not create thumbnail destination file")
	}
	err = png.Encode(out, dest)
	if err != nil {
		fmt.Println("Could not encode png for thumbnail")
	}
}
