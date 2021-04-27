/*
 * Copyright 2019-21 Aletheia Ware LLC
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

package cache

import (
	"aletheiaware.com/bcgo"
	"encoding/base64"
)

type Memory struct {
	Blocks  map[string]*bcgo.Block
	Heads   map[string]*bcgo.Reference
	Entries map[string][]*bcgo.BlockEntry
	Mapping map[string]*bcgo.Block
}

func NewMemory(size int) *Memory {
	// TODO separate sizes for blocks, heads, entries, mappings
	// TODO implement LRU
	// TODO implement cache levels where a memory cache sits above a file cache
	return &Memory{
		Blocks:  make(map[string]*bcgo.Block, size),
		Heads:   make(map[string]*bcgo.Reference, size),
		Entries: make(map[string][]*bcgo.BlockEntry, size),
		Mapping: make(map[string]*bcgo.Block, size),
	}
}

func (m *Memory) Block(hash []byte) (*bcgo.Block, error) {
	key := base64.RawURLEncoding.EncodeToString(hash)
	block, ok := m.Blocks[key]
	if !ok {
		return nil, bcgo.ErrNoSuchBlock{Hash: key}
	}
	return block, nil
}

func (m *Memory) BlockEntries(channel string, timestamp uint64) ([]*bcgo.BlockEntry, error) {
	var results []*bcgo.BlockEntry
	for _, e := range m.Entries[channel] {
		if e.Record.Timestamp >= timestamp {
			results = append(results, e)
		}
	}
	return results, nil
}

func (m *Memory) BlockContainingRecord(channel string, hash []byte) (*bcgo.Block, error) {
	key := base64.RawURLEncoding.EncodeToString(hash)
	block, ok := m.Mapping[key]
	if !ok {
		return nil, bcgo.ErrNoSuchMapping{Hash: key}
	}
	return block, nil
}

func (m *Memory) Head(channel string) (*bcgo.Reference, error) {
	reference, ok := m.Heads[channel]
	if !ok {
		return nil, bcgo.ErrNoSuchHead{Channel: channel}
	}
	return reference, nil
}

func (m *Memory) PutBlock(hash []byte, block *bcgo.Block) error {
	m.Blocks[base64.RawURLEncoding.EncodeToString(hash)] = block
	for _, e := range block.Entry {
		m.Mapping[base64.RawURLEncoding.EncodeToString(e.RecordHash)] = block
	}
	return nil
}

func (m *Memory) PutBlockEntry(channel string, entry *bcgo.BlockEntry) error {
	m.Entries[channel] = append(m.Entries[channel], entry)
	return nil
}

func (m *Memory) PutHead(channel string, reference *bcgo.Reference) error {
	m.Heads[channel] = reference
	return nil
}

// func (m *Memory) DeleteBlock(hash []byte) error {
// 	delete(m.Blocks, base64.RawURLEncoding.EncodeToString(hash))
// 	return nil
// }
