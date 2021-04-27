/*
 * Copyright 2019-2021 Aletheia Ware LLC
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

package spaceclientgo

import (
	"aletheiaware.com/bcclientgo"
	"aletheiaware.com/bcgo"
	"aletheiaware.com/financego"
	"aletheiaware.com/spacego"
	"bytes"
	"context"
	"encoding/base64"
	"github.com/golang/protobuf/proto"
	"io"
	"log"
	"reflect"
	"time"
)

type SpaceClient interface {
	bcclientgo.BCClient

	Add(bcgo.Node, bcgo.MiningListener, string, string, io.Reader) (*bcgo.Reference, error)
	Amend(bcgo.Node, bcgo.MiningListener, bcgo.Channel, ...*spacego.Delta) error
	MetaForHash(bcgo.Node, []byte, spacego.MetaCallback) error
	AllMetas(bcgo.Node, spacego.MetaCallback) error
	ReadFile(bcgo.Node, []byte) (io.Reader, error)
	WriteFile(bcgo.Node, bcgo.MiningListener, []byte) (io.WriteCloser, error)
	WatchFile(context.Context, bcgo.Node, []byte, func())

	/*
		AddPreview(bcgo.Node, bcgo.MiningListener, []byte, []string) ([]*bcgo.Reference, error)
		AllPreviewsForHash(bcgo.Node, []byte, spacego.PreviewCallback) error
	*/

	AddTag(bcgo.Node, bcgo.MiningListener, []byte, []string) ([]*bcgo.Reference, error)
	AllTagsForHash(bcgo.Node, []byte, spacego.TagCallback) error

	SearchMeta(bcgo.Node, spacego.MetaFilter, spacego.MetaCallback) error
	SearchTag(bcgo.Node, spacego.TagFilter, spacego.MetaCallback) error

	Registration(string, financego.RegistrationCallback) error
	Subscription(string, financego.SubscriptionCallback) error
}

type spaceClient struct {
	bcclientgo.BCClient
}

func NewSpaceClient(peers ...string) SpaceClient {
	if len(peers) == 0 {
		peers = append(
			spacego.SpaceHosts(), // Add SPACE host as peer
			bcgo.BCHost(),        // Add BC host as peer
		)
	}
	return &spaceClient{
		BCClient: bcclientgo.NewBCClient(peers...),
	}
}

// Adds file
func (c *spaceClient) Add(node bcgo.Node, listener bcgo.MiningListener, name, mime string, reader io.Reader) (*bcgo.Reference, error) {
	account := node.Account()
	alias := account.Alias()
	metas := node.OpenChannel(spacego.MetaChannelName(alias), func() bcgo.Channel {
		return spacego.OpenMetaChannel(alias)
	})
	if err := metas.Refresh(node.Cache(), node.Network()); err != nil {
		log.Println(err)
	}

	// Create Meta
	meta := spacego.Meta{
		Name: name,
		Type: mime,
	}

	data, err := proto.Marshal(&meta)
	if err != nil {
		return nil, err
	}

	// Write meta data
	reference, err := node.Write(bcgo.Timestamp(), metas, []bcgo.Identity{account}, nil, data)
	if err != nil {
		return nil, err
	}

	// Mine meta channel
	if _, _, err := bcgo.Mine(node, metas, spacego.THRESHOLD_CUSTOMER, listener); err != nil {
		return nil, err
	}

	if n := node.Network(); n != nil && !reflect.ValueOf(n).IsNil() {
		// Push update to peers
		if err := metas.Push(node.Cache(), n); err != nil {
			log.Println(err)
		}
	}

	if reader == nil {
		return reference, nil
	}

	metaId := base64.RawURLEncoding.EncodeToString(reference.RecordHash)

	deltas := node.OpenChannel(spacego.DeltaChannelName(metaId), func() bcgo.Channel {
		return spacego.OpenDeltaChannel(metaId)
	})
	if err := deltas.Refresh(node.Cache(), node.Network()); err != nil {
		log.Println(err)
	}

	// TODO compress data

	var last uint64
	// Read data, create deltas, and write to cache
	if err := spacego.CreateDeltas(reader, spacego.MAX_SIZE_BYTES, func(delta *spacego.Delta) error {
		data, err := proto.Marshal(delta)
		if err != nil {
			return err
		}
		timestamp := bcgo.Timestamp()
		// Ensure timestamp is greater than previous to ensure deltas (sorted by timestamp) don't get out of order
		for last == timestamp {
			timestamp = bcgo.Timestamp()
		}
		last = timestamp
		_, record, err := bcgo.CreateRecord(timestamp, account, []bcgo.Identity{account}, nil, data)
		if err != nil {
			return err
		}
		if _, err := bcgo.WriteRecord(deltas.Name(), node.Cache(), record); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}

	// Mine file channel
	if _, _, err := bcgo.Mine(node, deltas, spacego.THRESHOLD_CUSTOMER, listener); err != nil {
		return nil, err
	}

	if n := node.Network(); n != nil && !reflect.ValueOf(n).IsNil() {
		// Push update to peers
		if err := deltas.Push(node.Cache(), n); err != nil {
			log.Println(err)
		}
	}

	// TODO Add preview
	return reference, nil
}

// Amend adds the given delta to the file
func (c *spaceClient) Amend(node bcgo.Node, listener bcgo.MiningListener, channel bcgo.Channel, deltas ...*spacego.Delta) error {
	if len(deltas) == 0 {
		return nil
	}
	account := node.Account()
	access := []bcgo.Identity{account}
	name := channel.Name()
	cache := node.Cache()
	for _, d := range deltas {
		data, err := proto.Marshal(d)
		if err != nil {
			return err
		}

		_, record, err := bcgo.CreateRecord(bcgo.Timestamp(), account, access, nil, data)
		if err != nil {
			return err
		}

		if _, err := bcgo.WriteRecord(name, cache, record); err != nil {
			return err
		}
	}

	// Mine file channel
	if _, _, err := bcgo.Mine(node, channel, spacego.THRESHOLD_CUSTOMER, listener); err != nil {
		return err
	}

	if n := node.Network(); n != nil && !reflect.ValueOf(n).IsNil() {
		// Push update to peers
		if err := channel.Push(node.Cache(), n); err != nil {
			log.Println(err)
		}
	}
	return nil
}

// MetaForHash owned by key with given meta ID
func (c *spaceClient) MetaForHash(node bcgo.Node, recordHash []byte, callback spacego.MetaCallback) error {
	alias := node.Account().Alias()
	metas := node.OpenChannel(spacego.MetaChannelName(alias), func() bcgo.Channel {
		return spacego.OpenMetaChannel(alias)
	})
	if err := metas.Refresh(node.Cache(), node.Network()); err != nil {
		log.Println(err)
	}
	return spacego.ReadMeta(metas, node.Cache(), node.Network(), node.Account(), recordHash, func(entry *bcgo.BlockEntry, meta *spacego.Meta) error {
		return callback(entry, meta)
	})
}

// AllMetas lists files owned by key
func (c *spaceClient) AllMetas(node bcgo.Node, callback spacego.MetaCallback) error {
	alias := node.Account().Alias()
	metas := node.OpenChannel(spacego.MetaChannelName(alias), func() bcgo.Channel {
		return spacego.OpenMetaChannel(alias)
	})
	if err := metas.Refresh(node.Cache(), node.Network()); err != nil {
		log.Println(err)
	}
	return spacego.ReadMeta(metas, node.Cache(), node.Network(), node.Account(), nil, func(entry *bcgo.BlockEntry, meta *spacego.Meta) error {
		return callback(entry, meta)
	})
}

// ReadFile with the given meta ID.
func (c *spaceClient) ReadFile(node bcgo.Node, metaId []byte) (io.Reader, error) {
	// TODO read from cache if file exists
	mId := base64.RawURLEncoding.EncodeToString(metaId)
	deltas := node.OpenChannel(spacego.DeltaChannelName(mId), func() bcgo.Channel {
		return spacego.OpenDeltaChannel(mId)
	})
	if err := deltas.Refresh(node.Cache(), node.Network()); err != nil {
		log.Println(err)
	}
	buffer := []byte{}
	if err := spacego.IterateDeltas(node, deltas, func(entry *bcgo.BlockEntry, delta *spacego.Delta) error {
		buffer = spacego.ApplyDelta(delta, buffer)
		return nil
	}); err != nil {
		return nil, err
	}
	return bytes.NewReader(buffer), nil
}

// WriteFile with the given meta ID.
func (c *spaceClient) WriteFile(node bcgo.Node, listener bcgo.MiningListener, metaId []byte) (io.WriteCloser, error) {
	mId := base64.RawURLEncoding.EncodeToString(metaId)
	deltas := node.OpenChannel(spacego.DeltaChannelName(mId), func() bcgo.Channel {
		return spacego.OpenDeltaChannel(mId)
	})
	if err := deltas.Refresh(node.Cache(), node.Network()); err != nil {
		log.Println(err)
	}
	// Read current file into a old buffer
	old := []byte{}
	if err := spacego.IterateDeltas(node, deltas, func(entry *bcgo.BlockEntry, delta *spacego.Delta) error {
		old = spacego.ApplyDelta(delta, old)
		return nil
	}); err != nil {
		return nil, err
	}
	var new bytes.Buffer
	return spacego.NewCloser(&new, func() error {
		return c.Amend(node, listener, deltas, spacego.Difference(old, new.Bytes())...)
	}), nil
}

// WatchFile triggers the given callback whenever the file with given meta ID updates.
func (c *spaceClient) WatchFile(ctx context.Context, node bcgo.Node, metaId []byte, callback func()) {
	initial := time.Second
	limit := time.Hour
	duration := initial
	ticker := time.NewTicker(duration)
	mId := base64.RawURLEncoding.EncodeToString(metaId)
	deltas := node.OpenChannel(spacego.DeltaChannelName(mId), func() bcgo.Channel {
		return spacego.OpenDeltaChannel(mId)
	})
	deltas.AddTrigger(func() {
		if ctx.Err() != nil {
			// Context was already cancelled
			return
		}
		go callback()
	})
	// TODO replace polling with mechanism to register with provider to listen for remote updates
	go func() {
		defer ticker.Stop()
		var errors int
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				head := deltas.Head()
				if err := deltas.Refresh(node.Cache(), node.Network()); err != nil {
					log.Println(err)
				}
				if bytes.Equal(head, deltas.Head()) {
					// No change
					errors++
					if errors > 3 {
						// Too many errors, exponential backoff
						duration *= 2
						if duration > limit {
							duration = limit
						}
						errors = 0
					}
				} else {
					// Change, reset duration
					duration = initial
					errors = 0
				}
				ticker.Stop()
				ticker = time.NewTicker(duration)
			}
		}
	}()
}

// SearchMeta searches files by metadata
func (c *spaceClient) SearchMeta(node bcgo.Node, filter spacego.MetaFilter, callback spacego.MetaCallback) error {
	account := node.Account()
	alias := account.Alias()
	metas := node.OpenChannel(spacego.MetaChannelName(alias), func() bcgo.Channel {
		return spacego.OpenMetaChannel(alias)
	})
	if err := metas.Refresh(node.Cache(), node.Network()); err != nil {
		log.Println(err)
	}
	if err := spacego.ReadMeta(metas, node.Cache(), node.Network(), account, nil, func(metaEntry *bcgo.BlockEntry, meta *spacego.Meta) error {
		if filter != nil && !filter.Filter(meta) {
			// Meta doesn't pass filter
			return nil
		}
		return callback(metaEntry, meta)
	}); err != nil {
		return err
	}
	return nil
}

// SearchTag searches files by tag
func (c *spaceClient) SearchTag(node bcgo.Node, filter spacego.TagFilter, callback spacego.MetaCallback) error {
	return c.SearchMeta(node, nil, func(metaEntry *bcgo.BlockEntry, meta *spacego.Meta) error {
		metaId := base64.RawURLEncoding.EncodeToString(metaEntry.RecordHash)
		tags := node.OpenChannel(spacego.TagChannelName(metaId), func() bcgo.Channel {
			return spacego.OpenTagChannel(metaId)
		})
		if err := tags.Refresh(node.Cache(), node.Network()); err != nil {
			log.Println(err)
		}
		return spacego.ReadTag(tags, node.Cache(), node.Network(), node.Account(), nil, func(tagEntry *bcgo.BlockEntry, tag *spacego.Tag) error {
			if filter != nil && !filter.Filter(tag) {
				// Tag doesn't pass filter
				return nil
			}
			return callback(metaEntry, meta)
		})
	})
}

// AddTag adds the given tag for the file with the given meta ID
func (c *spaceClient) AddTag(node bcgo.Node, listener bcgo.MiningListener, metaId []byte, tag []string) ([]*bcgo.Reference, error) {
	account := node.Account()
	alias := account.Alias()
	metas := node.OpenChannel(spacego.MetaChannelName(alias), func() bcgo.Channel {
		return spacego.OpenMetaChannel(alias)
	})
	if err := metas.Refresh(node.Cache(), node.Network()); err != nil {
		log.Println(err)
	}
	mId := base64.RawURLEncoding.EncodeToString(metaId)
	tags := node.OpenChannel(spacego.TagChannelName(mId), func() bcgo.Channel {
		return spacego.OpenTagChannel(mId)
	})
	if err := tags.Refresh(node.Cache(), node.Network()); err != nil {
		log.Println(err)
	}
	var references []*bcgo.Reference
	if err := spacego.ReadMeta(metas, node.Cache(), node.Network(), account, metaId, func(entry *bcgo.BlockEntry, meta *spacego.Meta) error {
		for _, t := range tag {
			tag := spacego.Tag{
				Value: t,
			}
			data, err := proto.Marshal(&tag)
			if err != nil {
				return err
			}
			references := []*bcgo.Reference{&bcgo.Reference{
				Timestamp:   entry.Record.Timestamp,
				ChannelName: metas.Name(),
				RecordHash:  metaId,
			}}
			reference, err := node.Write(bcgo.Timestamp(), tags, []bcgo.Identity{account}, references, data)
			if err != nil {
				return err
			}
			references = append(references, reference)
			if _, _, err := bcgo.Mine(node, tags, spacego.THRESHOLD_CUSTOMER, listener); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return references, nil
}

// AllTagsForHash lists all tags for the file with the given meta ID
func (c *spaceClient) AllTagsForHash(node bcgo.Node, metaId []byte, callback spacego.TagCallback) error {
	mId := base64.RawURLEncoding.EncodeToString(metaId)
	tags := node.OpenChannel(spacego.TagChannelName(mId), func() bcgo.Channel {
		return spacego.OpenTagChannel(mId)
	})
	if err := tags.Refresh(node.Cache(), node.Network()); err != nil {
		log.Println(err)
	}
	return spacego.ReadTag(tags, node.Cache(), node.Network(), node.Account(), nil, func(entry *bcgo.BlockEntry, tag *spacego.Tag) error {
		for _, reference := range entry.Record.Reference {
			if bytes.Equal(metaId, reference.RecordHash) {
				callback(entry, tag)
			}
		}
		return nil
	})
}

// Registration triggers the given callback for the most recent registration with the given merchant.
func (c *spaceClient) Registration(merchant string, callback financego.RegistrationCallback) error {
	node, err := c.Node()
	if err != nil {
		return err
	}
	registrations := node.OpenChannel(spacego.SPACE_REGISTRATION, func() bcgo.Channel {
		return spacego.OpenRegistrationChannel()
	})
	if err := registrations.Refresh(node.Cache(), node.Network()); err != nil {
		log.Println(err)
	}
	return financego.RegistrationAsync(registrations, node.Cache(), node.Network(), node.Account(), merchant, node.Account().Alias(), callback)
}

// Subscription triggers the given callback for the most recent subscription with the given merchant.
func (c *spaceClient) Subscription(merchant string, callback financego.SubscriptionCallback) error {
	node, err := c.Node()
	if err != nil {
		return err
	}
	subscriptions := node.OpenChannel(spacego.SPACE_SUBSCRIPTION, func() bcgo.Channel {
		return spacego.OpenSubscriptionChannel()
	})
	if err := subscriptions.Refresh(node.Cache(), node.Network()); err != nil {
		log.Println(err)
	}
	return financego.SubscriptionAsync(subscriptions, node.Cache(), node.Network(), node.Account(), merchant, node.Account().Alias(), "", "", callback)
}
