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
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type BlockView struct {
	widget.Form
	ui        UI
	client    *bcclientgo.BCClient
	hash      *widget.Label
	timestamp *TimestampLabel
	channel   *Link
	length    *widget.Label
	previous  *Link
	miner     *Link
	nonce     *widget.Label
	record    *fyne.Container
}

func NewBlockView(ui UI, client *bcclientgo.BCClient) *BlockView {
	v := &BlockView{
		ui:     ui,
		client: client,
		hash: &widget.Label{
			TextStyle: fyne.TextStyle{
				Monospace: true,
			},
			Wrapping: fyne.TextWrapBreak,
		},
		timestamp: NewTimestampLabel(0),
		channel: &Link{
			Hyperlink: widget.Hyperlink{
				TextStyle: fyne.TextStyle{
					Monospace: true,
				},
				Wrapping: fyne.TextWrapBreak,
			},
		},
		length: &widget.Label{
			TextStyle: fyne.TextStyle{
				Monospace: true,
			},
			Wrapping: fyne.TextWrapBreak,
		},
		previous: &Link{
			Hyperlink: widget.Hyperlink{
				TextStyle: fyne.TextStyle{
					Monospace: true,
				},
				Wrapping: fyne.TextWrapBreak,
			},
		},
		miner: &Link{
			Hyperlink: widget.Hyperlink{
				TextStyle: fyne.TextStyle{
					Monospace: true,
				},
				Wrapping: fyne.TextWrapBreak,
			},
		},
		nonce: &widget.Label{
			TextStyle: fyne.TextStyle{
				Monospace: true,
			},
			Wrapping: fyne.TextWrapBreak,
		},
		record: container.NewVBox(),
	}
	v.ExtendBaseWidget(v)
	v.hash.ExtendBaseWidget(v.hash)
	v.timestamp.ExtendBaseWidget(v.timestamp)
	v.channel.ExtendBaseWidget(v.channel)
	v.length.ExtendBaseWidget(v.length)
	v.previous.ExtendBaseWidget(v.previous)
	v.miner.ExtendBaseWidget(v.miner)
	v.nonce.ExtendBaseWidget(v.nonce)
	v.Append("Hash", v.hash)
	v.Append("Timestamp", v.timestamp)
	v.Append("Channel", v.channel)
	v.Append("Length", v.length)
	v.Append("Previous", v.previous)
	v.Append("Miner", v.miner)
	v.Append("Nonce", v.nonce)
	v.Append("Records", v.record)
	return v
}

func (v *BlockView) SetURI(uri storage.BlockURI) error {
	cache, err := v.client.GetCache()
	if err != nil {
		return err
	}
	network, err := v.client.GetNetwork()
	if err != nil {
		return err
	}
	name := uri.Channel()
	hash := uri.BlockHash()
	block, err := bcgo.GetBlock(name, cache, network, hash)
	if err != nil {
		return err
	}
	v.SetHash(hash)
	v.SetBlock(block)
	return nil
}

func (v *BlockView) SetHash(hash []byte) {
	v.hash.SetText(base64.RawURLEncoding.EncodeToString(hash))
}

func (v *BlockView) SetBlock(block *bcgo.Block) {
	v.timestamp.SetTimestamp(block.Timestamp)
	v.channel.SetText(block.ChannelName)
	v.channel.OnTapped = func() {
		v.ui.ShowURI(v.client, storage.NewChannelURI(block.ChannelName))
	}
	v.length.SetText(fmt.Sprintf("%d", block.Length))
	v.previous.SetText(base64.RawURLEncoding.EncodeToString(block.Previous))
	v.previous.OnTapped = func() {
		v.ui.ShowURI(v.client, storage.NewBlockURI(block.ChannelName, block.Previous))
	}
	v.miner.SetText(block.Miner)
	v.miner.OnTapped = func() {
		v.ui.ShowURI(v.client, storage.NewAliasURI(block.Miner))
	}
	v.nonce.SetText(fmt.Sprintf("%d", block.Nonce))
	var records []fyne.CanvasObject
	for _, e := range block.Entry {
		v := NewRecordView(v.ui, v.client)
		v.SetHash(e.RecordHash)
		v.SetRecord(e.Record)
		records = append(records, v)
	}
	v.record.Objects = records
	v.record.Refresh()
	v.Refresh()
}
