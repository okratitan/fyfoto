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

package bcgo

import (
	"encoding/base64"
	"fmt"
)

const (
	ERROR_DUPLICATE_BLOCK = "Duplicate Block: %s"
	ERROR_DUPLICATE_ENTRY = "Duplicate Entry: %s"
)

type UniqueValidator struct {
}

func (v *UniqueValidator) Validate(channel *Channel, cache Cache, network Network, hash []byte, block *Block) error {
	blocks := make(map[string]bool)
	entries := make(map[string]bool)
	return Iterate(channel.Name, hash, block, cache, network, func(h []byte, b *Block) error {
		id := base64.RawURLEncoding.EncodeToString(h)
		if _, ok := blocks[id]; ok {
			return fmt.Errorf(ERROR_DUPLICATE_BLOCK, id)
		} else {
			blocks[id] = true
		}
		for _, entry := range b.Entry {
			id := base64.RawURLEncoding.EncodeToString(entry.RecordHash)
			if _, ok := entries[id]; ok {
				return fmt.Errorf(ERROR_DUPLICATE_ENTRY, id)
			} else {
				entries[id] = true
			}
		}
		return nil
	})
}
