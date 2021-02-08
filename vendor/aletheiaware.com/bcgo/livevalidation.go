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
	"fmt"
	"os"
)

// Live Validation ensures the Live flag in each Records' Metadata matches the Live environment variable.

const (
	ERROR_DIFFERENT_LIVE_FLAG = "Different Live Flag; Expected '%s', got '%s'"
)

type LiveValidator struct {
}

// Validate ensures all records have a live flag in their metadata which matches the environment variable.
func (p *LiveValidator) Validate(channel *Channel, cache Cache, network Network, hash []byte, block *Block) error {
	expected := os.Getenv(LIVE_FLAG)
	return Iterate(channel.Name, hash, block, cache, network, func(h []byte, b *Block) error {
		for _, entry := range b.Entry {
			var flag string
			if meta := entry.Record.Meta; meta != nil {
				flag = meta[LIVE_FLAG]
			}
			if flag != expected {
				return fmt.Errorf(ERROR_DIFFERENT_LIVE_FLAG, expected, flag)
			}
		}
		return nil
	})
}
