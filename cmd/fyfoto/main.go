package main

import (
	"flag"
	"fmt"

	"fyne.io/fyne/storage"

	"github.com/okratitan/fyfoto/ui"
	"os/user"

	"fyne.io/fyne"
	"fyne.io/fyne/app"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/widget"
)

//FyFoto Globals
type FyFoto struct {
	app    fyne.App
	window fyne.Window

	//Browser
	browser  *fyne.Container
	bToolbar *widget.Toolbar
	bInfo    *widget.Label

	dirs *widget.Tree

	images    *fyne.Container
	iScroller *widget.ScrollContainer

	directories []string

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

	ff := &FyFoto{app: app.New()}
	ff.rootDir = storage.NewURI("file://" + *dirPtr)
	ff.window = ff.app.NewWindow("FyFoto")

	createBrowser(ff)
	createViewer(ff)
	showBrowser(ff, ff.rootDir)
	hideViewer(ff)

	ff.main = fyne.NewContainerWithLayout(layout.NewBorderLayout(nil, nil, nil, nil), ff.browser, ff.viewer)

	ff.window.SetContent(ff.main)
	ff.window.Resize(fyne.NewSize(1024, 576))

	ff.window.ShowAndRun()
}
