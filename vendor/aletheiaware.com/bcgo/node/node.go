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

package node

import (
	"aletheiaware.com/bcgo"
	"sort"
	"sync"
)

type node struct {
	account  bcgo.Account
	cache    bcgo.Cache
	network  bcgo.Network
	channels map[string]bcgo.Channel
	lock     sync.Mutex
}

func New(account bcgo.Account, cache bcgo.Cache, network bcgo.Network) bcgo.Node {
	return &node{
		account:  account,
		cache:    cache,
		network:  network,
		channels: make(map[string]bcgo.Channel),
	}
}

func (n *node) Account() bcgo.Account {
	return n.account
}

func (n *node) Cache() bcgo.Cache {
	return n.cache
}

func (n *node) Network() bcgo.Network {
	return n.network
}

func (n *node) AddChannel(channel bcgo.Channel) {
	n.lock.Lock()
	defer n.lock.Unlock()
	n.channels[channel.Name()] = channel
}

func (n *node) Channel(name string) (bcgo.Channel, error) {
	n.lock.Lock()
	defer n.lock.Unlock()
	c, ok := n.channels[name]
	if !ok {
		return nil, bcgo.ErrNoSuchChannel{Channel: name}
	}
	return c, nil
}

func (n *node) OpenChannel(name string, opener func() bcgo.Channel) bcgo.Channel {
	n.lock.Lock()
	c, ok := n.channels[name]
	n.lock.Unlock()
	if !ok {
		c = opener()
		if c != nil {
			if err := c.Load(n.cache, nil); err != nil {
				// Ignored
			}
			n.AddChannel(c)
		}
	}
	return c
}

func (n *node) Channels() []bcgo.Channel {
	n.lock.Lock()
	defer n.lock.Unlock()
	var keys []string
	for k := range n.channels {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var channels []bcgo.Channel
	for _, k := range keys {
		channels = append(channels, n.channels[k])
	}
	return channels
}

func (n *node) Write(timestamp uint64, channel bcgo.Channel, access []bcgo.Identity, references []*bcgo.Reference, payload []byte) (*bcgo.Reference, error) {
	size := uint64(len(payload))
	if size > bcgo.MAX_PAYLOAD_SIZE_BYTES {
		return nil, bcgo.ErrPayloadTooLarge{Size: size, Max: bcgo.MAX_PAYLOAD_SIZE_BYTES}
	}
	_, record, err := bcgo.CreateRecord(timestamp, n.account, access, references, payload)
	if err != nil {
		return nil, err
	}
	return bcgo.WriteRecord(channel.Name(), n.cache, record)
}
