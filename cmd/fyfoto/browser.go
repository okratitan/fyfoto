package main

import (
	"fmt"
	"io/ioutil"
	"math"
	"path"
	"path/filepath"
	"strings"

	"fyne.io/fyne"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/dialog"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/storage"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"
)

type gridImage struct {
	widget.Box

	imageFile   string
	imageDir    string
	imageObject *canvas.Image

	ff *FyFoto
}

func (gi *gridImage) Tapped(*fyne.PointEvent) {
	hideBrowser(gi.ff)
	showViewer(gi.ff, gi.imageFile)
}

func (gi *gridImage) TappedSecondary(*fyne.PointEvent) {
}

func fileIsImage(file string) bool {
	mimes := &storage.MimeTypeFileFilter{MimeTypes: []string{"image/*"}}
	return mimes.Matches(storage.NewURI("file://" + file))
}

func populate(ff *FyFoto, dir string) {
	ff.directories = nil
	ff.images.Objects = nil
	if widget.Renderer(ff.bInfo) != nil {
		ff.bInfo.SetText("Loading Images")
	}
	ff.dirs.Refresh()
	canvas.Refresh(ff.images)

	i := 0
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		fmt.Println("Could not determine directory contents")
	}

	thumbQueue := make(chan gridImage, len(files))
	quitQueue := make(chan string, 4)
	for workers := 0; workers < 4; workers++ {
		go thumbnail(ff, thumbQueue, quitQueue)
	}

	ff.up.OnTapped = func() {
		if ff.currentDir == "/" {
			return
		}
		parentDir := filepath.Dir(dir)
		thumbQueue = nil
		ff.currentDir = parentDir
		for workers := 0; workers < 4; workers++ {
			quitQueue <- "stop"
		}
		go populate(ff, parentDir)
	}

	ff.dirs.OnItemSelected = func(index int) {
		thumbQueue = nil
		ff.currentDir = ff.directories[index]
		for workers := 0; workers < 4; workers++ {
			quitQueue <- "stop"
		}
		go populate(ff, ff.currentDir)
	}

	for _, f := range files {
		if strings.HasPrefix(f.Name(), ".") == false {
			if f.IsDir() {
				newDir := dir + "/" + f.Name()
				ff.directories = append(ff.directories, newDir)
				ff.dirs.Refresh()
			} else if fileIsImage(dir + "/" + f.Name()) {
				gi := &gridImage{imageFile: dir + "/" + f.Name(), imageDir: dir, ff: ff}
				thumbQueue <- *gi
				i++
			}
		}
	}
	output := "No Images"
	if i > 0 {
		output = fmt.Sprintf("Total: %d Images", i)
	}
	ff.bInfo.SetText(output)
}

func hideFolders(ff *FyFoto) {
	ff.dirs.Hide()
	ff.dirBox.Hide()
	ff.dirsHidden = 1
	canvas.Refresh(ff.browser)
}

func showFolders(ff *FyFoto) {
	ff.dirs.Show()
	ff.dirBox.Show()
	ff.dirsHidden = 0
	canvas.Refresh(ff.browser)
}

func hideBrowser(ff *FyFoto) {
	ff.images.Hide()
	ff.iScroller.Hide()
	ff.dirs.Hide()
	ff.dirBox.Hide()
	ff.bToolbar.Hide()
	ff.bInfo.Hide()
}

func showBrowser(ff *FyFoto, dir string) {
	ff.images.Show()
	ff.iScroller.Show()
	ff.dirs.Show()
	ff.dirBox.Show()
	ff.bToolbar.Show()
	ff.bInfo.Show()
	ff.window.SetTitle("FyFoto - " + dir)
	canvas.Refresh(ff.main)

	if dir != ff.currentDir {
		ff.currentDir = dir
		go populate(ff, dir)
	}
}

func createBrowser(ff *FyFoto) {
	ff.bToolbar = widget.NewToolbar(
		widget.NewToolbarAction(theme.FolderIcon(),
			func() {
				if ff.dirsHidden > 0 {
					showFolders(ff)
				} else {
					hideFolders(ff)
				}
			}),
		widget.NewToolbarSpacer(),
		widget.NewToolbarAction(theme.SettingsIcon(),
			func() {
				dialog.ShowInformation("About", "FyFoto - A Cross-Platform Image Application", ff.window)
			}))

	ff.up = widget.NewButtonWithIcon("..", theme.MoveUpIcon(), nil)
	ff.dirs = widget.NewList(func() int {
		return len(ff.directories)
	},
		func() fyne.CanvasObject {
			return fyne.NewContainerWithLayout(layout.NewHBoxLayout(), widget.NewIcon(theme.FolderIcon()), widget.NewLabel("Directory Name"))
		},
		func(index int, item fyne.CanvasObject) {
			item.(*fyne.Container).Objects[1].(*widget.Label).SetText(path.Base(ff.directories[index]))
		})
	ff.dirBox = fyne.NewContainerWithLayout(layout.NewBorderLayout(ff.up, nil, nil, nil), ff.up, ff.dirs)

	size := int(math.Floor(float64(128 * ff.window.Canvas().Scale())))
	ff.images = fyne.NewContainerWithLayout(layout.NewFixedGridLayout(fyne.NewSize(size, size)))
	ff.iScroller = widget.NewScrollContainer(ff.images)

	ff.bInfo = widget.NewLabelWithStyle("No Images", fyne.TextAlignCenter, fyne.TextStyle{})

	ff.browser = fyne.NewContainerWithLayout(layout.NewBorderLayout(ff.bToolbar, ff.bInfo, ff.dirBox, nil),
		ff.bToolbar, ff.bInfo, ff.dirBox, ff.iScroller)
}
