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
	"fmt"
)

const (
	ERROR_HASH_TOO_WEAK = "Hash doesn't meet Proof-of-Work threshold: %d vs %d"
)

func OpenPoWChannel(name string, threshold uint64) *Channel {
	return &Channel{
		Name: name,
		Validators: []Validator{
			&PoWValidator{
				Threshold: threshold,
			},
		},
	}
}

type PoWValidator struct {
	Threshold uint64
}

func (p *PoWValidator) Validate(channel *Channel, cache Cache, network Network, hash []byte, block *Block) error {
	return Iterate(channel.Name, hash, block, cache, network, func(h []byte, b *Block) error {
		// Check hash ones pass threshold
		ones := Ones(h)
		if ones < p.Threshold {
			return fmt.Errorf(ERROR_HASH_TOO_WEAK, ones, p.Threshold)
		}
		return nil
	})
}
