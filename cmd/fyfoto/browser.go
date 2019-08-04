package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"fyne.io/fyne"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/dialog"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"
)

func fileIsImage(file string) bool {
	fi, err := os.Open(file)
	if err != nil {
		fmt.Println("Could not read file")
	} else {
		buf := make([]byte, 512)
		_, err = fi.Read(buf)
		if err != nil {
			fmt.Println("Could not read file")
		} else {
			if strings.Contains(http.DetectContentType(buf), "image") {
				fi.Close()
				return true
			}
		}
	}
	fi.Close()
	return false
}

func populate(ff *FyFoto, dir string) {
	ff.dirs.Objects = nil
	ff.images.Objects = nil
	if widget.Renderer(ff.bInfo) != nil {
		ff.bInfo.SetText("Loading Images")
	}
	canvas.Refresh(ff.dirs)
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
	if ff.currentDir != "/" {
		parentDir := filepath.Dir(dir)
		up := widget.NewButtonWithIcon("..", theme.MoveUpIcon(),
			func() {
				thumbQueue = nil
				ff.currentDir = parentDir
				for workers := 0; workers < 4; workers++ {
					quitQueue <- "stop"
				}
				go populate(ff, parentDir)
			})
		ff.dirs.AddObject(up)
	}
	for _, f := range files {
		if strings.HasPrefix(f.Name(), ".") == false {
			if f.IsDir() {
				newDir := dir + "/" + f.Name()
				b := widget.NewButtonWithIcon(f.Name(), theme.FolderIcon(),
					func() {
						thumbQueue = nil
						ff.currentDir = newDir
						for workers := 0; workers < 4; workers++ {
							quitQueue <- "stop"
						}
						go populate(ff, newDir)
					})
				ff.dirs.AddObject(b)
			} else if fileIsImage(dir + "/" + f.Name()) {
				gi := &gridImage{imageFile: dir + "/" + f.Name(), imageDir: dir, ff: ff}
				thumbQueue <- *gi
				i++
			}
		}
	}
	canvas.Refresh(ff.dirs)
	output := "No Images"
	if i > 0 {
		output = fmt.Sprintf("Total: %d Images", i)
	}
	ff.bInfo.SetText(output)
}

func hideFolders(ff *FyFoto) {
	ff.dirs.Hide()
	ff.dScroller.Hide()
	ff.dirsHidden = 1
	canvas.Refresh(ff.browser)
}

func showFolders(ff *FyFoto) {
	ff.dirs.Show()
	ff.dScroller.Show()
	ff.dirsHidden = 0
	canvas.Refresh(ff.browser)
}

func hideBrowser(ff *FyFoto) {
	ff.images.Hide()
	ff.iScroller.Hide()
	ff.dirs.Hide()
	ff.dScroller.Hide()
	ff.bToolbar.Hide()
	ff.bInfo.Hide()
}

func showBrowser(ff *FyFoto, dir string) {
	ff.images.Show()
	ff.iScroller.Show()
	ff.dirs.Show()
	ff.dScroller.Show()
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
	ff.images = fyne.NewContainerWithLayout(layout.NewFixedGridLayout(fyne.NewSize(128, 128)))
	ff.dirs = fyne.NewContainerWithLayout(layout.NewVBoxLayout())

	ff.iScroller = widget.NewScrollContainer(ff.images)
	ff.dScroller = widget.NewScrollContainer(ff.dirs)

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

	ff.bInfo = widget.NewLabelWithStyle("No Images", fyne.TextAlignCenter, fyne.TextStyle{})

	ff.browser = fyne.NewContainerWithLayout(layout.NewBorderLayout(ff.bToolbar, ff.bInfo, ff.dScroller, nil),
		ff.bToolbar, ff.bInfo, ff.dScroller, ff.iScroller)
}
