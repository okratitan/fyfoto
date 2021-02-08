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
	"aletheiaware.com/aliasgo"
	"aletheiaware.com/bcclientgo"
	"aletheiaware.com/bcfynego/storage"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

type AliasLabel struct {
	widget.Label
}

func NewAliasLabel(alias string) *AliasLabel {
	a := &AliasLabel{
		Label: widget.Label{
			Alignment: fyne.TextAlignLeading,
			TextStyle: fyne.TextStyle{Monospace: true},
			Wrapping:  fyne.TextWrapBreak,
		},
	}
	a.ExtendBaseWidget(a)
	a.SetAlias(alias)
	return a
}

func (a *AliasLabel) SetAlias(alias string) {
	a.SetText(alias)
}

type AliasView struct {
	widget.Form
	ui        UI
	client    *bcclientgo.BCClient
	timestamp *TimestampLabel
	alias     *AliasLabel
	key       *KeyLabel
}

func NewAliasView(ui UI, client *bcclientgo.BCClient) *AliasView {
	v := &AliasView{
		ui:        ui,
		client:    client,
		timestamp: NewTimestampLabel(0),
		alias:     NewAliasLabel(""),
		key:       NewKeyLabel(nil),
	}
	v.ExtendBaseWidget(v)
	v.alias.ExtendBaseWidget(v.alias)
	v.key.ExtendBaseWidget(v.key)
	v.timestamp.ExtendBaseWidget(v.timestamp)
	v.Append("Timestamp", v.timestamp)
	v.Append("Alias", v.alias)
	v.Append("Key", v.key)
	return v
}

func (v *AliasView) SetURI(uri storage.AliasURI) error {
	cache, err := v.client.GetCache()
	if err != nil {
		return err
	}
	network, err := v.client.GetNetwork()
	if err != nil {
		return err
	}
	aliases := aliasgo.OpenAliasChannel()
	if err := aliases.Refresh(cache, network); err != nil {
		// Ignored
	}
	alias := uri.Alias()
	r, a, err := aliasgo.GetRecord(aliases, cache, network, alias)
	if err != nil {
		return err
	}
	v.alias.SetText(alias)
	v.key.SetKey(a.PublicKey)
	v.timestamp.SetTimestamp(r.Timestamp)
	v.Refresh()
	return nil
}
