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
	"bytes"
	"encoding/base64"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"sort"
)

type RecordView struct {
	widget.Form
	ui                   UI
	client               bcclientgo.BCClient
	hash                 *widget.Label
	timestamp            *TimestampLabel
	creator              *Link
	access               *fyne.Container
	payload              *widget.Label
	compressionAlgorithm *widget.Label
	encryptionAlgorithm  *widget.Label
	signature            *widget.Label
	signatureAlgorithm   *widget.Label
	reference            *fyne.Container
	meta                 *fyne.Container
}

func NewRecordView(ui UI, client bcclientgo.BCClient) *RecordView {
	v := &RecordView{
		ui:     ui,
		client: client,
		hash: &widget.Label{
			TextStyle: fyne.TextStyle{
				Monospace: true,
			},
			Wrapping: fyne.TextWrapBreak,
		},
		timestamp: NewTimestampLabel(0),
		creator: &Link{
			Hyperlink: widget.Hyperlink{
				TextStyle: fyne.TextStyle{
					Monospace: true,
				},
				Wrapping: fyne.TextWrapBreak,
			},
		},
		access: container.NewVBox(),
		payload: &widget.Label{
			TextStyle: fyne.TextStyle{
				Monospace: true,
			},
			Wrapping: fyne.TextWrapBreak,
		},
		compressionAlgorithm: &widget.Label{
			TextStyle: fyne.TextStyle{
				Monospace: true,
			},
			Wrapping: fyne.TextWrapBreak,
		},
		encryptionAlgorithm: &widget.Label{
			TextStyle: fyne.TextStyle{
				Monospace: true,
			},
			Wrapping: fyne.TextWrapBreak,
		},
		signature: &widget.Label{
			TextStyle: fyne.TextStyle{
				Monospace: true,
			},
			Wrapping: fyne.TextWrapBreak,
		},
		signatureAlgorithm: &widget.Label{
			TextStyle: fyne.TextStyle{
				Monospace: true,
			},
			Wrapping: fyne.TextWrapBreak,
		},
		reference: container.NewVBox(),
		meta:      container.NewVBox(),
	}
	v.ExtendBaseWidget(v)
	v.hash.ExtendBaseWidget(v.hash)
	v.timestamp.ExtendBaseWidget(v.timestamp)
	v.creator.ExtendBaseWidget(v.creator)
	v.payload.ExtendBaseWidget(v.payload)
	v.compressionAlgorithm.ExtendBaseWidget(v.compressionAlgorithm)
	v.encryptionAlgorithm.ExtendBaseWidget(v.encryptionAlgorithm)
	v.signature.ExtendBaseWidget(v.signature)
	v.signatureAlgorithm.ExtendBaseWidget(v.signatureAlgorithm)
	v.Append("Hash", v.hash)
	v.Append("Timestamp", v.timestamp)
	v.Append("Creator", v.creator)
	v.Append("Access", v.access)
	v.Append("Payload", v.payload)
	v.Append("Compression", v.compressionAlgorithm)
	v.Append("Encryption", v.encryptionAlgorithm)
	v.Append("Signature", v.signature)
	v.Append("Signature", v.signatureAlgorithm)
	v.Append("References", v.reference)
	v.Append("Metadata", v.meta)
	return v
}

func (v *RecordView) SetURI(uri storage.RecordURI) error {
	name := uri.Channel()
	node, err := v.client.Node()
	if err != nil {
		return err
	}
	blockHash := uri.BlockHash()
	recordHash := uri.RecordHash()
	var block *bcgo.Block
	if blockHash == nil || len(blockHash) == 0 {
		block, err = bcgo.LoadBlockContainingRecord(name, node.Cache(), node.Network(), recordHash)
	} else {
		block, err = bcgo.LoadBlock(name, node.Cache(), node.Network(), blockHash)
	}
	if err != nil {
		return err
	}
	v.SetHash(recordHash)
	for _, entry := range block.Entry {
		if bytes.Equal(recordHash, entry.RecordHash) {
			v.SetRecord(entry.Record)
			break
		}
	}
	return nil
}

func (v *RecordView) SetHash(hash []byte) {
	v.hash.SetText(base64.RawURLEncoding.EncodeToString(hash))
}

func (v *RecordView) SetRecord(record *bcgo.Record) {
	v.timestamp.SetTimestamp(record.Timestamp)
	v.creator.SetText(record.Creator)
	v.creator.OnTapped = func() {
		v.ui.ShowURI(v.client, storage.NewAliasURI(record.Creator))
	}
	var accesses []fyne.CanvasObject
	for _, a := range record.Access {
		v := NewAccessView(v.ui, v.client)
		v.SetAccess(a)
		accesses = append(accesses, v)
	}
	v.access.Objects = accesses
	v.access.Refresh()
	v.payload.SetText(base64.RawURLEncoding.EncodeToString(record.Payload))
	v.compressionAlgorithm.SetText(record.CompressionAlgorithm.String())
	v.encryptionAlgorithm.SetText(record.EncryptionAlgorithm.String())
	v.signature.SetText(base64.RawURLEncoding.EncodeToString(record.Signature))
	v.signatureAlgorithm.SetText(record.SignatureAlgorithm.String())
	var references []fyne.CanvasObject
	for _, r := range record.Reference {
		v := NewReferenceView(v.ui, v.client)
		v.SetReference(r)
		references = append(references, v)
	}
	v.reference.Objects = references
	v.reference.Refresh()
	keys := make([]string, len(record.Meta))
	i := 0
	for k := range record.Meta {
		keys[i] = k
		i++
	}
	sort.Strings(keys)
	var metas []fyne.CanvasObject
	for _, k := range keys {
		metas = append(metas, container.NewGridWithColumns(2,
			&widget.Label{
				Text: k,
				TextStyle: fyne.TextStyle{
					Monospace: true,
				},
			},
			&widget.Label{
				Text: record.Meta[k],
				TextStyle: fyne.TextStyle{
					Monospace: true,
				},
			},
		))
	}
	v.meta.Objects = metas
	v.meta.Refresh()
	v.Refresh()
}
