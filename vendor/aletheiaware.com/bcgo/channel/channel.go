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

package channel

import (
	"aletheiaware.com/bcgo"
	"aletheiaware.com/cryptogo"
	"bytes"
	"reflect"
	"sync"
)

type channel struct {
	name       string
	head       []byte
	timestamp  uint64
	triggers   []func()
	validators []bcgo.Validator
	lock       sync.Mutex
}

func New(name string) bcgo.Channel {
	return &channel{
		name: name,
	}
}

func (c *channel) Name() string {
	return c.name
}

func (c *channel) Head() []byte {
	return c.head
}

func (c *channel) Timestamp() uint64 {
	return c.timestamp
}

func (c *channel) String() string {
	return c.name
}

func (c *channel) AddTrigger(trigger func()) {
	c.triggers = append(c.triggers, trigger)
	trigger()
}

func (c *channel) AddValidator(validator bcgo.Validator) {
	c.validators = append(c.validators, validator)
}

func (c *channel) Update(cache bcgo.Cache, network bcgo.Network, head []byte, block *bcgo.Block) error {
	c.lock.Lock()
	defer c.lock.Unlock()
	if bytes.Equal(c.head, head) {
		// Channel up to date
		return nil
	}

	if c.name != block.ChannelName {
		return bcgo.ErrChannelNameIncorrect{Expected: c.name, Actual: block.ChannelName}
	}

	if err := bcgo.ValidateName(block.ChannelName); err != nil {
		return err
	}

	// Check hash matches block hash
	h, err := cryptogo.HashProtobuf(block)
	if err != nil {
		return err
	}
	if !bytes.Equal(head, h) {
		return bcgo.ErrBlockHashIncorrect{}
	}

	// Check block is valid
	for _, v := range c.validators {
		if err := v.Validate(c, cache, network, head, block); err != nil {
			return bcgo.ErrChainInvalid{Reason: err.Error()}
		}
	}

	if c.head != nil {
		b, err := cache.Block(c.head)
		if err != nil {
			return err
		}
		// Check block chain is longer than current head
		if b != nil && b.Length >= block.Length {
			valid := true
			// Check current head is still valid
			for _, v := range c.validators {
				if err := v.Validate(c, cache, network, c.head, b); err != nil {
					valid = false
				}
			}
			if valid {
				// Current head is still valid and update is not long enough to replace it
				return bcgo.ErrChainTooShort{LengthA: b.Length, LengthB: block.Length}
			}
		}
	}

	if err := cache.PutHead(c.name, &bcgo.Reference{
		Timestamp:   block.Timestamp,
		ChannelName: c.name,
		BlockHash:   head,
	}); err != nil {
		return err
	}
	if err := cache.PutBlock(head, block); err != nil {
		return err
	}
	c.Set(block.Timestamp, head)
	return nil
}

func (c *channel) Load(cache bcgo.Cache, network bcgo.Network) error {
	reference, err := bcgo.LoadHead(c.name, cache, network)
	if err != nil {
		return err
	}
	if c.timestamp < reference.Timestamp {
		c.Set(reference.Timestamp, reference.BlockHash)
	}
	return nil
}

func (c *channel) Refresh(cache bcgo.Cache, network bcgo.Network) error {
	// Load Channel
	err := c.Load(cache, nil)
	// Pull from network regardless of above err
	if network == nil || reflect.ValueOf(network).IsNil() {
		return err
	}
	// Pull Channel
	err = c.Pull(cache, network)
	switch e := err.(type) {
	case bcgo.ErrChainTooShort:
		if e.LengthA > e.LengthB {
			// The local chain is longer than the remote, push local chain to remote.
			err = c.Push(cache, network)
		}
	}
	return err
}

func (c *channel) Pull(cache bcgo.Cache, network bcgo.Network) error {
	if network == nil || reflect.ValueOf(network).IsNil() {
		return nil
	}
	reference, err := network.Head(c.name)
	if err != nil {
		return err
	}
	hash := reference.BlockHash
	if bytes.Equal(c.head, hash) {
		// Channel up-to-date
		return nil
	}
	// Load head block
	block, err := bcgo.LoadBlock(c.name, cache, network, hash)
	if err != nil {
		return err
	}
	// Ensure all previous blocks are loaded
	b := block
	for b != nil {
		h := b.Previous
		if h != nil && len(h) > 0 {
			b, err = bcgo.LoadBlock(c.name, cache, network, h)
			if err != nil {
				return err
			}
		} else {
			b = nil
		}
	}
	if err := c.Update(cache, network, hash, block); err != nil {
		return err
	}
	return nil
}

func (c *channel) Push(cache bcgo.Cache, network bcgo.Network) error {
	hash := c.head
	if hash == nil {
		reference, err := cache.Head(c.name)
		if err != nil {
			return err
		}
		hash = reference.BlockHash
	}
	block, err := cache.Block(hash)
	if err != nil {
		return err
	}
	return network.Broadcast(c, cache, hash, block)
}

func (c *channel) Set(timestamp uint64, head []byte) {
	c.timestamp = timestamp
	c.head = head
	for _, t := range c.triggers {
		t()
	}
}
