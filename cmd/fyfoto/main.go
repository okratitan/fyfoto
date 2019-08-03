package main

import (
	"flag"
	"fmt"
	"os/user"

	"fyne.io/fyne"
	"fyne.io/fyne/app"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/dialog"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"
)

//FyFoto Globals
type FyFoto struct {
	app    fyne.App
	window fyne.Window

	//Browser
	browser   *fyne.Container
	dirs      *fyne.Container
	dscroller *widget.ScrollContainer
	images    *fyne.Container
	iscroller *widget.ScrollContainer

	//Image Viewer
	viewer      *fyne.Container
	viewerImage *canvas.Image

	//Main Layout
	main    *fyne.Container
	content *fyne.Container
	info    *widget.Label
	toolbar *widget.Toolbar

	currentDir string
	dirsHidden int
}

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

func main() {
	usr, err := user.Current()
	if err != nil {
		fmt.Println("Could not find the current user")
	}

	dirPtr := flag.String("path", usr.HomeDir, "Path to a directory")
	flag.Parse()

	ff := &FyFoto{app: app.New()}
	ff.window = ff.app.NewWindow("FyFoto")

	ff.toolbar = widget.NewToolbar(widget.NewToolbarAction(theme.NavigateBackIcon(),
		func() {
			showBrowser(ff, ff.currentDir)
			hideViewer(ff)
		}),
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

	ff.info = widget.NewLabelWithStyle("No Images", fyne.TextAlignCenter, fyne.TextStyle{})

	createBrowser(ff)
	createViewer(ff)
	showBrowser(ff, *dirPtr)
	hideViewer(ff)

	ff.content = fyne.NewContainerWithLayout(layout.NewBorderLayout(nil, nil, nil, nil), ff.browser, ff.viewer)

	ff.main = fyne.NewContainerWithLayout(layout.NewBorderLayout(ff.toolbar, ff.info, nil, nil),
		ff.toolbar, ff.info, ff.content)

	ff.window.SetContent(ff.main)
	ff.window.Resize(fyne.NewSize(1024, 576))

	ff.window.ShowAndRun()
}
