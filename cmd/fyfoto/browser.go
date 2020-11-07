package main

import (
	"encoding/base64"
	"fmt"
	"fyne.io/fyne"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/container"
	"fyne.io/fyne/dialog"
	"fyne.io/fyne/storage"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"
	bcui "github.com/AletheiaWareLLC/bcfynego/ui"
	bcuidata "github.com/AletheiaWareLLC/bcfynego/ui/data"
	"github.com/AletheiaWareLLC/bcgo"
	"github.com/AletheiaWareLLC/spacefynego"
	spaceuidata "github.com/AletheiaWareLLC/spacefynego/ui/data"
	"github.com/AletheiaWareLLC/spacego"
	"github.com/okratitan/fyfoto/internal/filesystem"
	"github.com/okratitan/fyfoto/ui"
	"os"
	"path/filepath"
)

func fileIsImage(file fyne.URI) bool {
	mimes := &storage.MimeTypeFileFilter{MimeTypes: []string{"image/*"}}
	return mimes.Matches(file)
}

func populateLocal(ff *FyFoto, dir fyne.URI) {
	// TODO Show Progress Bar

	// Update Status
	if widget.Renderer(ff.bInfo) != nil {
		ff.bInfo.SetText("Loading Images")
	}

	// Update Table
	ff.localImages.Update(dir)

	// Update Status
	output := "No Images"
	count := ff.localImages.Count()
	if count > 0 {
		output = fmt.Sprintf("Total: %d Images", count)
	}
	ff.bInfo.SetText(output)
}

func populateSpace(ff *FyFoto) {
	// TODO Show Progress Bar

	// Create Space Fyne
	f := spacefynego.NewSpaceFyne(ff.app, ff.window, ff.space)

	// Get BC Node
	n, err := f.GetNode(&ff.space.BCClient)
	if err != nil {
		f.ShowError(err)
		return
	}

	// Update Status
	if widget.Renderer(ff.bInfo) != nil {
		ff.bInfo.SetText("Loading Images")
	}

	// Update Table
	ff.spaceImages.Update(n)

	// Update Status
	output := "No Images"
	count := ff.spaceImages.Count()
	if count > 0 {
		output = fmt.Sprintf("Total: %d Images", count)
	}
	ff.bInfo.SetText(output)
}

func hideFolders(ff *FyFoto) {
	ff.localDirs.Hide()
	ff.dirsHidden = 1
	canvas.Refresh(ff.browser)
}

func showFolders(ff *FyFoto) {
	ff.localDirs.Show()
	ff.dirsHidden = 0
	canvas.Refresh(ff.browser)
}

func hideBrowser(ff *FyFoto) {
	ff.browser.Hide()
}

func showBrowser(ff *FyFoto, dir fyne.URI) {
	ff.browser.Show()
	ff.window.SetTitle("FyFoto - " + dir.String())
	canvas.Refresh(ff.main)

	if dir != ff.currentDir {
		ff.currentDir = dir
		go populateLocal(ff, dir)
	}
}

func showAbout(ff *FyFoto) {
	dialog.ShowInformation("About", "FyFoto - A Cross-Platform Image Application", ff.window)
}

func createBrowser(ff *FyFoto) {
	ff.localToolbar = widget.NewToolbar(
		widget.NewToolbarAction(theme.FolderIcon(), func() {
			if ff.dirsHidden > 0 {
				showFolders(ff)
			} else {
				hideFolders(ff)
			}
		}),
		widget.NewToolbarSpacer(),
		widget.NewToolbarAction(theme.InfoIcon(), func() {
			showAbout(ff)
		}),
	)
	ff.localDirs = ui.NewFileTree(ff.rootDir)
	ff.localDirs.OnSelected = func(uid string) {
		u := storage.NewURI(uid)
		ff.currentDir = u
		go populateLocal(ff, u)
	}
	ff.localImages = ui.NewLocalThumbnailTable(func(id string, uri fyne.URI) {
		hideBrowser(ff)
		showViewer(ff, uri.Name(), uri)
	})

	ff.spaceToolbar = widget.NewToolbar(
		widget.NewToolbarAction(theme.ContentAddIcon(), func() {
			f := spacefynego.NewSpaceFyne(ff.app, ff.window, ff.space)
			go func() {
				node, err := f.GetNode(&ff.space.BCClient)
				if err != nil {
					f.ShowError(err)
					return
				}
				d := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
					if err != nil {
						f.ShowError(err)
						return
					}
					if reader == nil {
						return
					}

					// Show progress dialog
					progress := dialog.NewProgress("Uploading", "Uploading "+reader.Name(), f.Window)
					progress.Show()
					defer progress.Hide()
					listener := &bcui.ProgressMiningListener{Func: progress.SetValue}

					reference, err := ff.space.Add(node, listener, reader.Name(), reader.URI().MimeType(), reader)
					if err != nil {
						f.ShowError(err)
					}
					fmt.Println("Uploaded:", reference)
					go populateSpace(ff)
				}, f.Window)
				d.SetFilter(storage.NewMimeTypeFileFilter([]string{"image/*"}))
				d.Show()
			}()
		}),
		widget.NewToolbarAction(theme.ViewRefreshIcon(), func() {
			go populateSpace(ff)
		}),
		widget.NewToolbarAction(theme.SearchIcon(), func() {
			f := spacefynego.NewSpaceFyne(ff.app, ff.window, ff.space)
			go f.SearchFile(ff.space)
		}),
		widget.NewToolbarSpacer(),
		widget.NewToolbarAction(theme.NewThemedResource(spaceuidata.StorageIcon, nil), func() {
			f := spacefynego.NewSpaceFyne(ff.app, ff.window, ff.space)
			go f.ShowStorage(ff.space)
		}),
		widget.NewToolbarAction(bcuidata.NewPrimaryThemedResource(bcuidata.AccountIcon), func() {
			f := spacefynego.NewSpaceFyne(ff.app, ff.window, ff.space)
			go f.ShowAccount(&ff.space.BCClient)
		}),
		widget.NewToolbarAction(theme.HelpIcon(), func() {
			f := spacefynego.NewSpaceFyne(ff.app, ff.window, ff.space)
			go f.ShowHelp(ff.space)
		}),
		widget.NewToolbarAction(theme.InfoIcon(), func() {
			showAbout(ff)
		}),
	)
	// Create list of thumbnails
	ff.spaceImages = ui.NewSpaceThumbnailTable(ff.space, func(id string, timestamp uint64, meta *spacego.Meta) {
		// Create Space Fyne
		f := spacefynego.NewSpaceFyne(ff.app, ff.window, ff.space)
		node, err := f.GetNode(&ff.space.BCClient)
		if err != nil {
			f.ShowError(err)
			return
		}

		c, err := filesystem.ImageCache()
		if err != nil {
			f.ShowError(err)
			return
		}

		file := filepath.Join(c, id)
		if _, err := os.Stat(file); os.IsNotExist(err) {
			hash, err := base64.RawURLEncoding.DecodeString(id)
			if err != nil {
				f.ShowError(err)
				return
			}
			out, err := os.Create(file)
			if err != nil {
				f.ShowError(err)
				return
			}
			// TODO display and update progress bar
			count, err := ff.space.Read(node, hash, out)
			if err != nil {
				f.ShowError(err)
				return
			}
			fmt.Println("Wrote", bcgo.BinarySizeToString(count), "to", file)
		}
		hideBrowser(ff)
		showViewer(ff, meta.Name, storage.NewFileURI(file))
	})

	ff.bSources = container.NewAppTabs(
		widget.NewTabItem("Local", container.NewBorder(ff.localToolbar, nil, ff.localDirs, nil, ff.localImages)),
		widget.NewTabItem(spacego.SPACE, container.NewBorder(ff.spaceToolbar, nil, nil, nil, ff.spaceImages)),
	)
	ff.bSources.OnChanged = func(tab *widget.TabItem) {
		switch tab.Text {
		case "Local":
			ff.window.SetTitle("FyFoto - " + ff.currentDir.String())
			go populateLocal(ff, ff.currentDir)
		case spacego.SPACE:
			ff.window.SetTitle("FyFoto - " + spacego.SPACE)
			go populateSpace(ff)
		}
	}
	ff.bInfo = widget.NewLabelWithStyle("No Images", fyne.TextAlignCenter, fyne.TextStyle{})

	ff.browser = container.NewBorder(nil, ff.bInfo, nil, nil, ff.bSources)
}
