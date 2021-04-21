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
	"aletheiaware.com/cryptogo"
	"encoding/base64"
	"github.com/golang/protobuf/proto"
	"log"
	"time"
)

// Periodic Validation Chains strengthen the Network by increasing the computational resources needed to attack it.

const (
	PERIOD_HOURLY       = time.Hour
	PERIOD_DAILY        = PERIOD_HOURLY * 24
	PERIOD_WEEKLY       = PERIOD_HOURLY * 168    // (24 * 7)
	PERIOD_YEARLY       = PERIOD_HOURLY * 8766   // (24 * 365.25)
	PERIOD_DECENNIALLY  = PERIOD_HOURLY * 87660  // (24 * 365.25 * 10)
	PERIOD_CENTENNIALLY = PERIOD_HOURLY * 876600 // (24 * 365.25 * 100)

	THRESHOLD_PERIOD_HOUR    = bcgo.THRESHOLD_F
	THRESHOLD_PERIOD_DAY     = bcgo.THRESHOLD_E
	THRESHOLD_PERIOD_WEEK    = bcgo.THRESHOLD_D
	THRESHOLD_PERIOD_YEAR    = bcgo.THRESHOLD_C
	THRESHOLD_PERIOD_DECADE  = bcgo.THRESHOLD_B
	THRESHOLD_PERIOD_CENTURY = bcgo.THRESHOLD_A
)

type Periodic interface {
	bcgo.Validator
	AddChannel(string)
	FillChannelSet(map[string]bool, bcgo.Cache, bcgo.Network) error
	Update(bcgo.Cache, bcgo.Network, uint64) error
	Start()
	Stop()
}

type periodic struct {
	// TODO add validator that each block holds the full channel set of the previous
	// TODO add validator that the duration between block timestamps equals or exceeds the period
	// TODO add validator that each head reference in block is the longest chain before timestamp
	Node      bcgo.Node
	Channel   bcgo.Channel
	Threshold uint64
	Listener  bcgo.MiningListener
	Period    time.Duration
	Ticker    *time.Ticker
	Channels  map[string]bool
}

func NewPeriodic(node bcgo.Node, channel bcgo.Channel, threshold uint64, listener bcgo.MiningListener, period time.Duration) Periodic {
	p := &periodic{
		Node:      node,
		Channel:   channel,
		Threshold: threshold,
		Listener:  listener,
		Period:    period,
		Channels:  make(map[string]bool),
	}
	p.Channel.AddTrigger(p.updateValidatedChannels)
	return p
}

func NewHourly(node bcgo.Node, channel bcgo.Channel, listener bcgo.MiningListener) Periodic {
	return NewPeriodic(node, channel, bcgo.THRESHOLD_F, listener, PERIOD_HOURLY)
}

func NewDaily(node bcgo.Node, channel bcgo.Channel, listener bcgo.MiningListener) Periodic {
	return NewPeriodic(node, channel, bcgo.THRESHOLD_E, listener, PERIOD_DAILY)
}

func NewWeekly(node bcgo.Node, channel bcgo.Channel, listener bcgo.MiningListener) Periodic {
	return NewPeriodic(node, channel, bcgo.THRESHOLD_D, listener, PERIOD_WEEKLY)
}

func NewYearly(node bcgo.Node, channel bcgo.Channel, listener bcgo.MiningListener) Periodic {
	return NewPeriodic(node, channel, bcgo.THRESHOLD_C, listener, PERIOD_YEARLY)
}

func NewDecennially(node bcgo.Node, channel bcgo.Channel, listener bcgo.MiningListener) Periodic {
	return NewPeriodic(node, channel, bcgo.THRESHOLD_B, listener, PERIOD_DECENNIALLY)
}

func NewCentennially(node bcgo.Node, channel bcgo.Channel, listener bcgo.MiningListener) Periodic {
	return NewPeriodic(node, channel, bcgo.THRESHOLD_A, listener, PERIOD_CENTENNIALLY)
}

func (p *periodic) AddChannel(channel string) {
	p.Channels[channel] = true
}

// Fills the given set with the names of all channels validated in this chain
func (p *periodic) FillChannelSet(set map[string]bool, cache bcgo.Cache, network bcgo.Network) error {
	return bcgo.Iterate(p.Channel.Name(), p.Channel.Head(), nil, cache, network, func(h []byte, b *bcgo.Block) error {
		for _, entry := range b.Entry {
			// Unmarshal as Reference
			r := &bcgo.Reference{}
			err := proto.Unmarshal(entry.Record.Payload, r)
			if err != nil {
				return err
			}
			set[r.ChannelName] = true
		}
		return nil
	})
}

// Ensures all block hashes in validation chain for given channel appear in its chain
func (p *periodic) Validate(channel bcgo.Channel, cache bcgo.Cache, network bcgo.Network, hash []byte, block *bcgo.Block) error {
	name := channel.Name()
	// Mark all block hashes for channel in p.Channel
	set := make(map[string]bool)
	if err := bcgo.Iterate(p.Channel.Name(), p.Channel.Head(), nil, cache, network, func(h []byte, b *bcgo.Block) error {
		for _, entry := range b.Entry {
			// Unmarshal as Reference
			r := &bcgo.Reference{}
			err := proto.Unmarshal(entry.Record.Payload, r)
			if err != nil {
				return err
			}
			if r.ChannelName == name {
				set[base64.RawURLEncoding.EncodeToString(r.BlockHash)] = true
			}
		}
		return nil
	}); err != nil {
		return err
	}

	// Unmark all block hashes which appear is chain
	if err := bcgo.Iterate(name, hash, block, cache, network, func(h []byte, b *bcgo.Block) error {
		set[base64.RawURLEncoding.EncodeToString(h)] = false
		return nil
	}); err != nil {
		return err
	}

	// Collect all marked block hashes
	var missing []string
	for hash, marked := range set {
		if marked {
			missing = append(missing, hash)
		}
	}
	if len(missing) > 0 {
		return ErrMissingValidatedBlock{PVC: p.Channel.Name(), Channel: name, Missing: missing}
	}
	return nil
}

func (p *periodic) Update(cache bcgo.Cache, network bcgo.Network, timestamp uint64) error {
	entries, err := CreateValidationEntries(timestamp, p.Node, p.Channels)
	if err != nil {
		return err
	}
	name := p.Channel.Name()
	head := p.Channel.Head()
	var block *bcgo.Block
	if head != nil {
		block, err = bcgo.LoadBlock(name, cache, network, head)
		if err != nil {
			return err
		}
	}
	b := CreateValidationBlock(timestamp, name, p.Node.Account().Alias(), head, block, entries)
	_, _, err = bcgo.MineBlock(p.Node, p.Channel, p.Threshold, p.Listener, b)
	if err != nil {
		return err
	}
	return nil
}

// Periodically mines a new block into the chain containing the head hashes of all open channels
func (p *periodic) Start() {
	cache := p.Node.Cache()
	network := p.Node.Network()
	// 3 times per period
	p.Ticker = time.NewTicker(p.Period / 3)
	c := p.Ticker.C
	for {
		go func() {
			for {
				now := time.Now().UTC()
				last := int64(p.Channel.Timestamp())
				next := time.Unix(0, last).Add(p.Period)
				var timestamp uint64
				if last == 0 {
					timestamp = uint64(now.UnixNano())
				} else if now.After(next) {
					timestamp = uint64(next.UnixNano())
				} else {
					break
				}
				if err := p.Update(cache, network, timestamp); err != nil {
					log.Println(err)
					break
				}
				if err := p.Channel.Push(cache, network); err != nil {
					log.Println(err)
				}
			}
		}()
		if _, ok := <-c; !ok {
			return
		}
	}
}

func (p *periodic) Stop() {
	if p.Ticker != nil {
		p.Ticker.Stop()
		p.Ticker = nil
	}
}

func (p *periodic) updateValidatedChannels() {
	cache := p.Node.Cache()
	network := p.Node.Network()
	// Try update all open channels in node to the head listed in the latest validator block
	block, err := bcgo.LoadBlock(p.Channel.Name(), cache, network, p.Channel.Head())
	if err != nil {
		log.Println(err)
		return
	}
	for _, entry := range block.Entry {
		// Unmarshal as Reference
		r := &bcgo.Reference{}
		err := proto.Unmarshal(entry.Record.Payload, r)
		if err != nil {
			log.Println(err)
			continue
		}
		c, err := p.Node.Channel(r.ChannelName)
		if err != nil {
			log.Println(err)
			continue
		}
		b, err := bcgo.LoadBlock(r.ChannelName, cache, network, r.BlockHash)
		if err == nil {
			err = c.Update(cache, network, r.BlockHash, b)
		}
		if err != nil {
			log.Println(err)
			err = c.Pull(cache, network)
			if err != nil {
				log.Println(err)
			}
		}
		// TODO re-mine and push channel
	}
}

func CreateValidationBlock(timestamp uint64, channel, alias string, head []byte, block *bcgo.Block, entries []*bcgo.BlockEntry) *bcgo.Block {
	b := &bcgo.Block{
		Timestamp:   timestamp,
		ChannelName: channel,
		Length:      1,
		Miner:       alias,
		Entry:       entries,
	}

	if head != nil && block != nil {
		b.Length = block.Length + 1
		b.Previous = head
	}

	return b
}

func CreateValidationEntries(timestamp uint64, node bcgo.Node, channels map[string]bool) ([]*bcgo.BlockEntry, error) {
	var entries []*bcgo.BlockEntry
	for _, channel := range node.Channels() {
		name := channel.Name()
		if b, ok := channels[name]; !ok || !b {
			continue
		}
		head := channel.Head()
		updated := channel.Timestamp()
		if updated > timestamp {
			// Channel was updated after Validation Cycle started
			h := head
			head = nil
			// Iterate back through channel blocks until block.Timestamp <= timestamp
			if err := bcgo.Iterate(name, h, nil, node.Cache(), node.Network(), func(h []byte, b *bcgo.Block) error {
				if b.Timestamp <= timestamp {
					head = h
					updated = b.Timestamp
					return bcgo.ErrStopIteration{}
				}
				return nil
			}); err != nil {
				switch err.(type) {
				case bcgo.ErrStopIteration:
					// Do nothing
				default:
					return nil, err
				}
			}
		}
		if head == nil {
			continue
		}
		entry, err := CreateValidationEntry(timestamp, node, name, updated, head)
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	return entries, nil
}

func CreateValidationEntry(timestamp uint64, node bcgo.Node, channel string, updated uint64, head []byte) (*bcgo.BlockEntry, error) {
	reference := &bcgo.Reference{
		Timestamp:   updated,
		ChannelName: channel,
		BlockHash:   head,
	}
	payload, err := proto.Marshal(reference)
	if err != nil {
		return nil, err
	}
	_, record, err := bcgo.CreateRecord(timestamp, node.Account(), nil, nil, payload)
	if err != nil {
		return nil, err
	}
	hash, err := cryptogo.HashProtobuf(record)
	if err != nil {
		return nil, err
	}
	return &bcgo.BlockEntry{
		RecordHash: hash,
		Record:     record,
	}, nil
}
