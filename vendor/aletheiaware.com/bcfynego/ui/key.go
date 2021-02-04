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
	"encoding/base64"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

type KeyLabel struct {
	widget.Label
}

func NewKeyLabel(key []byte) *KeyLabel {
	k := &KeyLabel{
		widget.Label{
			Alignment: fyne.TextAlignLeading,
			TextStyle: fyne.TextStyle{Monospace: true},
			Wrapping:  fyne.TextWrapBreak,
		},
	}
	k.ExtendBaseWidget(k)
	k.SetKey(key)
	return k
}

func (k *KeyLabel) SetKey(key []byte) {
	k.SetText(base64.RawURLEncoding.EncodeToString(key))
}
