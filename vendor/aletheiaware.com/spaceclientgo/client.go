/*
 * Copyright 2019-2020 Aletheia Ware LLC
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
	"crypto/rsa"
	"encoding/base64"
	"github.com/golang/protobuf/proto"
	"io"
	"log"
)

type SpaceClient struct {
	bcclientgo.BCClient
}

func NewSpaceClient(peers ...string) *SpaceClient {
	if len(peers) == 0 {
		peers = append(
			spacego.GetSpaceHosts(), // Add SPACE host as peer
			bcgo.GetBCHost(),        // Add BC host as peer
		)
	}
	return &SpaceClient{
		BCClient: *bcclientgo.NewBCClient(peers...),
	}
}

func (c *SpaceClient) Init(listener bcgo.MiningListener) (*bcgo.Node, error) {
	root, err := c.GetRoot()
	if err != nil {
		return nil, err
	}

	// Add Space hosts to peers
	for _, host := range spacego.GetSpaceHosts() {
		if err := bcgo.AddPeer(root, host); err != nil {
			return nil, err
		}
	}

	// Add BC host to peers
	if err := bcgo.AddPeer(root, bcgo.GetBCHost()); err != nil {
		return nil, err
	}

	return c.BCClient.Init(listener)
}

// Adds file
func (c *SpaceClient) Add(node *bcgo.Node, listener bcgo.MiningListener, name, mime string, reader io.Reader) (*bcgo.Reference, error) {
	metas := spacego.OpenMetaChannel(node.Alias)
	if err := metas.Refresh(node.Cache, node.Network); err != nil {
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

	acl := map[string]*rsa.PublicKey{
		node.Alias: &node.Key.PublicKey,
	}

	// Write meta data
	reference, err := node.Write(bcgo.Timestamp(), metas, acl, nil, data)
	if err != nil {
		return nil, err
	}

	// Mine meta channel
	if _, _, err := node.Mine(metas, spacego.THRESHOLD, listener); err != nil {
		return nil, err
	}

	if n := node.Network; n != nil {
		// Push update to peers
		if err := metas.Push(node.Cache, n); err != nil {
			log.Println(err)
		}
	}

	metaId := base64.RawURLEncoding.EncodeToString(reference.RecordHash)
	deltas := spacego.OpenDeltaChannel(metaId)
	if err := deltas.Refresh(node.Cache, node.Network); err != nil {
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
		_, record, err := bcgo.CreateRecord(timestamp, node.Alias, node.Key, acl, nil, data)
		if err != nil {
			return err
		}
		if _, err := bcgo.WriteRecord(deltas.Name, node.Cache, record); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}

	// Mine file channel
	if _, _, err := node.Mine(deltas, spacego.THRESHOLD, listener); err != nil {
		return nil, err
	}

	if n := node.Network; n != nil {
		// Push update to peers
		if err := deltas.Push(node.Cache, n); err != nil {
			log.Println(err)
		}
	}

	// TODO Add preview
	return reference, nil
}

// Append adds the given delta to the file
func (c *SpaceClient) Append(node *bcgo.Node, listener bcgo.MiningListener, deltas *bcgo.Channel, acl map[string]*rsa.PublicKey, delta *spacego.Delta) error {
	data, err := proto.Marshal(delta)
	if err != nil {
		return err
	}
	_, record, err := bcgo.CreateRecord(bcgo.Timestamp(), node.Alias, node.Key, acl, nil, data)
	if err != nil {
		return err
	}
	if _, err := bcgo.WriteRecord(deltas.Name, node.Cache, record); err != nil {
		return err
	}
	// Mine file channel
	if _, _, err := node.Mine(deltas, spacego.THRESHOLD, listener); err != nil {
		return err
	}

	if n := node.Network; n != nil {
		// Push update to peers
		if err := deltas.Push(node.Cache, n); err != nil {
			log.Println(err)
		}
	}
	return nil
}

// List files owned by key
func (c *SpaceClient) List(node *bcgo.Node, callback spacego.MetaCallback) error {
	metas := spacego.OpenMetaChannel(node.Alias)
	if err := metas.Refresh(node.Cache, node.Network); err != nil {
		log.Println(err)
	}
	return spacego.GetMeta(metas, node.Cache, node.Network, node.Alias, node.Key, nil, func(entry *bcgo.BlockEntry, meta *spacego.Meta) error {
		return callback(entry, meta)
	})
}

// GetMeta owned by key with given hash
func (c *SpaceClient) GetMeta(node *bcgo.Node, recordHash []byte, callback spacego.MetaCallback) error {
	metas := spacego.OpenMetaChannel(node.Alias)
	if err := metas.Refresh(node.Cache, node.Network); err != nil {
		log.Println(err)
	}
	return spacego.GetMeta(metas, node.Cache, node.Network, node.Alias, node.Key, recordHash, func(entry *bcgo.BlockEntry, meta *spacego.Meta) error {
		return callback(entry, meta)
	})
}

// ReadFile with the given hash
func (c *SpaceClient) ReadFile(node *bcgo.Node, metaId []byte) (io.Reader, error) {
	// TODO read from cache if file exists
	deltas := spacego.OpenDeltaChannel(base64.RawURLEncoding.EncodeToString(metaId))
	if err := deltas.Refresh(node.Cache, node.Network); err != nil {
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

// Search files owned by key
func (c *SpaceClient) Search(node *bcgo.Node, terms []string, callback spacego.MetaCallback) error {
	metas := spacego.OpenMetaChannel(node.Alias)
	if err := metas.Refresh(node.Cache, node.Network); err != nil {
		log.Println(err)
	}
	if err := spacego.GetMeta(metas, node.Cache, node.Network, node.Alias, node.Key, nil, func(metaEntry *bcgo.BlockEntry, meta *spacego.Meta) error {
		tags := spacego.OpenTagChannel(base64.RawURLEncoding.EncodeToString(metaEntry.RecordHash))
		if err := tags.Refresh(node.Cache, node.Network); err != nil {
			log.Println(err)
		}
		return spacego.GetTag(tags, node.Cache, node.Network, node.Alias, node.Key, nil, func(tagEntry *bcgo.BlockEntry, tag *spacego.Tag) error {
			for _, value := range terms {
				if tag.Value == value {
					return callback(metaEntry, meta)
				}
			}
			return nil
		})
	}); err != nil {
		return err
	}
	return nil
}

// AddTag adds the given tag for the file with the given hash
func (c *SpaceClient) AddTag(node *bcgo.Node, listener bcgo.MiningListener, metaId []byte, tag []string) ([]*bcgo.Reference, error) {
	metas := spacego.OpenMetaChannel(node.Alias)
	if err := metas.Refresh(node.Cache, node.Network); err != nil {
		log.Println(err)
	}
	tags := spacego.OpenTagChannel(base64.RawURLEncoding.EncodeToString(metaId))
	if err := tags.Refresh(node.Cache, node.Network); err != nil {
		log.Println(err)
	}
	var references []*bcgo.Reference
	if err := spacego.GetMeta(metas, node.Cache, node.Network, node.Alias, node.Key, metaId, func(entry *bcgo.BlockEntry, meta *spacego.Meta) error {
		for _, t := range tag {
			tag := spacego.Tag{
				Value: t,
			}
			data, err := proto.Marshal(&tag)
			if err != nil {
				return err
			}
			acl := map[string]*rsa.PublicKey{
				node.Alias: &node.Key.PublicKey,
			}
			references := []*bcgo.Reference{&bcgo.Reference{
				Timestamp:   entry.Record.Timestamp,
				ChannelName: metas.Name,
				RecordHash:  metaId,
			}}
			reference, err := node.Write(bcgo.Timestamp(), tags, acl, references, data)
			if err != nil {
				return err
			}
			references = append(references, reference)
			if _, _, err := node.Mine(tags, spacego.THRESHOLD, listener); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return references, nil
}

// GetTag returns all tags for the file with the given hash
func (c *SpaceClient) GetTag(node *bcgo.Node, metaId []byte, callback func(entry *bcgo.BlockEntry, tag *spacego.Tag)) error {
	tags := spacego.OpenTagChannel(base64.RawURLEncoding.EncodeToString(metaId))
	if err := tags.Refresh(node.Cache, node.Network); err != nil {
		log.Println(err)
	}
	return spacego.GetTag(tags, node.Cache, node.Network, node.Alias, node.Key, nil, func(entry *bcgo.BlockEntry, tag *spacego.Tag) error {
		for _, reference := range entry.Record.Reference {
			if bytes.Equal(metaId, reference.RecordHash) {
				callback(entry, tag)
			}
		}
		return nil
	})
}

// GetRegistration triggers the given callback for the most recent registration with the given merchant.
func (c *SpaceClient) GetRegistration(merchant string, callback financego.RegistrationCallback) error {
	node, err := c.GetNode()
	if err != nil {
		return err
	}
	registrations := spacego.OpenRegistrationChannel()
	if err := registrations.Refresh(node.Cache, node.Network); err != nil {
		log.Println(err)
	}
	return financego.GetRegistrationAsync(registrations, node.Cache, node.Network, merchant, nil, node.Alias, node.Key, callback)
}

// GetSubscription triggers the given callback for the most recent subscription with the given merchant.
func (c *SpaceClient) GetSubscription(merchant string, callback financego.SubscriptionCallback) error {
	node, err := c.GetNode()
	if err != nil {
		return err
	}
	subscriptions := spacego.OpenSubscriptionChannel()
	if err := subscriptions.Refresh(node.Cache, node.Network); err != nil {
		log.Println(err)
	}
	return financego.GetSubscriptionAsync(subscriptions, node.Cache, node.Network, merchant, nil, node.Alias, node.Key, "", "", callback)
}
