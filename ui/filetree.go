package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"strings"
)

func NewFileTree(root fyne.URI) *widget.Tree {
	tree := &widget.Tree{
		Root: root.String(),
		IsBranch: func(uid string) bool {
			_, err := storage.ListerForURI(storage.NewURI(uid))
			return err == nil
		},
		CreateNode: func(branch bool) fyne.CanvasObject {
			var icon fyne.CanvasObject
			if branch {
				icon = widget.NewIcon(nil)
			} else {
				icon = widget.NewFileIcon(nil)
			}
			return fyne.NewContainerWithLayout(layout.NewHBoxLayout(), icon, widget.NewLabel("Template Object"))
		},
	}
	tree.ChildUIDs = func(uid string) (c []string) {
		luri, err := storage.ListerForURI(storage.NewURI(uid))
		if err != nil {
			fyne.LogError("Unable to get lister for "+uid, err)
		} else {
			uris, err := luri.List()
			if err != nil {
				return
			} else {
				// Filter URIs
				var us []fyne.URI
				for _, u := range uris {
					_, err := storage.ListerForURI(u)
					if err == nil && !strings.HasPrefix(u.Name(), ".") {
						us = append(us, u)
					}
				}
				// Convert to Strings
				for _, u := range us {
					c = append(c, u.String())
				}
			}
		}
		return
	}
	tree.UpdateNode = func(uid string, branch bool, node fyne.CanvasObject) {
		uri := storage.NewURI(uid)
		c := node.(*fyne.Container)
		if branch {
			var r fyne.Resource
			if tree.IsBranchOpen(uid) {
				// Set open folder icon
				r = theme.FolderOpenIcon()
			} else {
				// Set folder icon
				r = theme.FolderIcon()
			}
			c.Objects[0].(*widget.Icon).SetResource(r)
		} else {
			// Set file uri to update icon
			c.Objects[0].(*widget.FileIcon).SetURI(uri)
		}
		l := c.Objects[1].(*widget.Label)
		if tree.Root == uid {
			l.SetText(uid)
		} else {
			l.SetText(uri.Name())
		}
	}
	tree.ExtendBaseWidget(tree)
	return tree
}
