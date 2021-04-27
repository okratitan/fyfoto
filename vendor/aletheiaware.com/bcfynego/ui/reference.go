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
	"aletheiaware.com/bcfynego/storage"
	"aletheiaware.com/bcgo"
	"encoding/base64"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"strconv"
)

type ReferenceView struct {
	widget.Form
	ui        UI
	client    bcclientgo.BCClient
	timestamp *TimestampLabel
	channel   *Link
	block     *Link
	record    *Link
	index     *widget.Label
}

func NewReferenceView(ui UI, client bcclientgo.BCClient) *ReferenceView {
	v := &ReferenceView{
		ui:        ui,
		client:    client,
		timestamp: NewTimestampLabel(0),
		channel: &Link{
			Hyperlink: widget.Hyperlink{
				TextStyle: fyne.TextStyle{
					Monospace: true,
				},
				Wrapping: fyne.TextWrapBreak,
			},
		},
		block: &Link{
			Hyperlink: widget.Hyperlink{
				TextStyle: fyne.TextStyle{
					Monospace: true,
				},
				Wrapping: fyne.TextWrapBreak,
			},
		},
		record: &Link{
			Hyperlink: widget.Hyperlink{
				TextStyle: fyne.TextStyle{
					Monospace: true,
				},
				Wrapping: fyne.TextWrapBreak,
			},
		},
		index: &widget.Label{
			TextStyle: fyne.TextStyle{
				Monospace: true,
			},
			Wrapping: fyne.TextWrapBreak,
		},
	}
	v.ExtendBaseWidget(v)
	v.timestamp.ExtendBaseWidget(v.timestamp)
	v.channel.ExtendBaseWidget(v.channel)
	v.block.ExtendBaseWidget(v.block)
	v.record.ExtendBaseWidget(v.record)
	v.index.ExtendBaseWidget(v.index)
	v.Append("Timestamp", v.timestamp)
	v.Append("Channel", v.channel)
	v.Append("Block", v.block)
	v.Append("Record", v.record)
	v.Append("Index", v.index)
	return v
}

func (v *ReferenceView) SetReference(reference *bcgo.Reference) {
	v.timestamp.SetTimestamp(reference.Timestamp)
	v.channel.SetText(reference.ChannelName)
	v.channel.OnTapped = func() {
		v.ui.ShowURI(v.client, storage.NewChannelURI(reference.ChannelName))
	}
	v.block.SetText(base64.RawURLEncoding.EncodeToString(reference.BlockHash))
	v.block.OnTapped = func() {
		v.ui.ShowURI(v.client, storage.NewBlockURI(reference.ChannelName, reference.BlockHash))
	}
	v.record.SetText(base64.RawURLEncoding.EncodeToString(reference.RecordHash))
	v.record.OnTapped = func() {
		v.ui.ShowURI(v.client, storage.NewRecordURI(reference.ChannelName, reference.BlockHash, reference.RecordHash))
	}
	v.index.SetText(strconv.FormatUint(reference.Index, 10))
	v.Refresh()
}
