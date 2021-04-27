package ui

import (
	"aletheiaware.com/bcgo"
	"aletheiaware.com/spaceclientgo"
	"aletheiaware.com/spacego"
	"encoding/base64"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"
	"github.com/okratitan/fyfoto/internal/filesystem"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type SpaceThumbnailTable struct {
	*thumbnailTable
	client     spaceclientgo.SpaceClient
	node       bcgo.Node
	metas      map[string]*spacego.Meta
	timestamps map[string]uint64
}

func NewSpaceThumbnailTable(client spaceclientgo.SpaceClient, callback func(id string, timestamp uint64, meta *spacego.Meta)) *SpaceThumbnailTable {
	t := &SpaceThumbnailTable{
		thumbnailTable: newThumbnailTable(),
		client:         client,
		metas:          make(map[string]*spacego.Meta),
		timestamps:     make(map[string]uint64),
	}
	t.UpdateCell = func(id widget.TableCellID, item fyne.CanvasObject) {
		index := t.idToIndex(id)
		if index < 0 || index >= len(t.ids) {
			return
		}
		if c, ok := item.(*fyne.Container); ok {
			i := t.ids[index]
			t.RLock()
			thumbnail, ok := t.thumbnails[i]
			t.RUnlock()
			if ok && thumbnail != nil {
				c.Objects = []fyne.CanvasObject{thumbnail}
				c.Refresh()
			} else if meta, ok := t.metas[i]; ok {
				go t.createThumbnail(i, meta)
			}
		}
	}
	t.OnSelected = func(id widget.TableCellID) {
		index := t.idToIndex(id)
		if index < 0 || index >= len(t.ids) {
			return
		}
		i := t.ids[index]
		if m, ok := t.metas[i]; ok && callback != nil {
			callback(i, t.timestamps[i], m)
		}
		t.Unselect(id) // TODO FIXME Hack
	}
	t.ExtendBaseWidget(t)
	return t
}

func (t *SpaceThumbnailTable) AddMeta(entry *bcgo.BlockEntry, meta *spacego.Meta) error {
	if !strings.HasPrefix(meta.Type, "image/") {
		return nil
	}
	id := base64.RawURLEncoding.EncodeToString(entry.RecordHash)
	if _, ok := t.metas[id]; !ok {
		t.metas[id] = meta
		t.timestamps[id] = entry.Record.Timestamp
		t.ids = append(t.ids, id)
		sort.Slice(t.ids, func(i, j int) bool {
			return t.timestamps[t.ids[i]] < t.timestamps[t.ids[j]]
		})
	}
	return nil
}

func (t *SpaceThumbnailTable) Clear() {
	t.ids = nil
	t.Refresh()
}

func (t *SpaceThumbnailTable) Update(node bcgo.Node) error {
	t.node = node
	if err := t.client.AllMetas(node, t.AddMeta); err != nil {
		return err
	}
	t.Refresh()
	return nil
}

func (t *SpaceThumbnailTable) createThumbnail(id string, meta *spacego.Meta) {
	thumbnailCache, err := filesystem.ThumbnailCache()
	if err != nil {
		return
	}
	destfile := filepath.Join(thumbnailCache, id+".png")

	needThumb := 1
	if _, err := os.Stat(destfile); !os.IsNotExist(err) {
		needThumb = 0
	}
	if needThumb == 1 {
		// TODO if the same file can't be read/decoded 3 times then add id to map and stop trying
		hash, err := base64.RawURLEncoding.DecodeString(id)
		if err != nil {
			fmt.Println(err)
			return
		}
		reader, err := t.client.ReadFile(t.node, hash)
		if err != nil {
			fmt.Println(err)
			return
		}
		src, _, err := image.Decode(reader)
		if err != nil {
			fmt.Println("Could not decode source image for thumbnail")
			return
		}
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
	return
}
