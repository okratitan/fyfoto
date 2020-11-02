package main

import (
	"flag"
	"fmt"
	"fyne.io/fyne"
	"fyne.io/fyne/app"
	"fyne.io/fyne/container"
	"fyne.io/fyne/storage"
	"fyne.io/fyne/widget"
	"github.com/AletheiaWareLLC/spaceclientgo"
	"github.com/okratitan/fyfoto/ui"
	"os/user"
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
	usr, err := user.Current()
	if err != nil {
		fmt.Println("Could not find the current user")
	}

	dirPtr := flag.String("path", usr.HomeDir, "Path to a directory")
	flag.Parse()

	ff := &FyFoto{
		app:   app.New(),
		space: spaceclientgo.NewSpaceClient(),
	}
	ff.rootDir = storage.NewURI("file://" + *dirPtr)
	ff.window = ff.app.NewWindow("FyFoto")

	createBrowser(ff)
	createViewer(ff)
	showBrowser(ff, ff.rootDir)
	hideViewer(ff)

	ff.main = container.NewMax(ff.browser, ff.viewer)

	ff.window.SetContent(ff.main)
	ff.window.Resize(fyne.NewSize(1024, 576))
	ff.window.CenterOnScreen()
	ff.window.ShowAndRun()
}
