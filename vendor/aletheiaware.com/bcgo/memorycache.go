/*
 * Copyright 2019 Aletheia Ware LLC
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

package bcgo

import (
	"encoding/base64"
	"fmt"
)

const (
	ERROR_BLOCK_NOT_FOUND                   = "Block not found %s"
	ERROR_HEAD_NOT_FOUND                    = "Head not found %s"
	ERROR_RECORD_TO_BLOCK_MAPPING_NOT_FOUND = "Record to Block Mapping not found %s"
)

type MemoryCache struct {
	Blocks  map[string]*Block
	Heads   map[string]*Reference
	Entries map[string][]*BlockEntry
	Mapping map[string]*Block
}

func NewMemoryCache(size int) *MemoryCache {
	// TODO size for blocks, heads, entries
	// TODO implement LRU
	// TODO implement cache levels where
	return &MemoryCache{
		Blocks:  make(map[string]*Block, size),
		Heads:   make(map[string]*Reference, size),
		Entries: make(map[string][]*BlockEntry, size),
		Mapping: make(map[string]*Block, size),
	}
}

func (m *MemoryCache) GetBlock(hash []byte) (*Block, error) {
	key := base64.RawURLEncoding.EncodeToString(hash)
	block, ok := m.Blocks[key]
	if !ok {
		return nil, fmt.Errorf(ERROR_BLOCK_NOT_FOUND, key)
	}
	return block, nil
}

func (m *MemoryCache) GetBlockEntries(channel string, timestamp uint64) ([]*BlockEntry, error) {
	var results []*BlockEntry
	for _, e := range m.Entries[channel] {
		if e.Record.Timestamp >= timestamp {
			results = append(results, e)
		}
	}
	return results, nil
}

func (m *MemoryCache) GetBlockContainingRecord(channel string, hash []byte) (*Block, error) {
	key := base64.RawURLEncoding.EncodeToString(hash)
	block, ok := m.Mapping[key]
	if !ok {
		return nil, fmt.Errorf(ERROR_RECORD_TO_BLOCK_MAPPING_NOT_FOUND, key)
	}
	return block, nil
}

func (m *MemoryCache) GetHead(channel string) (*Reference, error) {
	reference, ok := m.Heads[channel]
	if !ok {
		return nil, fmt.Errorf(ERROR_HEAD_NOT_FOUND, channel)
	}
	return reference, nil
}

func (m *MemoryCache) PutBlock(hash []byte, block *Block) error {
	m.Blocks[base64.RawURLEncoding.EncodeToString(hash)] = block
	for _, e := range block.Entry {
		m.Mapping[base64.RawURLEncoding.EncodeToString(e.RecordHash)] = block
	}
	return nil
}

func (m *MemoryCache) PutBlockEntry(channel string, entry *BlockEntry) error {
	m.Entries[channel] = append(m.Entries[channel], entry)
	return nil
}

func (m *MemoryCache) PutHead(channel string, reference *Reference) error {
	m.Heads[channel] = reference
	return nil
}

// func (m *MemoryCache) DeleteBlock(hash []byte) error {
// 	delete(m.Blocks, base64.RawURLEncoding.EncodeToString(hash))
// 	return nil
// }
