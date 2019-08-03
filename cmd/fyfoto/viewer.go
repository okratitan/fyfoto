package main

import (
	"fyne.io/fyne"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/layout"
)

func hideViewer(ff *FyFoto) {
	ff.viewerImage.Hide()
	ff.viewer.Hide()
}

func showViewer(ff *FyFoto, imagePath string) {
	ff.viewerImage.File = imagePath
	ff.viewerImage.FillMode = canvas.ImageFillContain

	ff.viewerImage.Show()
	ff.viewer.Show()
	ff.window.SetTitle("Fyfoto - " + imagePath)
	canvas.Refresh(ff.main)
}

func createViewer(ff *FyFoto) {
	ff.viewerImage = &canvas.Image{FillMode: canvas.ImageFillContain}

	ff.viewer = fyne.NewContainerWithLayout(layout.NewGridLayout(1), ff.viewerImage)
}
