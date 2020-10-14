package main

import (
	"flag"
	"fmt"
	"os/user"

	"fyne.io/fyne"
	"fyne.io/fyne/app"
	"fyne.io/fyne/canvas"
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

	dirBox *fyne.Container
	up     *widget.Button
	dirs   *widget.List

	images    *fyne.Container
	iScroller *widget.ScrollContainer

	directories []string

	//Image Viewer
	viewer   *fyne.Container
	vContent *fyne.Container
	vImage   *canvas.Image
	vInfo    *widget.Label
	vRect    *canvas.Rectangle
	vToolbar *widget.Toolbar

	//Main Layout
	main *fyne.Container

	currentDir string
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
	ff.window = ff.app.NewWindow("FyFoto")

	createBrowser(ff)
	createViewer(ff)
	showBrowser(ff, *dirPtr)
	hideViewer(ff)

	ff.main = fyne.NewContainerWithLayout(layout.NewBorderLayout(nil, nil, nil, nil), ff.browser, ff.viewer)

	ff.window.SetContent(ff.main)
	ff.window.Resize(fyne.NewSize(1024, 576))

	ff.window.ShowAndRun()
}
