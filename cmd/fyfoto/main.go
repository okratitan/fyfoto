package main

import (
	"aletheiaware.com/bcgo"
	"aletheiaware.com/spaceclientgo"
	"aletheiaware.com/spacefynego"
	"flag"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
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

	spaceClient  *spaceclientgo.SpaceClient
	spaceFyne    *spacefynego.SpaceFyne
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
		app:         app.New(),
		spaceClient: spaceclientgo.NewSpaceClient(),
	}
	ff.rootDir = storage.NewFileURI(*dirPtr)
	ff.window = ff.app.NewWindow("FyFoto")
	ff.spaceFyne = spacefynego.NewSpaceFyne(ff.app, ff.window, ff.spaceClient)
	onSignedIn := ff.spaceFyne.OnSignedIn
	ff.spaceFyne.OnSignedIn = func(node *bcgo.Node) {
		if onSignedIn != nil {
			onSignedIn(node)
		}
		go populateSpaceWithNode(ff, node)
	}
	onSignedOut := ff.spaceFyne.OnSignedOut
	ff.spaceFyne.OnSignedOut = func() {
		if onSignedOut != nil {
			onSignedOut()
		}
		ff.spaceImages.Clear()
		ff.bInfo.SetText("")
	}

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
