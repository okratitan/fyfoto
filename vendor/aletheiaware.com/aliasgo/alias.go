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

package aliasgo

import (
	"aletheiaware.com/bcgo"
	"aletheiaware.com/bcgo/channel"
	"aletheiaware.com/bcgo/identity"
	"aletheiaware.com/bcgo/validation"
	"aletheiaware.com/cryptogo"
	"crypto/rsa"
	"encoding/base64"
	"fmt"
	"github.com/golang/protobuf/proto"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"unicode"
)

const (
	ALIAS = "Alias"

	ALIAS_THRESHOLD = bcgo.THRESHOLD_G

	MAX_ALIAS_LENGTH = 100
	MIN_ALIAS_LENGTH = 1
)

func OpenAliasChannel() bcgo.Channel {
	c := channel.New(ALIAS)
	c.AddValidator(&validation.Live{})
	c.AddValidator(validation.NewPoW(ALIAS_THRESHOLD))
	c.AddValidator(&AliasValidator{})
	return c
}

// Validates alias is the correct length and all characters are in the set [a-zA-Z0-9.-_]
func ValidateAlias(alias string) error {
	if strings.IndexFunc(alias, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '.' && r != '-' && r != '_'
	}) != -1 {
		return ErrAliasInvalid{Alias: alias}
	}
	length := len(alias)
	if length < MIN_ALIAS_LENGTH {
		return ErrAliasTooShort{Size: length, Min: MIN_ALIAS_LENGTH}
	}
	if length > MAX_ALIAS_LENGTH {
		return ErrAliasTooLong{Size: length, Max: MAX_ALIAS_LENGTH}
	}
	return nil
}

func UniqueAlias(channel bcgo.Channel, cache bcgo.Cache, network bcgo.Network, alias string) error {
	return bcgo.Iterate(channel.Name(), channel.Head(), nil, cache, network, func(hash []byte, block *bcgo.Block) error {
		for _, entry := range block.Entry {
			record := entry.Record
			if record.Creator == alias {
				a := &Alias{}
				err := proto.Unmarshal(record.Payload, a)
				if err != nil {
					return err
				}
				if a.Alias == alias {
					return ErrAliasAlreadyRegistered{Alias: alias}
				}
			}
		}
		return nil
	})
}

func IterateAliases(channel bcgo.Channel, cache bcgo.Cache, network bcgo.Network, callback func(*bcgo.BlockEntry, *Alias) error) error {
	return bcgo.Iterate(channel.Name(), channel.Head(), nil, cache, network, func(hash []byte, block *bcgo.Block) error {
		for _, entry := range block.Entry {
			alias := &Alias{}
			err := proto.Unmarshal(entry.Record.Payload, alias)
			if err != nil {
				return err
			}
			err = callback(entry, alias)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func AliasForKey(channel bcgo.Channel, cache bcgo.Cache, network bcgo.Network, publicKey *rsa.PublicKey) (*Alias, error) {
	var result *Alias
	if err := bcgo.Iterate(channel.Name(), channel.Head(), nil, cache, network, func(hash []byte, block *bcgo.Block) error {
		for _, entry := range block.Entry {
			record := entry.Record
			a := &Alias{}
			err := proto.Unmarshal(record.Payload, a)
			if err != nil {
				return err
			}
			pk, err := cryptogo.ParseRSAPublicKey(a.PublicKey, a.PublicFormat)
			if err != nil {
				return err
			}
			if publicKey.N.Cmp(pk.N) == 0 && publicKey.E == pk.E {
				result = a
				return bcgo.ErrStopIteration{}
			}
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
	if result == nil {
		return nil, ErrAliasNotFound{}
	}
	return result, nil
}

func PublicKeyForAlias(channel bcgo.Channel, cache bcgo.Cache, network bcgo.Network, alias string) (*rsa.PublicKey, error) {
	var result *rsa.PublicKey
	if err := bcgo.Iterate(channel.Name(), channel.Head(), nil, cache, network, func(hash []byte, block *bcgo.Block) error {
		for _, entry := range block.Entry {
			record := entry.Record
			a := &Alias{}
			err := proto.Unmarshal(record.Payload, a)
			if err != nil {
				return err
			}
			if a.Alias == alias {
				result, err = cryptogo.ParseRSAPublicKey(a.PublicKey, a.PublicFormat)
				if err != nil {
					return err
				}
				return bcgo.ErrStopIteration{}
			}
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
	if result == nil {
		return nil, ErrPublicKeyNotFound{Alias: alias}
	}
	return result, nil
}

func PublicKeysForAliases(channel bcgo.Channel, cache bcgo.Cache, network bcgo.Network, aliases []string) (access []bcgo.Identity) {
	if len(aliases) > 0 {
		alias := &Alias{}
		bcgo.Iterate(channel.Name(), channel.Head(), nil, cache, network, func(hash []byte, block *bcgo.Block) error {
			for _, entry := range block.Entry {
				err := proto.Unmarshal(entry.Record.Payload, alias)
				if err != nil {
					return err
				}
				for _, a := range aliases {
					if alias.Alias == a {
						publicKey, err := cryptogo.ParseRSAPublicKey(alias.PublicKey, alias.PublicFormat)
						if err != nil {
							return err
						}
						access = append(access, identity.NewRSA(a, publicKey))
					}
				}
			}
			return nil
		})
	}
	return
}

func AllPublicKeys(channel bcgo.Channel, cache bcgo.Cache, network bcgo.Network) (map[string]*rsa.PublicKey, error) {
	aliases := make(map[string]*rsa.PublicKey)
	if err := IterateAliases(channel, cache, network, func(e *bcgo.BlockEntry, a *Alias) error {
		key, err := cryptogo.ParseRSAPublicKey(a.PublicKey, a.PublicFormat)
		if err != nil {
			return err
		}
		aliases[a.Alias] = key
		return nil
	}); err != nil {
		return nil, err
	}
	return aliases, nil
}

func Record(channel bcgo.Channel, cache bcgo.Cache, network bcgo.Network, alias string) (*bcgo.Record, *Alias, error) {
	var recordResult *bcgo.Record
	var aliasResult *Alias
	if err := bcgo.Iterate(channel.Name(), channel.Head(), nil, cache, network, func(hash []byte, block *bcgo.Block) error {
		for _, entry := range block.Entry {
			record := entry.Record
			if record.Creator == alias {
				recordResult = record
				aliasResult = &Alias{}
				err := proto.Unmarshal(record.Payload, aliasResult)
				if err != nil {
					return err
				}
				return bcgo.ErrStopIteration{}
			}
		}
		return nil
	}); err != nil {
		switch err.(type) {
		case bcgo.ErrStopIteration:
			// Do nothing
			break
		default:
			return nil, nil, err
		}
	}
	if recordResult == nil || aliasResult == nil {
		return nil, nil, ErrAliasNotFound{}
	}
	return recordResult, aliasResult, nil
}

type AliasValidator struct {
}

func (a *AliasValidator) Validate(channel bcgo.Channel, cache bcgo.Cache, network bcgo.Network, hash []byte, block *bcgo.Block) error {
	register := make(map[string]bool)
	return bcgo.Iterate(channel.Name(), hash, block, cache, network, func(h []byte, b *bcgo.Block) error {
		for _, entry := range b.Entry {
			record := entry.Record
			if len(record.Access) != 0 {
				return ErrAliasNotPublic{}
			}
			a := &Alias{}
			err := proto.Unmarshal(record.Payload, a)
			if err != nil {
				return err
			}
			if err := ValidateAlias(a.Alias); err != nil {
				return err
			}
			v, exists := register[a.Alias]
			if exists || v {
				return ErrAliasAlreadyRegistered{Alias: a.Alias}
			}
			register[a.Alias] = true
		}
		return nil
	})
}

func Register(node bcgo.Node, listener bcgo.MiningListener) error {
	account := node.Account()
	alias := account.Alias()
	cache := node.Cache()
	network := node.Network()
	// Open Alias Channel
	aliases := OpenAliasChannel()
	if err := aliases.Refresh(cache, network); err != nil {
		log.Println(err)
	}
	if err := ValidateAlias(alias); err != nil {
		return err
	}
	// Check Alias is unique
	if err := UniqueAlias(aliases, cache, network, alias); err != nil {
		return err
	}
	// Register Alias
	if err := RegisterAlias(bcgo.BCWebsite(), account); err != nil {
		log.Println("Could not register alias remotely:", err)
		log.Println("Registering locally")
		// Create record
		record, err := CreateSignedAliasRecord(account)
		if err != nil {
			return err
		}

		// Write record to cache
		if _, err := bcgo.WriteRecord(ALIAS, cache, record); err != nil {
			return err
		}

		// Mine record into blockchain
		if _, _, err := bcgo.Mine(node, aliases, ALIAS_THRESHOLD, listener); err != nil {
			return err
		}

		// Push update to peers
		if err := aliases.Push(cache, network); err != nil {
			return err
		}
	}
	return nil
}

func CreateSignedAliasRecord(account bcgo.Account) (*bcgo.Record, error) {
	alias := account.Alias()
	if err := ValidateAlias(alias); err != nil {
		return nil, err
	}

	publicKey, publicKeyFormat, err := account.PublicKey()
	if err != nil {
		return nil, err
	}

	a := &Alias{
		Alias:        alias,
		PublicKey:    publicKey,
		PublicFormat: publicKeyFormat,
	}
	data, err := proto.Marshal(a)
	if err != nil {
		return nil, err
	}

	signature, algorithm, err := account.Sign(data)
	if err != nil {
		return nil, err
	}

	return CreateAliasRecord(alias, publicKey, publicKeyFormat, signature, algorithm)
}

func CreateAliasRecord(alias string, publicKey []byte, publicKeyFormat cryptogo.PublicKeyFormat, signature []byte, signatureAlgorithm cryptogo.SignatureAlgorithm) (*bcgo.Record, error) {
	if err := ValidateAlias(alias); err != nil {
		return nil, err
	}

	pubKey, err := cryptogo.ParseRSAPublicKey(publicKey, publicKeyFormat)
	if err != nil {
		return nil, err
	}

	a := &Alias{
		Alias:        alias,
		PublicKey:    publicKey,
		PublicFormat: publicKeyFormat,
	}
	data, err := proto.Marshal(a)
	if err != nil {
		return nil, err
	}

	if err := cryptogo.VerifySignature(pubKey, cryptogo.Hash(data), signature, signatureAlgorithm); err != nil {
		return nil, err
	}

	record := &bcgo.Record{
		Timestamp:           bcgo.Timestamp(),
		Creator:             alias,
		Payload:             data,
		EncryptionAlgorithm: cryptogo.EncryptionAlgorithm_UNKNOWN_ENCRYPTION,
		Signature:           signature,
		SignatureAlgorithm:  signatureAlgorithm,
	}
	if l, ok := os.LookupEnv(bcgo.LIVE_FLAG); ok {
		record.Meta = map[string]string{
			bcgo.LIVE_FLAG: l,
		}
	}
	return record, nil
}

func RegisterAlias(host string, account bcgo.Account) error {
	alias := account.Alias()
	if err := ValidateAlias(alias); err != nil {
		return err
	}

	publicKey, publicKeyFormat, err := account.PublicKey()
	if err != nil {
		return err
	}

	data, err := proto.Marshal(&Alias{
		Alias:        alias,
		PublicKey:    publicKey,
		PublicFormat: publicKeyFormat,
	})
	if err != nil {
		return err
	}
	signature, algorithm, err := account.Sign(data)
	if err != nil {
		return err
	}

	response, err := http.PostForm(host+"/alias-register", url.Values{
		"alias":              {alias},
		"publicKey":          {base64.RawURLEncoding.EncodeToString(publicKey)},
		"publicKeyFormat":    {publicKeyFormat.String()},
		"signature":          {base64.RawURLEncoding.EncodeToString(signature)},
		"signatureAlgorithm": {algorithm.String()},
	})
	if err != nil {
		return err
	}
	switch response.StatusCode {
	case http.StatusOK:
		return nil
	default:
		return fmt.Errorf("Registration status: %s", response.Status)
	}
}
