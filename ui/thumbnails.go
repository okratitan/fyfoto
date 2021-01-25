package ui

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"golang.org/x/image/draw"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	"math"
	"os"
	"sync"
)

const ThumbnailSize = float32(128)

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
		cols = int(t.Size().Width / ThumbnailSize)
		if count < cols {
			cols = count
		}
		rows = count / cols
		if count%cols > 0 {
			rows++
		}
	}
	return rows, cols
}

func (t *thumbnailTable) idToIndex(id widget.TableCellID) int {
	cols := t.Size().Width / ThumbnailSize
	index := id.Row*int(cols) + id.Col
	return index
}

func writeThumbnail(img image.Image, output string) {
	bounds := img.Bounds()
	newWidth, newHeight := int(ThumbnailSize), int(ThumbnailSize)
	if bounds.Dx() > bounds.Dy() {
		scale := float32(bounds.Dx()) / float32(bounds.Dy())
		newHeight = int(math.Round(float64(ThumbnailSize / scale)))
	} else if bounds.Dy() > bounds.Dx() {
		scale := float32(bounds.Dy()) / float32(bounds.Dx())
		newWidth = int(math.Round(float64(ThumbnailSize / scale)))
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
