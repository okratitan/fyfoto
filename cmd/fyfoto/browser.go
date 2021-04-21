package main

import (
	bcui "aletheiaware.com/bcfynego/ui"
	bcuidata "aletheiaware.com/bcfynego/ui/data"
	"aletheiaware.com/bcgo"
	spaceuidata "aletheiaware.com/spacefynego/ui/data"
	"aletheiaware.com/spacego"
	"encoding/base64"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/okratitan/fyfoto/internal/filesystem"
	"github.com/okratitan/fyfoto/ui"
	"io"
	"os"
	"path/filepath"
)

func fileIsImage(file fyne.URI) bool {
	mimes := &storage.MimeTypeFileFilter{MimeTypes: []string{"image/*"}}
	return mimes.Matches(file)
}

func populateLocal(ff *FyFoto, dir fyne.URI) {
	// TODO Should this show a Progress Bar?

	/* TODO this API was removed in Fyne v2.0.0
	// Update Status
	if widget.Renderer(ff.bInfo) != nil {
		ff.bInfo.SetText("Loading Images")
	}
	*/

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
	// Get BC Node
	node, err := ff.spaceFyne.Node(ff.spaceClient)
	if err != nil {
		ff.spaceFyne.ShowError(err)
		return
	}
	populateSpaceWithNode(ff, node)
}

func populateSpaceWithNode(ff *FyFoto, node bcgo.Node) {
	// TODO Should this show a Progress Bar?

	/* TODO this API was removed in Fyne v2.0.0
	// Update Status
	if widget.Renderer(ff.bInfo) != nil {
		ff.bInfo.SetText("Loading Images")
	}
	*/

	// Update Table
	ff.spaceImages.Update(node)

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
	dialog.ShowInformation("About", "FyFoto - A Cross-Platform Image Application\n\nS P A C E - A Secure, Private, Storage Platform ", ff.window)
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
			go func() {
				node, err := ff.spaceFyne.Node(ff.spaceClient)
				if err != nil {
					ff.spaceFyne.ShowError(err)
					return
				}
				d := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
					if err != nil {
						ff.spaceFyne.ShowError(err)
						return
					}
					if reader == nil {
						return
					}
					name := reader.URI().Name()

					// Show progress dialog
					progress := dialog.NewProgress("Uploading", "Uploading "+name, ff.spaceFyne.Window())
					progress.Show()
					listener := &bcui.ProgressMiningListener{Func: progress.SetValue}

					reference, err := ff.spaceClient.Add(node, listener, name, reader.URI().MimeType(), reader)

					// Hide progress dialog
					progress.Hide()

					if err != nil {
						ff.spaceFyne.ShowError(err)
						return
					}
					fmt.Println("Uploaded:", reference)
					go populateSpaceWithNode(ff, node)
				}, ff.spaceFyne.Window())
				d.SetFilter(storage.NewMimeTypeFileFilter([]string{"image/*"}))
				d.Show()
			}()
		}),
		widget.NewToolbarAction(theme.ViewRefreshIcon(), func() {
			go populateSpace(ff)
		}),
		widget.NewToolbarAction(theme.SearchIcon(), func() {
			go ff.spaceFyne.SearchFile(ff.spaceClient)
		}),
		widget.NewToolbarSpacer(),
		widget.NewToolbarAction(theme.NewThemedResource(spaceuidata.StorageIcon), func() {
			go ff.spaceFyne.ShowStorage(ff.spaceClient)
		}),
		widget.NewToolbarAction(theme.NewThemedResource(bcuidata.AccountIcon), func() {
			go ff.spaceFyne.ShowAccount(ff.spaceClient)
		}),
		widget.NewToolbarAction(theme.HelpIcon(), func() {
			go ff.spaceFyne.ShowHelp(ff.spaceClient)
		}),
		widget.NewToolbarAction(theme.InfoIcon(), func() {
			showAbout(ff)
		}),
	)
	// Create list of thumbnails
	ff.spaceImages = ui.NewSpaceThumbnailTable(ff.spaceClient, func(id string, timestamp uint64, meta *spacego.Meta) {
		node, err := ff.spaceFyne.Node(ff.spaceClient)
		if err != nil {
			ff.spaceFyne.ShowError(err)
			return
		}

		c, err := filesystem.ImageCache()
		if err != nil {
			ff.spaceFyne.ShowError(err)
			return
		}

		file := filepath.Join(c, id)
		if _, err := os.Stat(file); os.IsNotExist(err) {
			progress := dialog.NewProgressInfinite("Downloading", "Downloading "+meta.Name, ff.spaceFyne.Window())
			progress.Show()
			defer progress.Hide()

			hash, err := base64.RawURLEncoding.DecodeString(id)
			if err != nil {
				ff.spaceFyne.ShowError(err)
				return
			}
			out, err := os.Create(file)
			if err != nil {
				ff.spaceFyne.ShowError(err)
				return
			}
			reader, err := ff.spaceClient.ReadFile(node, hash)
			if err != nil {
				ff.spaceFyne.ShowError(err)
				return
			}
			count, err := io.Copy(out, reader)
			if err != nil {
				ff.spaceFyne.ShowError(err)
				return
			}
			fmt.Println("Wrote", bcgo.BinarySizeToString(uint64(count)), "to", file)
		}
		hideBrowser(ff)
		showViewer(ff, meta.Name, storage.NewFileURI(file))
	})

	ff.bSources = container.NewAppTabs(
		container.NewTabItem("Local", container.NewBorder(ff.localToolbar, nil, ff.localDirs, nil, ff.localImages)),
		container.NewTabItem(spacego.SPACE, container.NewBorder(ff.spaceToolbar, nil, nil, nil, ff.spaceImages)),
	)
	ff.bSources.OnChanged = func(tab *container.TabItem) {
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
