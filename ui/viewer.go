package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"image/color"
	"strings"
)

var _ fyne.Widget = (*Viewer)(nil)

type Viewer struct {
	widget.BaseWidget
	Name string
	URI  fyne.URI
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

func (v *Viewer) SetSource(name string, uri fyne.URI) {
	v.Name = name
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
	r.Info.SetText(r.Viewer.Name)
	if uri := r.Viewer.URI; uri != nil {
		r.Image.File = strings.TrimPrefix(uri.String(), "file://")
	}
	r.Image.FillMode = canvas.ImageFillContain
	r.Image.Refresh()
}
