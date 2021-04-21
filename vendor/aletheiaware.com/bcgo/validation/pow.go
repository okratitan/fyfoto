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

package validation

import (
	"aletheiaware.com/bcgo"
)

type PoW interface {
	bcgo.Validator
}

type pow struct {
	threshold uint64
}

func NewPoW(threshold uint64) PoW {
	return &pow{
		threshold: threshold,
	}
}

func (v *pow) Validate(channel bcgo.Channel, cache bcgo.Cache, network bcgo.Network, hash []byte, block *bcgo.Block) error {
	return bcgo.Iterate(channel.Name(), hash, block, cache, network, func(h []byte, b *bcgo.Block) error {
		// Check hash ones pass threshold
		ones := bcgo.Ones(h)
		if ones < v.threshold {
			return ErrHashTooWeak{Expected: v.threshold, Actual: ones}
		}
		return nil
	})
}
