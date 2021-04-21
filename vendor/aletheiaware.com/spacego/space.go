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
	"aletheiaware.com/bcgo/channel"
	"aletheiaware.com/bcgo/validation"
	"aletheiaware.com/financego"
	"github.com/golang/protobuf/proto"
	"io"
	"log"
	"sort"
	"strings"
)

const (
	SPACE              = "S P A C E"
	SPACE_CHARGE       = "Space-Charge"
	SPACE_INVOICE      = "Space-Invoice"
	SPACE_REGISTRAR    = "Space-Registrar"
	SPACE_REGISTRATION = "Space-Registration"
	SPACE_SUBSCRIPTION = "Space-Subscription"
	SPACE_USAGE_RECORD = "Space-Usage-Record"

	SPACE_PREFIX            = "Space-"
	SPACE_PREFIX_DELTA      = "Space-Delta-"
	SPACE_PREFIX_META       = "Space-Meta-"
	SPACE_PREFIX_PREVIEW    = "Space-Preview-"
	SPACE_PREFIX_TAG        = "Space-Tag-"
	SPACE_PREFIX_VALIDATION = "Space-Validation-"

	THRESHOLD_ACCOUNTING = bcgo.THRESHOLD_G // Charge, Invoice, Registrar, Registration, Subscription, Usage Record Channels
	THRESHOLD_CUSTOMER   = bcgo.THRESHOLD_Z // Delta, Meta, Preview, Tag Channels
	THRESHOLD_VALIDATION = validation.THRESHOLD_PERIOD_DAY

	PERIOD_VALIDATION = validation.PERIOD_DAILY

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

func SpaceHosts() []string {
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

func MimeTypes() []string {
	mimes := []string{
		MIME_TYPE_IMAGE_JPEG,
		MIME_TYPE_IMAGE_JPG,
		MIME_TYPE_IMAGE_GIF,
		MIME_TYPE_IMAGE_PNG,
		MIME_TYPE_IMAGE_WEBP,
		MIME_TYPE_TEXT_PLAIN,
		MIME_TYPE_PDF,
		MIME_TYPE_PROTOBUF,
		MIME_TYPE_VIDEO_MPEG,
		MIME_TYPE_AUDIO_MPEG,
	}
	sort.Strings(mimes)
	return mimes
}

func Validator(node bcgo.Node, channel bcgo.Channel, listener bcgo.MiningListener) validation.Periodic {
	return validation.NewDaily(node, channel, listener)
}

func openChannel(name string, threshold uint64) bcgo.Channel {
	c := channel.NewPoW(name, threshold)
	c.AddValidator(&validation.Live{})
	// TODO c.AddValidator(&validation.Signature{})
	return c
}

/* TODO
func openCustomerChannel(customer, name string, threshold uint64) bcgo.Channel {
	c := openChannel(name, threshold)
	c.AddValidator(&validation.Creator{
		Creator: customer,
	})
	return c
}
*/

func OpenChargeChannel() bcgo.Channel {
	return openChannel(SPACE_CHARGE, THRESHOLD_ACCOUNTING)
}

func OpenInvoiceChannel() bcgo.Channel {
	return openChannel(SPACE_INVOICE, THRESHOLD_ACCOUNTING)
}

func OpenRegistrarChannel() bcgo.Channel {
	return openChannel(SPACE_REGISTRAR, THRESHOLD_ACCOUNTING)
}

func OpenRegistrationChannel() bcgo.Channel {
	return openChannel(SPACE_REGISTRATION, THRESHOLD_ACCOUNTING)
}

func OpenSubscriptionChannel() bcgo.Channel {
	return openChannel(SPACE_SUBSCRIPTION, THRESHOLD_ACCOUNTING)
}

func OpenUsageRecordChannel() bcgo.Channel {
	return openChannel(SPACE_USAGE_RECORD, THRESHOLD_ACCOUNTING)
}

func DeltaChannelName(metaId string) string {
	return SPACE_PREFIX_DELTA + metaId
}

func MetaChannelName(alias string) string {
	return SPACE_PREFIX_META + alias
}

func PreviewChannelName(metaId string) string {
	return SPACE_PREFIX_PREVIEW + metaId
}

func TagChannelName(metaId string) string {
	return SPACE_PREFIX_TAG + metaId
}

func ValidationChannelName(alias string) string {
	return SPACE_PREFIX_VALIDATION + alias
}

func OpenDeltaChannel(metaId string) bcgo.Channel {
	// TODO return openCustomerChannel(alias, DeltaChannelName(metaId), THRESHOLD_CUSTOMER)
	return openChannel(DeltaChannelName(metaId), THRESHOLD_CUSTOMER)
}

func OpenMetaChannel(alias string) bcgo.Channel {
	// TODO return openCustomerChannel(alias, MetaChannelName(alias), THRESHOLD_CUSTOMER)
	return openChannel(MetaChannelName(alias), THRESHOLD_CUSTOMER)
}

func OpenPreviewChannel(metaId string) bcgo.Channel {
	// TODO return openCustomerChannel(alias, PreviewChannelName(metaId), THRESHOLD_CUSTOMER)
	return openChannel(PreviewChannelName(metaId), THRESHOLD_CUSTOMER)
}

func OpenTagChannel(metaId string) bcgo.Channel {
	// TODO return openCustomerChannel(alias, TagChannelName(metaId), THRESHOLD_CUSTOMER)
	return openChannel(TagChannelName(metaId), THRESHOLD_CUSTOMER)
}

func OpenValidationChannel(alias string) bcgo.Channel {
	return openChannel(ValidationChannelName(alias), THRESHOLD_VALIDATION)
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

func Threshold(channel string) uint64 {
	switch channel {
	case SPACE_CHARGE,
		SPACE_INVOICE,
		SPACE_REGISTRAR,
		SPACE_REGISTRATION,
		SPACE_SUBSCRIPTION,
		SPACE_USAGE_RECORD:
		return THRESHOLD_ACCOUNTING
	default:
		switch {
		case strings.HasPrefix(channel, SPACE_PREFIX_VALIDATION):
			return THRESHOLD_VALIDATION
		case strings.HasPrefix(channel, SPACE_PREFIX_DELTA),
			strings.HasPrefix(channel, SPACE_PREFIX_META),
			strings.HasPrefix(channel, SPACE_PREFIX_PREVIEW),
			strings.HasPrefix(channel, SPACE_PREFIX_TAG):
			return THRESHOLD_CUSTOMER
		default:
			return 0
		}
	}
}

func RegistrarForAlias(registrars bcgo.Channel, cache bcgo.Cache, network bcgo.Network, alias string) (*Registrar, error) {
	var registrar *Registrar
	if err := bcgo.Read(registrars.Name(), registrars.Head(), nil, cache, network, nil, nil, func(entry *bcgo.BlockEntry, key, data []byte) error {
		// Unmarshal as Registrar
		r := &Registrar{}
		err := proto.Unmarshal(data, r)
		if err != nil {
			return err
		}
		if r.Merchant.Alias == alias {
			registrar = r
			return bcgo.ErrStopIteration{}
		}
		return nil
	}); err != nil {
		switch err.(type) {
		case bcgo.ErrStopIteration:
			// Do nothing
			break
		default:
			return nil, err
		}
	}
	return registrar, nil
}

// AllRegistrars triggers the given callback for each registrar.
func AllRegistrars(node bcgo.Node, callback RegistrarCallback) error {
	registrars := node.OpenChannel(SPACE_REGISTRAR, func() bcgo.Channel {
		return OpenRegistrarChannel()
	})
	if err := registrars.Refresh(node.Cache(), node.Network()); err != nil {
		log.Println(err)
	}
	return bcgo.Read(registrars.Name(), registrars.Head(), nil, node.Cache(), node.Network(), node.Account(), nil, func(entry *bcgo.BlockEntry, key, data []byte) error {
		// Unmarshal as Registrar
		r := &Registrar{}
		err := proto.Unmarshal(data, r)
		if err != nil {
			return err
		}
		return callback(entry, r)
	})
}

// AllRegistrarsForNode triggers the given callback for each registrar with which the given node is registered, and optionally subscribed
func AllRegistrarsForNode(node bcgo.Node, callback func(*Registrar, *financego.Registration, *financego.Subscription) error) error {
	// Get registrars
	as := make(map[string]*Registrar)
	if err := AllRegistrars(node, func(e *bcgo.BlockEntry, r *Registrar) error {
		as[r.Merchant.Alias] = r
		return nil
	}); err != nil {
		return err
	}
	// Get registrations
	rs := make(map[string]*financego.Registration)
	if err := AllRegistrationsForNode(node, func(e *bcgo.BlockEntry, r *financego.Registration) error {
		if _, ok := as[r.MerchantAlias]; ok {
			rs[r.MerchantAlias] = r
		}
		return nil
	}); err != nil {
		return err
	}
	// Get subscriptions
	ss := make(map[string]*financego.Subscription)
	if err := AllSubscriptionsForNode(node, func(e *bcgo.BlockEntry, s *financego.Subscription) error {
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

// AllRegistrationsForNode triggers the given callback for each registration.
func AllRegistrationsForNode(node bcgo.Node, callback financego.RegistrationCallback) error {
	registrations := node.OpenChannel(SPACE_REGISTRATION, func() bcgo.Channel {
		return OpenRegistrationChannel()
	})
	if err := registrations.Refresh(node.Cache(), node.Network()); err != nil {
		log.Println(err)
	}
	alias := node.Account().Alias()
	return bcgo.Read(registrations.Name(), registrations.Head(), nil, node.Cache(), node.Network(), node.Account(), nil, func(entry *bcgo.BlockEntry, key, data []byte) error {
		// Unmarshal as Registration
		r := &financego.Registration{}
		err := proto.Unmarshal(data, r)
		if err != nil {
			return err
		}
		if alias == r.CustomerAlias {
			return callback(entry, r)
		}
		return nil
	})
}

// AllSubscriptionsForNode triggers the given callback for each subscription.
func AllSubscriptionsForNode(node bcgo.Node, callback financego.SubscriptionCallback) error {
	subscriptions := node.OpenChannel(SPACE_SUBSCRIPTION, func() bcgo.Channel {
		return OpenSubscriptionChannel()
	})
	if err := subscriptions.Refresh(node.Cache(), node.Network()); err != nil {
		log.Println(err)
	}
	alias := node.Account().Alias()
	return bcgo.Read(subscriptions.Name(), subscriptions.Head(), nil, node.Cache(), node.Network(), node.Account(), nil, func(entry *bcgo.BlockEntry, key, data []byte) error {
		// Unmarshal as Subscription
		s := &financego.Subscription{}
		err := proto.Unmarshal(data, s)
		if err != nil {
			return err
		}
		if alias == s.CustomerAlias {
			return callback(entry, s)
		}
		return nil
	})
}

func ReadDelta(deltas bcgo.Channel, cache bcgo.Cache, network bcgo.Network, account bcgo.Account, recordHash []byte, callback DeltaCallback) error {
	return bcgo.Read(deltas.Name(), deltas.Head(), nil, cache, network, account, recordHash, func(entry *bcgo.BlockEntry, key, data []byte) error {
		// Unmarshal as Delta
		d := &Delta{}
		err := proto.Unmarshal(data, d)
		if err != nil {
			return err
		}
		return callback(entry, d)
	})
}

func ReadMeta(metas bcgo.Channel, cache bcgo.Cache, network bcgo.Network, account bcgo.Account, recordHash []byte, callback MetaCallback) error {
	return bcgo.Read(metas.Name(), metas.Head(), nil, cache, network, account, recordHash, func(entry *bcgo.BlockEntry, key, data []byte) error {
		// Unmarshal as Meta
		meta := &Meta{}
		if err := proto.Unmarshal(data, meta); err != nil {
			return err
		}
		return callback(entry, meta)
	})
}

func ReadPreview(previews bcgo.Channel, cache bcgo.Cache, network bcgo.Network, account bcgo.Account, recordHash []byte, callback PreviewCallback) error {
	return bcgo.Read(previews.Name(), previews.Head(), nil, cache, network, account, recordHash, func(entry *bcgo.BlockEntry, key, data []byte) error {
		// Unmarshal as Preview
		preview := &Preview{}
		if err := proto.Unmarshal(data, preview); err != nil {
			return err
		}
		return callback(entry, preview)
	})
}

func ReadTag(tags bcgo.Channel, cache bcgo.Cache, network bcgo.Network, account bcgo.Account, recordHash []byte, callback TagCallback) error {
	return bcgo.Read(tags.Name(), tags.Head(), nil, cache, network, account, recordHash, func(entry *bcgo.BlockEntry, key, data []byte) error {
		// Unmarshal as Tag
		tag := &Tag{}
		if err := proto.Unmarshal(data, tag); err != nil {
			return err
		}
		return callback(entry, tag)
	})
}

func MinimumRegistrars() int {
	if bcgo.IsLive() {
		return 3
	}
	return 1
}

func IterateDeltas(node bcgo.Node, deltas bcgo.Channel, callback DeltaCallback) error {
	account := node.Account()
	alias := account.Alias()
	// Iterate through chain chronologically
	return bcgo.IterateChronologically(deltas.Name(), deltas.Head(), nil, node.Cache(), node.Network(), func(hash []byte, block *bcgo.Block) error {
		for _, entry := range block.Entry {
			for _, access := range entry.Record.Access {
				if alias == access.Alias {
					if err := account.Decrypt(entry, access, func(entry *bcgo.BlockEntry, key []byte, data []byte) error {
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
