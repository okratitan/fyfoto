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
	"bytes"
	"fmt"
	"log"
	"reflect"
	"strings"
	"unicode"
)

type Channel interface {
	fmt.Stringer
	Name() string
	Head() []byte
	Timestamp() uint64
	AddTrigger(func())
	AddValidator(Validator)
	Update(Cache, Network, []byte, *Block) error
	Set(uint64, []byte)
	Load(Cache, Network) error
	Refresh(Cache, Network) error
	Pull(Cache, Network) error
	Push(Cache, Network) error
}

func ReadKey(channel string, hash []byte, block *Block, cache Cache, network Network, account Account, recordHash []byte, callback func([]byte) error) error {
	alias := account.Alias()
	return Iterate(channel, hash, block, cache, network, func(h []byte, b *Block) error {
		for _, entry := range b.Entry {
			if recordHash == nil || bytes.Equal(recordHash, entry.RecordHash) {
				if len(entry.Record.Access) == 0 {
					// No Access Declared - Data is public and unencrypted
					if err := callback(nil); err != nil {
						return err
					}
				} else {
					for _, access := range entry.Record.Access {
						if alias == "" || alias == access.Alias {
							decryptedKey, err := account.DecryptKey(access.EncryptionAlgorithm, access.SecretKey)
							if err != nil {
								return err
							}
							return callback(decryptedKey)
						}
					}
				}
			}
		}
		return nil
	})
}

func Read(channel string, hash []byte, block *Block, cache Cache, network Network, account Account, recordHash []byte, callback func(*BlockEntry, []byte, []byte) error) error {
	alias := account.Alias()
	// Decrypt each record in chain and pass to the given callback
	return Iterate(channel, hash, block, cache, network, func(h []byte, b *Block) error {
		for _, entry := range b.Entry {
			if recordHash == nil || bytes.Equal(recordHash, entry.RecordHash) {
				if len(entry.Record.Access) == 0 {
					// No Access Declared - Data is public and unencrypted
					if err := callback(entry, nil, entry.Record.Payload); err != nil {
						return err
					}
				} else {
					for _, access := range entry.Record.Access {
						if alias == "" || alias == access.Alias {
							decryptedKey, err := account.DecryptKey(access.EncryptionAlgorithm, access.SecretKey)
							if err != nil {
								return err
							}
							decryptedPayload, err := account.Decrypt(entry.Record.EncryptionAlgorithm, entry.Record.Payload, decryptedKey)
							if err != nil {
								return err
							}
							return callback(entry, decryptedKey, decryptedPayload)
						}
					}
				}
			}
		}
		return nil
	})
}

func Iterate(channel string, hash []byte, block *Block, cache Cache, network Network, callback func([]byte, *Block) error) error {
	// Iterate throught each block in the chain
	if hash == nil {
		return nil
	}
	var err error
	b := block
	if b == nil {
		b, err = LoadBlock(channel, cache, network, hash)
		if err != nil {
			return err
		}
	}
	for b != nil {
		if err = callback(hash, b); err != nil {
			return err
		}
		hash = b.Previous
		if hash != nil && len(hash) > 0 {
			b, err = LoadBlock(channel, cache, network, hash)
			if err != nil {
				return err
			}
		} else {
			b = nil
		}
	}
	return nil
}

func IterateChronologically(channel string, hash []byte, block *Block, cache Cache, network Network, callback func([]byte, *Block) error) error {
	// Iterate through chain and populate a list of block hashes
	var hashes [][]byte
	if err := Iterate(channel, hash, block, cache, network, func(h []byte, b *Block) error {
		hashes = append(hashes, h)
		return nil
	}); err != nil {
		return err
	}
	// Iterate list of block hashes chronologically
	for i := len(hashes) - 1; i >= 0; i-- {
		hash := hashes[i]
		block, err := LoadBlock(channel, cache, network, hash)
		if err != nil {
			return err
		}
		if err := callback(hash, block); err != nil {
			return err
		}
	}
	return nil
}

func LoadHead(channel string, cache Cache, network Network) (*Reference, error) {
	reference, err := cache.Head(channel)
	if err != nil {
		if network == nil || reflect.ValueOf(network).IsNil() {
			return nil, err
		} else {
			log.Println(err)
		}
	} else {
		return reference, nil
	}
	reference, err = network.Head(channel)
	if err != nil {
		return nil, err
	}
	return reference, nil
}

func LoadBlock(channel string, cache Cache, network Network, hash []byte) (*Block, error) {
	b, err := cache.Block(hash)
	if err != nil {
		if network == nil || reflect.ValueOf(network).IsNil() {
			return nil, err
		} else {
			log.Println(err)
		}
	} else {
		return b, nil
	}

	b, err = network.Block(&Reference{
		ChannelName: channel,
		BlockHash:   hash,
	})
	if err != nil {
		return nil, err
	}

	if err := cache.PutBlock(hash, b); err != nil {
		return nil, err
	}

	return b, nil
}

func LoadBlockContainingRecord(channel string, cache Cache, network Network, hash []byte) (*Block, error) {
	b, err := cache.BlockContainingRecord(channel, hash)
	if err != nil {
		if network == nil || reflect.ValueOf(network).IsNil() {
			return nil, err
		} else {
			log.Println(err)
		}
	} else {
		return b, nil
	}

	b, err = network.Block(&Reference{
		ChannelName: channel,
		RecordHash:  hash,
	})
	if err != nil {
		return nil, err
	}

	bh, err := cryptogo.HashProtobuf(b)
	if err != nil {
		return nil, err
	}

	if err := cache.PutBlock(bh, b); err != nil {
		return nil, err
	}

	return b, nil
}

// ValidateName ensures all characters are in the set [a-zA-Z0-9.-_]
func ValidateName(name string) error {
	if strings.IndexFunc(name, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '.' && r != '-' && r != '_'
	}) != -1 {
		return ErrChannelNameInvalid{name}
	}
	return nil
}

func WriteRecord(channel string, cache Cache, record *Record) (*Reference, error) {
	hash, err := cryptogo.HashProtobuf(record)
	if err != nil {
		return nil, err
	}
	if err := cache.PutBlockEntry(channel, &BlockEntry{
		RecordHash: hash,
		Record:     record,
	}); err != nil {
		return nil, err
	}
	return &Reference{
		Timestamp:   record.Timestamp,
		ChannelName: channel,
		RecordHash:  hash,
	}, nil
}
