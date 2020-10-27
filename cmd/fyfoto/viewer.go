package main

import (
	"github.com/okratitan/fyfoto/ui"

	"fyne.io/fyne"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/dialog"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"
)

func hideViewer(ff *FyFoto) {
	ff.vWidget.Hide()
	ff.viewer.Hide()
	ff.vToolbar.Hide()
}

func showViewer(ff *FyFoto, image fyne.URI) {
	ff.vWidget.SetURI(image)
	ff.vWidget.Show()
	ff.viewer.Show()
	ff.vToolbar.Show()
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
		widget.NewToolbarAction(theme.SettingsIcon(),
			func() {
				dialog.ShowInformation("About", "FyFoto - A Cross-Platform Image Application", ff.window)
			}))

	ff.viewer = fyne.NewContainerWithLayout(layout.NewBorderLayout(ff.vToolbar, nil, nil, nil), ff.vToolbar, ff.vWidget)
}
