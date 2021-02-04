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
	"encoding/base64"
	"fmt"
	"github.com/golang/protobuf/proto"
	"log"
	"strings"
	"time"
)

// Periodic Validation Chains strengthen the Network by increasing the computational resources needed to attack it.

const (
	ERROR_MISSING_VALIDATED_BLOCK = "Missing Validated Block %s"

	PERIOD_HOURLY       = time.Hour
	PERIOD_DAILY        = PERIOD_HOURLY * 24
	PERIOD_WEEKLY       = PERIOD_HOURLY * 168    // (24 * 7)
	PERIOD_YEARLY       = PERIOD_HOURLY * 8766   // (24 * 365.25)
	PERIOD_DECENNIALLY  = PERIOD_HOURLY * 87660  // (24 * 365.25 * 10)
	PERIOD_CENTENNIALLY = PERIOD_HOURLY * 876600 // (24 * 365.25 * 100)

	THRESHOLD_PERIOD_HOUR    = THRESHOLD_F
	THRESHOLD_PERIOD_DAY     = THRESHOLD_E
	THRESHOLD_PERIOD_WEEK    = THRESHOLD_D
	THRESHOLD_PERIOD_YEAR    = THRESHOLD_C
	THRESHOLD_PERIOD_DECADE  = THRESHOLD_B
	THRESHOLD_PERIOD_CENTURY = THRESHOLD_A
)

type PeriodicValidator struct {
	// TODO add validator that each block holds the full channel set of the previous
	// TODO add validator that the duration between block timestamps equals or exceeds the period
	// TODO add validator that each head reference in block is the longest chain before timestamp
	Channel *Channel
	Period  time.Duration
	Ticker  *time.Ticker
}

func NewPeriodicValidator(channel *Channel, period time.Duration) *PeriodicValidator {
	return &PeriodicValidator{
		Channel: channel,
		Period:  period,
	}
}

func GetHourlyValidator(channel *Channel) *PeriodicValidator {
	return NewPeriodicValidator(channel, PERIOD_HOURLY)
}

func GetDailyValidator(channel *Channel) *PeriodicValidator {
	return NewPeriodicValidator(channel, PERIOD_DAILY)
}

func GetWeeklyValidator(channel *Channel) *PeriodicValidator {
	return NewPeriodicValidator(channel, PERIOD_WEEKLY)
}

func GetYearlyValidator(channel *Channel) *PeriodicValidator {
	return NewPeriodicValidator(channel, PERIOD_YEARLY)
}

func GetDecenniallyValidator(channel *Channel) *PeriodicValidator {
	return NewPeriodicValidator(channel, PERIOD_DECENNIALLY)
}

func GetCentenniallyValidator(channel *Channel) *PeriodicValidator {
	return NewPeriodicValidator(channel, PERIOD_CENTENNIALLY)
}

// Fills the given set with the names of all channels validated in this chain
func (p *PeriodicValidator) FillChannelSet(set map[string]bool, cache Cache, network Network) error {
	return Iterate(p.Channel.Name, p.Channel.Head, nil, cache, network, func(h []byte, b *Block) error {
		for _, entry := range b.Entry {
			// Unmarshal as Reference
			r := &Reference{}
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
func (p *PeriodicValidator) Validate(channel *Channel, cache Cache, network Network, hash []byte, block *Block) error {
	// Mark all block hashes for channel in p.Channel
	set := make(map[string]bool)
	if err := Iterate(p.Channel.Name, p.Channel.Head, nil, cache, network, func(h []byte, b *Block) error {
		for _, entry := range b.Entry {
			// Unmarshal as Reference
			r := &Reference{}
			err := proto.Unmarshal(entry.Record.Payload, r)
			if err != nil {
				return err
			}
			if r.ChannelName == channel.Name {
				set[base64.RawURLEncoding.EncodeToString(r.BlockHash)] = true
			}
		}
		return nil
	}); err != nil {
		return err
	}

	// Unmark all block hashes which appear is chain
	if err := Iterate(channel.Name, hash, block, cache, network, func(h []byte, b *Block) error {
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
		return fmt.Errorf(ERROR_MISSING_VALIDATED_BLOCK, strings.Join(missing, ","))
	}
	return nil
}

func (p *PeriodicValidator) Update(node *Node, threshold uint64, listener MiningListener, timestamp uint64) error {
	entries, err := CreateValidationEntries(timestamp, node)
	if err != nil {
		return err
	}
	name := p.Channel.Name
	head := p.Channel.Head
	var block *Block
	if head != nil {
		block, err = GetBlock(name, node.Cache, node.Network, head)
		if err != nil {
			return err
		}
	}
	b := CreateValidationBlock(timestamp, name, node.Alias, head, block, entries)
	_, _, err = node.MineBlock(p.Channel, threshold, listener, b)
	if err != nil {
		return err
	}
	return nil
}

// Periodically mines a new block into the chain containing the head hashes of all open channels
func (p *PeriodicValidator) Start(node *Node, threshold uint64, listener MiningListener) {
	// 3 times per period
	p.Ticker = time.NewTicker(p.Period / 3)
	c := p.Ticker.C
	for {
		go func() {
			for {
				now := time.Now().UTC()
				last := int64(p.Channel.Timestamp)
				next := time.Unix(0, last).Add(p.Period)
				var timestamp uint64
				if last == 0 {
					timestamp = uint64(now.UnixNano())
				} else if now.After(next) {
					timestamp = uint64(next.UnixNano())
				} else {
					break
				}
				if err := p.Update(node, threshold, listener, timestamp); err != nil {
					log.Println(err)
					p.Stop()
					break
				}
				if err := p.Channel.Push(node.Cache, node.Network); err != nil {
					log.Println(err)
				}
			}
		}()
		if _, ok := <-c; !ok {
			return
		}
	}
}

func (p *PeriodicValidator) Stop() {
	if p.Ticker != nil {
		p.Ticker.Stop()
		p.Ticker = nil
	}
}

func CreateValidationBlock(timestamp uint64, channel, alias string, head []byte, block *Block, entries []*BlockEntry) *Block {
	b := &Block{
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

func CreateValidationEntries(timestamp uint64, node *Node) ([]*BlockEntry, error) {
	var entries []*BlockEntry
	for _, channel := range node.GetChannels() {
		head := channel.Head
		updated := channel.Timestamp
		if updated > timestamp {
			// Channel was updated after Validation Cycle started
			head = nil
			// Iterate back through channel blocks until block.Timestamp <= timestamp
			if err := Iterate(channel.Name, channel.Head, nil, node.Cache, node.Network, func(h []byte, b *Block) error {
				if b.Timestamp <= timestamp {
					head = h
					updated = b.Timestamp
					return StopIterationError{}
				}
				return nil
			}); err != nil {
				switch err.(type) {
				case StopIterationError:
					// Do nothing
				default:
					return nil, err
				}
			}
		}
		if head == nil {
			continue
		}
		entry, err := CreateValidationEntry(timestamp, node, channel.Name, updated, head)
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	return entries, nil
}

func CreateValidationEntry(timestamp uint64, node *Node, channel string, updated uint64, head []byte) (*BlockEntry, error) {
	reference := &Reference{
		Timestamp:   updated,
		ChannelName: channel,
		BlockHash:   head,
	}
	payload, err := proto.Marshal(reference)
	if err != nil {
		return nil, err
	}
	_, record, err := CreateRecord(timestamp, node.Alias, node.Key, nil, nil, payload)
	if err != nil {
		return nil, err
	}
	hash, err := cryptogo.HashProtobuf(record)
	if err != nil {
		return nil, err
	}
	return &BlockEntry{
		RecordHash: hash,
		Record:     record,
	}, nil
}
