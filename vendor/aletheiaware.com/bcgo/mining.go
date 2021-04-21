/*
 * Copyright 2021 Aletheia Ware LLC
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
	"github.com/golang/protobuf/proto"
	"sort"
)

type MiningListener interface {
	OnMiningStarted(channel Channel, size uint64)
	OnNewMaxOnes(channel Channel, nonce, ones uint64)
	OnMiningThresholdReached(channel Channel, hash []byte, block *Block)
}

func MineProto(node Node, channel Channel, threshold uint64, listener MiningListener, access []Identity, references []*Reference, message proto.Message) error {
	data, err := proto.Marshal(message)
	if err != nil {
		return err
	}

	if _, err := node.Write(Timestamp(), channel, access, references, data); err != nil {
		return err
	}

	if _, _, err := Mine(node, channel, threshold, listener); err != nil {
		return err
	}

	if err := channel.Push(node.Cache(), node.Network()); err != nil {
		return err
	}
	return nil
}

func LastMinedTimestamp(node Node, channel Channel) (uint64, error) {
	alias := node.Account().Alias()
	var timestamp uint64
	// Iterate through the chain to find the most recent block mined by this node
	if err := Iterate(channel.Name(), channel.Head(), nil, node.Cache(), node.Network(), func(h []byte, b *Block) error {
		if b.Miner == alias {
			timestamp = b.Timestamp
			return ErrStopIteration{}
		}
		return nil
	}); err != nil {
		switch err.(type) {
		case ErrStopIteration:
			// Do nothing
			break
		default:
			return 0, err
		}
	}
	return timestamp, nil
}

func Mine(node Node, channel Channel, threshold uint64, listener MiningListener) ([]byte, *Block, error) {
	timestamp, err := LastMinedTimestamp(node, channel)
	if err != nil {
		return nil, nil, err
	}

	entries, err := node.Cache().BlockEntries(channel.Name(), timestamp)
	if err != nil {
		return nil, nil, err
	}

	// Sort by timestamp
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Record.Timestamp < entries[j].Record.Timestamp
	})

	return MineEntries(node, channel, threshold, listener, entries)
}

func MineEntries(node Node, channel Channel, threshold uint64, listener MiningListener, entries []*BlockEntry) ([]byte, *Block, error) {
	if len(entries) == 0 {
		return nil, nil, ErrNoEntriesToMine{channel.Name()}
	}

	// TODO check record signature of each entry

	block := &Block{
		Timestamp:   Timestamp(),
		ChannelName: channel.Name(),
		Length:      1,
		Miner:       node.Account().Alias(),
		Entry:       entries,
	}

	previousHash := channel.Head()
	if previousHash != nil {
		previousBlock, err := node.Cache().Block(previousHash)
		if err != nil {
			return nil, nil, err
		}
		block.Length = previousBlock.Length + 1
		block.Previous = previousHash
	}

	return MineBlock(node, channel, threshold, listener, block)
}

func MineBlock(node Node, channel Channel, threshold uint64, listener MiningListener, block *Block) ([]byte, *Block, error) {
	size := uint64(proto.Size(block))
	if size > MAX_BLOCK_SIZE_BYTES {
		return nil, nil, ErrBlockTooLarge{size, MAX_BLOCK_SIZE_BYTES}
	}

	if listener != nil {
		listener.OnMiningStarted(channel, size)
	}

	var max uint64
	// TODO only mine as long as channel head has not changed.
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
			if err := channel.Update(node.Cache(), node.Network(), hash, block); err != nil {
				return nil, nil, err
			}
			return hash, block, nil
		}
	}
	return nil, nil, ErrNonceWrapAround{}
}
