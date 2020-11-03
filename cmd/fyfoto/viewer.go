package main

import (
	"fyne.io/fyne"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/container"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"
	"github.com/okratitan/fyfoto/ui"
)

func hideViewer(ff *FyFoto) {
	ff.viewer.Hide()
}

func showViewer(ff *FyFoto, image fyne.URI) {
	ff.vWidget.SetURI(image)
	ff.viewer.Show()
	ff.window.SetTitle("Fyfoto - " + image.String())
	canvas.Refresh(ff.main)
}

func createViewer(ff *FyFoto) {
	ff.vWidget = ui.NewViewer()
	ff.vToolbar = widget.NewToolbar(widget.NewToolbarAction(theme.NavigateBackIcon(),
		func() {
			showBrowser(ff, ff.currentDir)
			hideViewer(ff)
		}),
		widget.NewToolbarSpacer(),
		widget.NewToolbarAction(theme.InfoIcon(), func() {
			showAbout(ff)
		}),
	)

	ff.viewer = container.NewBorder(ff.vToolbar, nil, nil, nil, ff.vToolbar, ff.vWidget)
}
