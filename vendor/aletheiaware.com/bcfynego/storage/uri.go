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

package storage

import (
	"encoding/base64"
	"fyne.io/fyne/v2"
)

const (
	ALIAS_SCHEME        = "alias"
	ALIAS_SCHEME_PREFIX = "alias:"
	BC_SCHEME           = "bc"
	BC_SCHEME_PREFIX    = "bc:"
)

type AliasURI interface {
	fyne.URI
	Alias() string
}

func NewAliasURI(alias string) AliasURI {
	return &aliasURI{
		alias: alias,
	}
}

type ChannelURI interface {
	fyne.URI
	Channel() string
}

func NewChannelURI(channel string) ChannelURI {
	return &channelURI{
		channel: channel,
	}
}

type BlockURI interface {
	ChannelURI
	BlockHash() []byte
}

func NewBlockURI(channel string, blockHash []byte) BlockURI {
	return &blockURI{
		channel:   channel,
		blockHash: blockHash,
	}
}

type RecordURI interface {
	BlockURI
	RecordHash() []byte
}

func NewRecordURI(channel string, blockHash, recordHash []byte) RecordURI {
	return &recordURI{
		channel:    channel,
		blockHash:  blockHash,
		recordHash: recordHash,
	}
}

type aliasURI struct {
	alias string
}

func (u *aliasURI) Alias() string {
	return u.alias
}

func (u *aliasURI) Authority() string {
	return ""
}

func (u *aliasURI) Extension() string {
	return ""
}

func (u *aliasURI) Fragment() string {
	return ""
}

func (u *aliasURI) MimeType() string {
	return "" // TODO protobuf type
}

func (u *aliasURI) Name() string {
	return u.alias
}

func (u *aliasURI) Path() string {
	return ""
}

func (u *aliasURI) Query() string {
	return ""
}

func (u *aliasURI) Scheme() string {
	return ALIAS_SCHEME
}

func (u *aliasURI) String() string {
	return ALIAS_SCHEME_PREFIX + u.alias
}

type channelURI struct {
	channel string
}

func (u *channelURI) Authority() string {
	return ""
}

func (u *channelURI) Channel() string {
	return u.channel
}

func (u *channelURI) Extension() string {
	return ""
}

func (u *channelURI) Fragment() string {
	return ""
}

func (u *channelURI) MimeType() string {
	return "" // TODO protobuf type
}

func (u *channelURI) Name() string {
	return u.channel
}

func (u *channelURI) Path() string {
	return ""
}

func (u *channelURI) Query() string {
	return ""
}

func (u *channelURI) Scheme() string {
	return BC_SCHEME
}

func (u *channelURI) String() string {
	return BC_SCHEME_PREFIX + u.channel
}

type blockURI struct {
	channel   string
	blockHash []byte
}

func (u *blockURI) Authority() string {
	return ""
}

func (u *blockURI) BlockHash() []byte {
	return u.blockHash
}

func (u *blockURI) Channel() string {
	return u.channel
}

func (u *blockURI) Extension() string {
	return ""
}

func (u *blockURI) Fragment() string {
	return ""
}

func (u *blockURI) MimeType() string {
	return "" // TODO protobuf type
}

func (u *blockURI) Name() string {
	return base64.RawURLEncoding.EncodeToString(u.blockHash)
}

func (u *blockURI) Path() string {
	return ""
}

func (u *blockURI) Query() string {
	return ""
}

func (u *blockURI) Scheme() string {
	return BC_SCHEME
}

func (u *blockURI) String() string {
	return BC_SCHEME_PREFIX + u.channel + "/" + base64.RawURLEncoding.EncodeToString(u.blockHash)
}

type recordURI struct {
	channel               string
	blockHash, recordHash []byte
}

func (u *recordURI) Authority() string {
	return ""
}

func (u *recordURI) BlockHash() []byte {
	return u.blockHash
}

func (u *recordURI) Channel() string {
	return u.channel
}

func (u *recordURI) Extension() string {
	return ""
}

func (u *recordURI) Fragment() string {
	return ""
}

func (u *recordURI) MimeType() string {
	return "" // TODO protobuf type
}

func (u *recordURI) Name() string {
	return base64.RawURLEncoding.EncodeToString(u.recordHash)
}

func (u *recordURI) Path() string {
	return ""
}

func (u *recordURI) Query() string {
	return ""
}

func (u *recordURI) RecordHash() []byte {
	return u.recordHash
}

func (u *recordURI) Scheme() string {
	return BC_SCHEME
}

func (u *recordURI) String() string {
	return BC_SCHEME_PREFIX + u.channel + "/" + base64.RawURLEncoding.EncodeToString(u.blockHash) + "/" + base64.RawURLEncoding.EncodeToString(u.recordHash)
}
