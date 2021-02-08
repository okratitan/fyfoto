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

package spacego

import (
	"aletheiaware.com/bcgo"
	"aletheiaware.com/financego"
	"crypto/rsa"
	"github.com/golang/protobuf/proto"
	"io"
)

const (
	SPACE              = "S P A C E"
	SPACE_HOUR         = "Space-Hour"
	SPACE_DAY          = "Space-Day"
	SPACE_YEAR         = "Space-Year"
	SPACE_CHARGE       = "Space-Charge"
	SPACE_INVOICE      = "Space-Invoice"
	SPACE_REGISTRAR    = "Space-Registrar"
	SPACE_REGISTRATION = "Space-Registration"
	SPACE_SUBSCRIPTION = "Space-Subscription"
	SPACE_USAGE_RECORD = "Space-Usage-Record"

	SPACE_PREFIX         = "Space-"
	SPACE_PREFIX_DELTA   = "Space-Delta-"
	SPACE_PREFIX_META    = "Space-Meta-"
	SPACE_PREFIX_PREVIEW = "Space-Preview-"
	SPACE_PREFIX_TAG     = "Space-Tag-"

	THRESHOLD = bcgo.THRESHOLD_G

	MIME_TYPE_UNKNOWN    = "?/?"
	MIME_TYPE_IMAGE_JPEG = "image/jpeg"
	MIME_TYPE_IMAGE_JPG  = "image/jpg"
	MIME_TYPE_IMAGE_GIF  = "image/gif"
	MIME_TYPE_IMAGE_PNG  = "image/png"
	MIME_TYPE_IMAGE_WEBP = "image/webp"
	MIME_TYPE_TEXT_PLAIN = "text/plain"
	MIME_TYPE_PDF        = "application/pdf"
	MIME_TYPE_PROTOBUF   = "application/x-protobuf"
	MIME_TYPE_VIDEO_MPEG = "video/mpeg"
	MIME_TYPE_AUDIO_MPEG = "audio/mpeg"

	MIME_TYPE_IMAGE_DEFAULT = "image/jpeg"
	MIME_TYPE_VIDEO_DEFAULT = "video/mpeg"
	MIME_TYPE_AUDIO_DEFAULT = "audio/mpeg"

	PREVIEW_IMAGE_SIZE  = 128
	PREVIEW_TEXT_LENGTH = 64

	MAX_SIZE_BYTES = bcgo.MAX_PAYLOAD_SIZE_BYTES - 1024 // 10Mb-1Kb (for delta protobuf stuff)
)

type DeltaCallback func(*bcgo.BlockEntry, *Delta) error

type MetaCallback func(*bcgo.BlockEntry, *Meta) error

type PreviewCallback func(*bcgo.BlockEntry, *Preview) error

type RegistrarCallback func(*bcgo.BlockEntry, *Registrar) error

type TagCallback func(*bcgo.BlockEntry, *Tag) error

func GetSpaceHosts() []string {
	if bcgo.IsLive() {
		return []string{
			"space-nyc.aletheiaware.com",
			"space-sfo.aletheiaware.com",
		}
	}
	return []string{
		"test-space.aletheiaware.com",
	}
}

func openLivePoWChannel(name string, threshold uint64) *bcgo.Channel {
	c := bcgo.OpenPoWChannel(name, threshold)
	c.AddValidator(&bcgo.LiveValidator{})
	return c
}

func OpenHourChannel() *bcgo.Channel {
	return openLivePoWChannel(SPACE_HOUR, bcgo.THRESHOLD_PERIOD_HOUR)
}

func OpenDayChannel() *bcgo.Channel {
	return openLivePoWChannel(SPACE_DAY, bcgo.THRESHOLD_PERIOD_DAY)
}

func OpenYearChannel() *bcgo.Channel {
	return openLivePoWChannel(SPACE_YEAR, bcgo.THRESHOLD_PERIOD_YEAR)
}

func OpenChargeChannel() *bcgo.Channel {
	return openLivePoWChannel(SPACE_CHARGE, THRESHOLD)
}

func OpenInvoiceChannel() *bcgo.Channel {
	return openLivePoWChannel(SPACE_INVOICE, THRESHOLD)
}

func OpenRegistrarChannel() *bcgo.Channel {
	return openLivePoWChannel(SPACE_REGISTRAR, THRESHOLD)
}

func OpenRegistrationChannel() *bcgo.Channel {
	return openLivePoWChannel(SPACE_REGISTRATION, THRESHOLD)
}

func OpenSubscriptionChannel() *bcgo.Channel {
	return openLivePoWChannel(SPACE_SUBSCRIPTION, THRESHOLD)
}

func OpenUsageRecordChannel() *bcgo.Channel {
	return openLivePoWChannel(SPACE_USAGE_RECORD, THRESHOLD)
}

func GetDeltaChannelName(metaId string) string {
	return SPACE_PREFIX_DELTA + metaId
}

func GetMetaChannelName(alias string) string {
	return SPACE_PREFIX_META + alias
}

func GetPreviewChannelName(metaId string) string {
	return SPACE_PREFIX_PREVIEW + metaId
}

func GetTagChannelName(metaId string) string {
	return SPACE_PREFIX_TAG + metaId
}

func OpenDeltaChannel(metaId string) *bcgo.Channel {
	return openLivePoWChannel(GetDeltaChannelName(metaId), THRESHOLD)
}

func OpenMetaChannel(alias string) *bcgo.Channel {
	return openLivePoWChannel(GetMetaChannelName(alias), THRESHOLD)
}

func OpenPreviewChannel(metaId string) *bcgo.Channel {
	return openLivePoWChannel(GetPreviewChannelName(metaId), THRESHOLD)
}

func OpenTagChannel(metaId string) *bcgo.Channel {
	return openLivePoWChannel(GetTagChannelName(metaId), THRESHOLD)
}

func ApplyDelta(delta *Delta, input []byte) []byte {
	count := 0
	length := uint64(len(input))
	output := make([]byte, length-delta.Delete+uint64(len(delta.Insert)))
	if delta.Offset <= length {
		count += copy(output, input[:delta.Offset])
	}
	count += copy(output[count:], delta.Insert)
	index := delta.Offset + delta.Delete
	if index < length {
		copy(output[count:], input[index:])
	}
	return output
}

func CreateDeltas(reader io.Reader, max uint64, callback func(*Delta) error) error {
	buffer := make([]byte, max)
	var size uint64
	for {
		count, err := reader.Read(buffer)
		if err != nil {
			if err == io.EOF {
				// Ignore EOFs
				break
			} else {
				return err
			}
		}
		insert := make([]byte, count)
		copy(insert, buffer[:count])
		delta := &Delta{
			Offset: size,
			Insert: insert,
		}
		if err := callback(delta); err != nil {
			return err
		}
		size += uint64(count)
	}
	return nil
}

func GetThreshold(channel string) uint64 {
	switch channel {
	case SPACE_HOUR:
		return bcgo.THRESHOLD_PERIOD_HOUR
	case SPACE_DAY:
		return bcgo.THRESHOLD_PERIOD_DAY
	case SPACE_YEAR:
		return bcgo.THRESHOLD_PERIOD_YEAR
	default:
		return THRESHOLD
	}
}

func GetRegistrar(registrars *bcgo.Channel, cache bcgo.Cache, network bcgo.Network, alias string) (*Registrar, error) {
	var registrar *Registrar
	if err := bcgo.Read(registrars.Name, registrars.Head, nil, cache, network, "", nil, nil, func(entry *bcgo.BlockEntry, key, data []byte) error {
		// Unmarshal as Registrar
		r := &Registrar{}
		err := proto.Unmarshal(data, r)
		if err != nil {
			return err
		}
		if r.Merchant.Alias == alias {
			registrar = r
			return bcgo.StopIterationError{}
		}
		return nil
	}); err != nil {
		switch err.(type) {
		case bcgo.StopIterationError:
			// Do nothing
			break
		default:
			return nil, err
		}
	}
	return registrar, nil
}

func GetRegistrars(registrars *bcgo.Channel, cache bcgo.Cache, network bcgo.Network) (rs []*Registrar) {
	bcgo.Read(registrars.Name, registrars.Head, nil, cache, network, "", nil, nil, func(entry *bcgo.BlockEntry, key, data []byte) error {
		// Unmarshal as Registrar
		r := &Registrar{}
		err := proto.Unmarshal(data, r)
		if err != nil {
			return err
		}
		rs = append(rs, r)
		return nil
	})
	return
}

// GetAllRegistrars triggers the given callback for each registrar.
func GetAllRegistrars(node *bcgo.Node, callback RegistrarCallback) error {
	registrars := node.GetOrOpenChannel(SPACE_REGISTRAR, func() *bcgo.Channel {
		return OpenRegistrarChannel()
	})
	if err := registrars.Refresh(node.Cache, node.Network); err != nil {
		// Ignored
	}
	return bcgo.Read(registrars.Name, registrars.Head, nil, node.Cache, node.Network, "", nil, nil, func(entry *bcgo.BlockEntry, key, data []byte) error {
		// Unmarshal as Registrar
		r := &Registrar{}
		err := proto.Unmarshal(data, r)
		if err != nil {
			return err
		}
		return callback(entry, r)
	})
}

// GetAllRegistrarsForNode triggers the given callback for each registrar with which the given node is registered, and optionally subscribed
func GetAllRegistrarsForNode(node *bcgo.Node, callback func(*Registrar, *financego.Registration, *financego.Subscription) error) error {
	// Get registrars
	as := make(map[string]*Registrar)
	if err := GetAllRegistrars(node, func(e *bcgo.BlockEntry, r *Registrar) error {
		as[r.Merchant.Alias] = r
		return nil
	}); err != nil {
		return err
	}
	// Get registrations
	rs := make(map[string]*financego.Registration)
	if err := GetAllRegistrationsForNode(node, func(e *bcgo.BlockEntry, r *financego.Registration) error {
		if _, ok := as[r.MerchantAlias]; ok {
			rs[r.MerchantAlias] = r
		}
		return nil
	}); err != nil {
		return err
	}
	// Get subscriptions
	ss := make(map[string]*financego.Subscription)
	if err := GetAllSubscriptionsForNode(node, func(e *bcgo.BlockEntry, s *financego.Subscription) error {
		if _, ok := as[s.MerchantAlias]; ok {
			ss[s.MerchantAlias] = s
		}
		return nil
	}); err != nil {
		return err
	}
	for merchant, registrar := range as {
		registration, ok := rs[merchant]
		if !ok {
			continue
		}
		if err := callback(registrar, registration, ss[merchant]); err != nil {
			return err
		}
	}
	return nil
}

// GetAllRegistrationsForNode triggers the given callback for each registration.
func GetAllRegistrationsForNode(node *bcgo.Node, callback financego.RegistrationCallback) error {
	registrations := node.GetOrOpenChannel(SPACE_REGISTRATION, func() *bcgo.Channel {
		return OpenRegistrationChannel()
	})
	if err := registrations.Refresh(node.Cache, node.Network); err != nil {
		// Ignored
	}
	return bcgo.Read(registrations.Name, registrations.Head, nil, node.Cache, node.Network, node.Alias, node.Key, nil, func(entry *bcgo.BlockEntry, key, data []byte) error {
		// Unmarshal as Registration
		r := &financego.Registration{}
		err := proto.Unmarshal(data, r)
		if err != nil {
			return err
		}
		if node.Alias == r.CustomerAlias {
			return callback(entry, r)
		}
		return nil
	})
}

// GetAllSubscriptionsForNode triggers the given callback for each subscription.
func GetAllSubscriptionsForNode(node *bcgo.Node, callback financego.SubscriptionCallback) error {
	subscriptions := node.GetOrOpenChannel(SPACE_SUBSCRIPTION, func() *bcgo.Channel {
		return OpenSubscriptionChannel()
	})
	if err := subscriptions.Refresh(node.Cache, node.Network); err != nil {
		// Ignored
	}
	return bcgo.Read(subscriptions.Name, subscriptions.Head, nil, node.Cache, node.Network, node.Alias, node.Key, nil, func(entry *bcgo.BlockEntry, key, data []byte) error {
		// Unmarshal as Subscription
		s := &financego.Subscription{}
		err := proto.Unmarshal(data, s)
		if err != nil {
			return err
		}
		if node.Alias == s.CustomerAlias {
			return callback(entry, s)
		}
		return nil
	})
}

func GetDelta(deltas *bcgo.Channel, cache bcgo.Cache, network bcgo.Network, alias string, key *rsa.PrivateKey, recordHash []byte, callback DeltaCallback) error {
	return bcgo.Read(deltas.Name, deltas.Head, nil, cache, network, alias, key, recordHash, func(entry *bcgo.BlockEntry, key, data []byte) error {
		// Unmarshal as Delta
		d := &Delta{}
		err := proto.Unmarshal(data, d)
		if err != nil {
			return err
		}
		return callback(entry, d)
	})
}

func GetMeta(metas *bcgo.Channel, cache bcgo.Cache, network bcgo.Network, alias string, key *rsa.PrivateKey, recordHash []byte, callback MetaCallback) error {
	return bcgo.Read(metas.Name, metas.Head, nil, cache, network, alias, key, recordHash, func(entry *bcgo.BlockEntry, key, data []byte) error {
		// Unmarshal as Meta
		meta := &Meta{}
		if err := proto.Unmarshal(data, meta); err != nil {
			return err
		}
		return callback(entry, meta)
	})
}

func GetPreview(previews *bcgo.Channel, cache bcgo.Cache, network bcgo.Network, alias string, key *rsa.PrivateKey, recordHash []byte, callback PreviewCallback) error {
	return bcgo.Read(previews.Name, previews.Head, nil, cache, network, alias, key, recordHash, func(entry *bcgo.BlockEntry, key, data []byte) error {
		// Unmarshal as Preview
		preview := &Preview{}
		if err := proto.Unmarshal(data, preview); err != nil {
			return err
		}
		return callback(entry, preview)
	})
}

func GetTag(tags *bcgo.Channel, cache bcgo.Cache, network bcgo.Network, alias string, key *rsa.PrivateKey, recordHash []byte, callback TagCallback) error {
	return bcgo.Read(tags.Name, tags.Head, nil, cache, network, alias, key, recordHash, func(entry *bcgo.BlockEntry, key, data []byte) error {
		// Unmarshal as Tag
		tag := &Tag{}
		if err := proto.Unmarshal(data, tag); err != nil {
			return err
		}
		return callback(entry, tag)
	})
}

func GetMinimumRegistrars() int {
	if bcgo.IsLive() {
		return 3
	}
	return 1
}

func IterateDeltas(node *bcgo.Node, deltas *bcgo.Channel, callback DeltaCallback) error {
	// Iterate through chain chronologically
	return bcgo.IterateChronologically(deltas.Name, deltas.Head, nil, node.Cache, node.Network, func(hash []byte, block *bcgo.Block) error {
		for _, entry := range block.Entry {
			for _, access := range entry.Record.Access {
				if node.Alias == access.Alias {
					if err := bcgo.DecryptRecord(entry, access, node.Key, func(entry *bcgo.BlockEntry, key []byte, data []byte) error {
						// Unmarshal as Delta
						d := &Delta{}
						if err := proto.Unmarshal(data, d); err != nil {
							return err
						}
						return callback(entry, d)
					}); err != nil {
						return err
					}
				}
			}
		}
		return nil
	})
}
