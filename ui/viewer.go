package ui

import (
	"image/color"
	"strings"

	"fyne.io/fyne"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/widget"
)

var _ fyne.Widget = (*Viewer)(nil)

type Viewer struct {
	widget.BaseWidget
	URI fyne.URI
}

func NewViewer() *Viewer {
	v := &Viewer{}
	v.ExtendBaseWidget(v)
	return v
}

func (v *Viewer) MinSize() fyne.Size {
	v.ExtendBaseWidget(v)
	return v.BaseWidget.MinSize()
}

func (v *Viewer) CreateRenderer() fyne.WidgetRenderer {
	v.ExtendBaseWidget(v)
	image := &canvas.Image{FillMode: canvas.ImageFillContain}
	info := widget.NewLabelWithStyle("Image Info Placeholder", fyne.TextAlignCenter, fyne.TextStyle{})
	return &viewerRenderer{
		Viewer: v,
		Content: fyne.NewContainerWithLayout(
			layout.NewBorderLayout(nil, info, nil, nil),
			info,
			image),
		Image: image,
		Info:  info,
	}
}

func (v *Viewer) SetURI(uri fyne.URI) {
	v.URI = uri
	v.Refresh()
}

var _ fyne.WidgetRenderer = (*viewerRenderer)(nil)

type viewerRenderer struct {
	Viewer  *Viewer
	Content *fyne.Container
	Image   *canvas.Image
	Info    *widget.Label
}

func (r *viewerRenderer) BackgroundColor() color.Color {
	return color.Black // TODO color should depend on current theme
}

func (r *viewerRenderer) Destroy() {
}

func (r *viewerRenderer) Layout(size fyne.Size) {
	r.Content.Resize(size)
}

func (r *viewerRenderer) MinSize() fyne.Size {
	return r.Content.MinSize()
}

func (r *viewerRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.Content}
}

func (r *viewerRenderer) Refresh() {
	uri := r.Viewer.URI
	if uri != nil {
		r.Info.SetText(uri.Name())
		r.Image.File = strings.TrimPrefix(uri.String(), "file://")
	}
	r.Image.FillMode = canvas.ImageFillContain
	r.Image.Refresh()
}
