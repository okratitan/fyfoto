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
	"aletheiaware.com/cryptogo"
	"crypto/rsa"
	"errors"
	"fmt"
	"github.com/golang/protobuf/proto"
	"sort"
)

const (
	ERROR_NO_ENTRIES_TO_MINE = "No entries to mine for channel: %s"
	ERROR_NO_SUCH_CHANNEL    = "No such channel: %s"
	ERROR_PAYLOAD_TOO_LARGE  = "Payload too large: %s max: %s"
	ERROR_BLOCK_TOO_LARGE    = "Block too large: %s max: %s"
	ERROR_NONCE_WRAP_AROUND  = "Nonce wrapped around before reaching threshold"
)

type MiningListener interface {
	OnMiningStarted(channel *Channel, size uint64)
	OnNewMaxOnes(channel *Channel, nonce, ones uint64)
	OnMiningThresholdReached(channel *Channel, hash []byte, block *Block)
}

type Node struct {
	Alias    string
	Key      *rsa.PrivateKey
	Cache    Cache
	Network  Network
	Channels map[string]*Channel
}

func NewNode(directory string, cache Cache, network Network) (*Node, error) {
	// Get alias
	alias, err := GetAlias()
	if err != nil {
		return nil, err
	}
	keystore, err := GetKeyDirectory(directory)
	if err != nil {
		return nil, err
	}
	// Get private key
	key, err := cryptogo.GetOrCreateRSAPrivateKey(keystore, alias)
	if err != nil {
		return nil, err
	}
	return &Node{
		Alias:    alias,
		Key:      key,
		Cache:    cache,
		Network:  network,
		Channels: make(map[string]*Channel),
	}, nil
}

func (n *Node) AddChannel(channel *Channel) {
	n.Channels[channel.Name] = channel
}

func (n *Node) GetChannel(name string) (*Channel, error) {
	c, ok := n.Channels[name]
	if !ok {
		return nil, fmt.Errorf(ERROR_NO_SUCH_CHANNEL, name)
	}
	return c, nil
}

func (n *Node) GetOrOpenChannel(name string, opener func() *Channel) *Channel {
	c, ok := n.Channels[name]
	if !ok {
		c = opener()
		if c != nil {
			if err := c.LoadCachedHead(n.Cache); err != nil {
				// Ignored
			}
			n.AddChannel(c)
		}
	}
	return c
}

func (n *Node) GetChannels() []*Channel {
	var keys []string
	for k := range n.Channels {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var channels []*Channel
	for _, k := range keys {
		channels = append(channels, n.Channels[k])
	}
	return channels
}

func (n *Node) Write(timestamp uint64, channel *Channel, acl map[string]*rsa.PublicKey, references []*Reference, payload []byte) (*Reference, error) {
	size := uint64(len(payload))
	if size > MAX_PAYLOAD_SIZE_BYTES {
		return nil, fmt.Errorf(ERROR_PAYLOAD_TOO_LARGE, BinarySizeToString(size), BinarySizeToString(MAX_PAYLOAD_SIZE_BYTES))
	}
	_, record, err := CreateRecord(timestamp, n.Alias, n.Key, acl, references, payload)
	if err != nil {
		return nil, err
	}
	return WriteRecord(channel.Name, n.Cache, record)
}

func (n *Node) MineProto(channel *Channel, threshold uint64, listener MiningListener, acl map[string]*rsa.PublicKey, references []*Reference, message proto.Message) error {
	data, err := proto.Marshal(message)
	if err != nil {
		return err
	}

	if _, err := n.Write(Timestamp(), channel, acl, references, data); err != nil {
		return err
	}

	if _, _, err := n.Mine(channel, threshold, listener); err != nil {
		return err
	}

	if err := channel.Push(n.Cache, n.Network); err != nil {
		return err
	}
	return nil
}

func (n *Node) GetLastMinedTimestamp(channel *Channel) (uint64, error) {
	var timestamp uint64
	// Iterate through the chain to find the most recent block mined by this node
	if err := Iterate(channel.Name, channel.Head, nil, n.Cache, n.Network, func(h []byte, b *Block) error {
		if b.Miner == n.Alias {
			timestamp = b.Timestamp
			return StopIterationError{}
		}
		return nil
	}); err != nil {
		switch err.(type) {
		case StopIterationError:
			// Do nothing
			break
		default:
			return 0, err
		}
	}
	return timestamp, nil
}

func (n *Node) Mine(channel *Channel, threshold uint64, listener MiningListener) ([]byte, *Block, error) {
	timestamp, err := n.GetLastMinedTimestamp(channel)
	if err != nil {
		return nil, nil, err
	}

	entries, err := n.Cache.GetBlockEntries(channel.Name, timestamp)
	if err != nil {
		return nil, nil, err
	}

	// Sort by timestamp
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Record.Timestamp < entries[j].Record.Timestamp
	})

	return n.MineEntries(channel, threshold, listener, entries)
}

func (n *Node) MineEntries(channel *Channel, threshold uint64, listener MiningListener, entries []*BlockEntry) ([]byte, *Block, error) {
	if len(entries) == 0 {
		return nil, nil, fmt.Errorf(ERROR_NO_ENTRIES_TO_MINE, channel.Name)
	}

	// TODO check record signature of each entry

	block := &Block{
		Timestamp:   Timestamp(),
		ChannelName: channel.Name,
		Length:      1,
		Miner:       n.Alias,
		Entry:       entries,
	}

	previousHash := channel.Head
	if previousHash != nil {
		previousBlock, err := n.Cache.GetBlock(previousHash)
		if err != nil {
			return nil, nil, err
		}
		block.Length = previousBlock.Length + 1
		block.Previous = previousHash
	}

	return n.MineBlock(channel, threshold, listener, block)
}

func (n *Node) MineBlock(channel *Channel, threshold uint64, listener MiningListener, block *Block) ([]byte, *Block, error) {
	size := uint64(proto.Size(block))
	if size > MAX_BLOCK_SIZE_BYTES {
		return nil, nil, fmt.Errorf(ERROR_BLOCK_TOO_LARGE, BinarySizeToString(size), BinarySizeToString(MAX_BLOCK_SIZE_BYTES))
	}

	if listener != nil {
		listener.OnMiningStarted(channel, size)
	}

	var max uint64
	for nonce := uint64(1); nonce > 0; nonce++ {
		block.Nonce = nonce
		hash, err := cryptogo.HashProtobuf(block)
		if err != nil {
			return nil, nil, err
		}
		ones := Ones(hash)
		if ones > max {
			if listener != nil {
				listener.OnNewMaxOnes(channel, nonce, ones)
			}
			max = ones
		}
		if ones > threshold {
			if listener != nil {
				listener.OnMiningThresholdReached(channel, hash, block)
			}
			if err := channel.Update(n.Cache, n.Network, hash, block); err != nil {
				return nil, nil, err
			}
			return hash, block, nil
		}
	}
	return nil, nil, errors.New(ERROR_NONCE_WRAP_AROUND)
}
