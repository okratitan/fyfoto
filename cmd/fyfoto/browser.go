package main

import (
	"fmt"
	"math"
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

	image fyne.URI
	imageObject *canvas.Image

	ff *FyFoto
}

func (gi *gridImage) Tapped(*fyne.PointEvent) {
	hideBrowser(gi.ff)
	showViewer(gi.ff, gi.image)
}

func (gi *gridImage) TappedSecondary(*fyne.PointEvent) {
}

func fileIsImage(file fyne.URI) bool {
	mimes := &storage.MimeTypeFileFilter{MimeTypes: []string{"image/*"}}
	return mimes.Matches(file)
}

func populate(ff *FyFoto, dir fyne.URI) {
	ff.images.Objects = nil
	if widget.Renderer(ff.bInfo) != nil {
		ff.bInfo.SetText("Loading Images")
	}
	canvas.Refresh(ff.images)

	i := 0
	luri, err := storage.ListerForURI(dir)
	if err != nil {
		return
	} else {
		uris, err := luri.List()
		if err != nil {
			return
		}

		thumbQueue := make(chan gridImage, len(uris))
		quitQueue := make(chan string, 4)
		for workers := 0; workers < 4; workers++ {
			go thumbnail(ff, thumbQueue, quitQueue)
		}

		for _, u := range uris {
			if strings.HasPrefix(u.Name(), ".") == false {
				if fileIsImage(u) {
					gi := &gridImage{image: u, ff: ff}
					thumbQueue <- *gi
					i++
				}
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
	ff.dirsHidden = 1
	canvas.Refresh(ff.browser)
}

func showFolders(ff *FyFoto) {
	ff.dirs.Show()
	ff.dirsHidden = 0
	canvas.Refresh(ff.browser)
}

func hideBrowser(ff *FyFoto) {
	ff.images.Hide()
	ff.iScroller.Hide()
	ff.dirs.Hide()
	ff.bToolbar.Hide()
	ff.bInfo.Hide()
}

func showBrowser(ff *FyFoto, dir fyne.URI) {
	ff.images.Show()
	ff.iScroller.Show()
	ff.dirs.Show()
	ff.bToolbar.Show()
	ff.bInfo.Show()
	ff.window.SetTitle("FyFoto - " + dir.String())
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

	ff.dirs = &widget.Tree {
		Root: ff.rootDir.String(),
		IsBranch: func(uid string) bool {
			_, err := storage.ListerForURI(storage.NewURI(uid))
			return err == nil
		},
		CreateNode: func(branch bool) fyne.CanvasObject {
			var icon fyne.CanvasObject
			if branch {
				icon = widget.NewIcon(nil)
			} else {
				icon = widget.NewFileIcon(nil)
			}
			return fyne.NewContainerWithLayout(layout.NewHBoxLayout(), icon, widget.NewLabel("Template Object"))
		},
	}
	ff.dirs.ChildUIDs = func(uid string) (c []string) {
		luri, err := storage.ListerForURI(storage.NewURI(uid))
		if err != nil {
			fyne.LogError("Unable to get lister for "+uid, err)
		} else {
			uris, err := luri.List()
			if err != nil {
				return
			} else {
				// Filter URIs
				var us []fyne.URI
				for _, u := range uris {
					_, err := storage.ListerForURI(u)
					if err == nil && !strings.HasPrefix(u.Name(), ".") {
						us = append(us, u)
					}
				}
				// Convert to Strings
				for _, u := range us {
					c = append(c, u.String())
				}
			}
		}
		return
	}
	ff.dirs.UpdateNode = func(uid string, branch bool, node fyne.CanvasObject) {
		uri := storage.NewURI(uid)
		c := node.(*fyne.Container)
		if branch {
			var r fyne.Resource
			if ff.dirs.IsBranchOpen(uid) {
				// Set open folder icon
				r = theme.FolderOpenIcon()
			} else {
				// Set folder icon
				r = theme.FolderIcon()
			}
			c.Objects[0].(*widget.Icon).SetResource(r)
		} else {
			// Set file uri to update icon
			c.Objects[0].(*widget.FileIcon).SetURI(uri)
		}
		l := c.Objects[1].(*widget.Label)
		if ff.dirs.Root == uid {
			l.SetText(uid)
		} else {
			l.SetText(uri.Name())
		}
	}
	ff.dirs.OnSelected = func(uid string) {
		u := storage.NewURI(uid)
		ff.currentDir = u
		go populate(ff, u)
	}

	size := int(math.Floor(float64(128 * ff.window.Canvas().Scale())))
	ff.images = fyne.NewContainerWithLayout(layout.NewFixedGridLayout(fyne.NewSize(size, size)))
	ff.iScroller = widget.NewScrollContainer(ff.images)

	ff.bInfo = widget.NewLabelWithStyle("No Images", fyne.TextAlignCenter, fyne.TextStyle{})

	ff.browser = fyne.NewContainerWithLayout(layout.NewBorderLayout(ff.bToolbar, ff.bInfo, ff.dirs, nil),
		ff.bToolbar, ff.bInfo, ff.dirs, ff.iScroller)
}
