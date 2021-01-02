package main

import (
	"aletheiaware.com/spaceclientgo"
	"flag"
	"fyne.io/fyne"
	"fyne.io/fyne/app"
	"fyne.io/fyne/container"
	"fyne.io/fyne/storage"
	"fyne.io/fyne/widget"
	"github.com/okratitan/fyfoto/internal/filesystem"
	"github.com/okratitan/fyfoto/ui"
	"log"
)

//FyFoto Globals
type FyFoto struct {
	app    fyne.App
	window fyne.Window

	//Browser
	browser  *fyne.Container
	bSources *container.AppTabs
	bInfo    *widget.Label

	localToolbar *widget.Toolbar
	localDirs    *widget.Tree
	localImages  *ui.LocalThumbnailTable

	space        *spaceclientgo.SpaceClient
	spaceToolbar *widget.Toolbar
	spaceImages  *ui.SpaceThumbnailTable

	//Image Viewer
	viewer   *fyne.Container
	vWidget  *ui.Viewer
	vToolbar *widget.Toolbar

	//Main Layout
	main *fyne.Container

	rootDir    fyne.URI
	currentDir fyne.URI
	dirsHidden int
}

func main() {
	rootDir, err := filesystem.RootDirectory()
	if err != nil {
		log.Fatal(err)
	}

	dirPtr := flag.String("path", rootDir, "Path to a directory")
	flag.Parse()

	ff := &FyFoto{
		app:   app.New(),
		space: spaceclientgo.NewSpaceClient(),
	}
	ff.rootDir = storage.NewFileURI(*dirPtr)
	ff.window = ff.app.NewWindow("FyFoto")

	createBrowser(ff)
	createViewer(ff)

	ff.main = container.NewMax(ff.browser, ff.viewer)

	showBrowser(ff, ff.rootDir)
	hideViewer(ff)

	ff.window.SetContent(ff.main)

	ff.window.Resize(fyne.NewSize(1024, 576))
	ff.window.CenterOnScreen()
	ff.window.ShowAndRun()
}
