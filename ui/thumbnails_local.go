package ui

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"fyne.io/fyne"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/storage"
	"fyne.io/fyne/widget"
	"github.com/okratitan/fyfoto/internal/filesystem"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type LocalThumbnailTable struct {
	*thumbnailTable
	root fyne.URI
	uris map[string]fyne.URI
}

func NewLocalThumbnailTable(callback func(id string, uri fyne.URI)) *LocalThumbnailTable {
	t := &LocalThumbnailTable{
		thumbnailTable: newThumbnailTable(),
		uris:           make(map[string]fyne.URI),
	}
	t.UpdateCell = func(id widget.TableCellID, item fyne.CanvasObject) {
		c, ok := item.(*fyne.Container)
		if !ok {
			return
		}
		index := t.idToIndex(id)
		if index < 0 || index >= len(t.ids) {
			c.Objects = nil
			c.Refresh()
			return
		}
		i := t.ids[index]
		t.RLock()
		thumbnail, ok := t.thumbnails[i]
		t.RUnlock()
		if ok && thumbnail != nil {
			c.Objects = []fyne.CanvasObject{thumbnail}
			c.Refresh()
		} else if uri, ok := t.uris[i]; ok {
			// TODO add file icon thumbnail c.Objects = []fyne.CanvasObject{default}
			c.Refresh()
			go t.createThumbnail(i, uri)
		}
	}
	t.OnSelected = func(id widget.TableCellID) {
		index := t.idToIndex(id)
		if index < 0 || index >= len(t.ids) {
			return
		}
		i := t.ids[index]
		if u, ok := t.uris[i]; ok && callback != nil {
			callback(i, u)
		}
		t.Unselect(id) // TODO FIXME Hack
	}
	t.ExtendBaseWidget(t)
	return t
}

func (t *LocalThumbnailTable) AddURI(uri fyne.URI) error {
	if !strings.HasPrefix(uri.MimeType(), "image/") {
		return nil
	}
	id := uri.String()
	if _, ok := t.uris[id]; !ok {
		t.uris[id] = uri
		t.ids = append(t.ids, id)
	}
	return nil
}

func (t *LocalThumbnailTable) Update(root fyne.URI) error {
	if t.root != root {
		t.root = root
		t.ids = nil
	}
	luri, err := storage.ListerForURI(root)
	if err != nil {
		return err
	}
	uris, err := luri.List()
	if err != nil {
		return err
	}

	for _, u := range uris {
		if strings.HasPrefix(u.Name(), ".") {
			continue
		}
		t.AddURI(u)
	}
	t.Refresh()
	return nil
}

func (t *LocalThumbnailTable) createThumbnail(id string, uri fyne.URI) {
	thumbnailCache, err := filesystem.ThumbnailCache()
	if err != nil {
		return
	}
	newfileMD5 := md5.New()
	newfileMD5.Write([]byte(id))
	newfile := hex.EncodeToString(newfileMD5.Sum(nil))
	destfile := filepath.Join(thumbnailCache, newfile+".png")

	needThumb := 1
	th, err := os.Stat(destfile)
	file := strings.TrimPrefix(id, "file://")
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
		// TODO if the same file can't be opened/loaded 3 times then add id to map and stop trying
		img, err := os.Open(file)
		if err != nil {
			fmt.Println("Could not open image to thumbnail")
			img.Close()
			return
		}
		src, _, err := image.Decode(img)
		if err != nil {
			fmt.Println("Could not decode source image for thumbnail")
			img.Close()
			return
		}
		img.Close()
		writeThumbnail(src, destfile)
	}
	image := canvas.NewImageFromFile(destfile)
	image.FillMode = canvas.ImageFillContain
	size := fyne.NewSize(ThumbnailSize, ThumbnailSize)
	image.SetMinSize(size)
	image.Resize(size)
	t.Lock()
	t.thumbnails[id] = image
	t.Unlock()
	t.Refresh()
}
