/*
 * Copyright 2020 Aletheia Ware LLC
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package ui

import (
	"aletheiaware.com/bcclientgo"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

var (
	DialogSize = fyne.NewSize(400, 300)
	WindowSize = fyne.NewSize(800, 600)
)

type UI interface {
	ShowError(error)
	ShowURI(*bcclientgo.BCClient, fyne.URI)
}

func ShortcutFocused(s fyne.Shortcut, w fyne.Window) {
	if focused, ok := w.Canvas().Focused().(fyne.Shortcutable); ok {
		focused.TypedShortcut(s)
	}
}

type Link struct {
	widget.Hyperlink
	OnTapped func()
}

func (l *Link) Tapped(*fyne.PointEvent) {
	if f := l.OnTapped; f != nil {
		f()
	}
}
