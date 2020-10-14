package main

import (
	"image/color"

	"fyne.io/fyne"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/dialog"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"
)

func hideViewer(ff *FyFoto) {
	ff.vImage.Hide()
	ff.viewer.Hide()
	ff.vContent.Hide()
	ff.vToolbar.Hide()
	ff.vInfo.Hide()
	ff.vRect.Hide()
}

func showViewer(ff *FyFoto, imagePath string) {
	ff.vImage.File = imagePath
	ff.vImage.FillMode = canvas.ImageFillContain

	ff.vImage.Show()
	ff.viewer.Show()
	ff.vContent.Show()
	ff.vToolbar.Show()
	ff.vInfo.Show()
	ff.vRect.Show()
	ff.window.SetTitle("Fyfoto - " + imagePath)
	canvas.Refresh(ff.main)
}

func createViewer(ff *FyFoto) {
	ff.vImage = &canvas.Image{FillMode: canvas.ImageFillContain}
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

	ff.vInfo = widget.NewLabelWithStyle("Image Info Placeholder", fyne.TextAlignCenter, fyne.TextStyle{})

	ff.vRect = canvas.NewRectangle(color.Black)
	ff.vContent = fyne.NewContainerWithLayout(layout.NewBorderLayout(ff.vToolbar, ff.vInfo, nil, nil), ff.vToolbar, ff.vInfo, ff.vImage)
	ff.viewer = fyne.NewContainerWithLayout(layout.NewBorderLayout(nil, nil, nil, nil), ff.vRect, ff.vContent)
}
