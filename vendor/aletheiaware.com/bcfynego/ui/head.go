/*
 * Copyright 2021 Aletheia Ware LLC
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
	"aletheiaware.com/bcfynego/storage"
	"aletheiaware.com/bcgo"
	"encoding/base64"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"log"
)

type HeadView struct {
	widget.Form
	ui        UI
	client    *bcclientgo.BCClient
	channel   *widget.Label
	hash      *Link
	timestamp *widget.Label
}

func NewHeadView(ui UI, client *bcclientgo.BCClient) *HeadView {
	v := &HeadView{
		ui:     ui,
		client: client,
		channel: &widget.Label{
			TextStyle: fyne.TextStyle{
				Monospace: true,
			},
			Wrapping: fyne.TextWrapBreak,
		},
		hash: &Link{
			Hyperlink: widget.Hyperlink{
				TextStyle: fyne.TextStyle{
					Monospace: true,
				},
				Wrapping: fyne.TextWrapBreak,
			},
		},
		timestamp: &widget.Label{
			TextStyle: fyne.TextStyle{
				Monospace: true,
			},
			Wrapping: fyne.TextWrapBreak,
		},
	}
	v.ExtendBaseWidget(v)
	v.channel.ExtendBaseWidget(v.channel)
	v.hash.ExtendBaseWidget(v.hash)
	v.timestamp.ExtendBaseWidget(v.timestamp)
	v.Append("Channel", v.channel)
	v.Append("Hash", v.hash)
	v.Append("Timestamp", v.timestamp)
	v.Append("", container.NewGridWithColumns(2,
		widget.NewButton("Pull", func() {
			cache, err := v.client.GetCache()
			if err != nil {
				v.ui.ShowError(err)
				return
			}
			network, err := v.client.GetNetwork()
			if err != nil {
				v.ui.ShowError(err)
				return
			}
			channel := bcgo.NewChannel(v.channel.Text)
			if err := channel.Pull(cache, network); err != nil {
				v.ui.ShowError(err)
				return
			}
		}),
		widget.NewButton("Push", func() {
			cache, err := v.client.GetCache()
			if err != nil {
				v.ui.ShowError(err)
				return
			}
			network, err := v.client.GetNetwork()
			if err != nil {
				v.ui.ShowError(err)
				return
			}
			channel := bcgo.NewChannel(v.channel.Text)
			if err := channel.Push(cache, network); err != nil {
				v.ui.ShowError(err)
				return
			}
		}),
		widget.NewButton("Write", func() {
			log.Println("// TODO go c.Write()")
		}),
		widget.NewButton("Mine", func() {
			log.Println("// TODO go c.Mine()")
		}),
	))
	return v
}

func (v *HeadView) SetURI(uri storage.ChannelURI) error {
	cache, err := v.client.GetCache()
	if err != nil {
		return err
	}
	network, err := v.client.GetNetwork()
	if err != nil {
		return err
	}

	name := uri.Channel()
	channel := bcgo.NewChannel(name)
	if err := channel.Refresh(cache, network); err != nil {
		// Ignored
	}
	v.hash.SetText(base64.RawURLEncoding.EncodeToString(channel.Head))
	v.hash.OnTapped = func() {
		v.ui.ShowURI(v.client, storage.NewBlockURI(name, channel.Head))
	}
	v.channel.SetText(name)
	v.timestamp.SetText(bcgo.TimestampToString(channel.Timestamp))
	v.Refresh()
	return nil
}
