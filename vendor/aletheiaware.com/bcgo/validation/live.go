/*
 * Copyright 2020-21 Aletheia Ware LLC
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

package validation

import (
	"aletheiaware.com/bcgo"
	"os"
)

// Live is a validator which ensures the Live flag in each Records' Metadata matches the Live environment variable.
type Live struct {
}

// Validate ensures all records have a live flag in their metadata which matches the environment variable.
func (v *Live) Validate(channel bcgo.Channel, cache bcgo.Cache, network bcgo.Network, hash []byte, block *bcgo.Block) error {
	expected := os.Getenv(bcgo.LIVE_FLAG)
	return bcgo.Iterate(channel.Name(), hash, block, cache, network, func(h []byte, b *bcgo.Block) error {
		for _, entry := range b.Entry {
			var flag string
			if meta := entry.Record.Meta; meta != nil {
				flag = meta[bcgo.LIVE_FLAG]
			}
			if flag != expected {
				return ErrDifferentLiveFlag{Expected: expected, Actual: flag}
			}
		}
		return nil
	})
}
