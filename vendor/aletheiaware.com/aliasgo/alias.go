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
	"aletheiaware.com/cryptogo"
	"crypto/rsa"
	"encoding/base64"
	"errors"
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

	ERROR_ALIAS_ALREADY_REGISTERED = "Alias Already Registered: %s"
	ERROR_ALIAS_INVALID            = "Alias Invalid: %s"
	ERROR_ALIAS_NOT_FOUND          = "Could Not Find Alias For Public Key"
	ERROR_ALIAS_NOT_PUBLIC         = "Cannot Register Private Alias"
	ERROR_ALIAS_TOO_LONG           = "Alias Too Long: %d Maximum: %d"
	ERROR_ALIAS_TOO_SHORT          = "Alias Too Short: %d Minimum: %d"
	ERROR_PUBLIC_KEY_NOT_FOUND     = "Could Not Find Public Key For Alias"
)

func OpenAliasChannel() *bcgo.Channel {
	return &bcgo.Channel{
		Name: ALIAS,
		Validators: []bcgo.Validator{
			&bcgo.LiveValidator{},
			&bcgo.PoWValidator{
				Threshold: ALIAS_THRESHOLD,
			},
			&AliasValidator{},
		},
	}
}

// Validates alias is the correct length and all characters are in the set [a-zA-Z0-9.-_]
func ValidateAlias(alias string) error {
	if strings.IndexFunc(alias, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '.' && r != '-' && r != '_'
	}) != -1 {
		return fmt.Errorf(ERROR_ALIAS_INVALID, alias)
	}
	length := len(alias)
	if length < MIN_ALIAS_LENGTH {
		return fmt.Errorf(ERROR_ALIAS_TOO_SHORT, length, MIN_ALIAS_LENGTH)
	}
	if length > MAX_ALIAS_LENGTH {
		return fmt.Errorf(ERROR_ALIAS_TOO_LONG, length, MAX_ALIAS_LENGTH)
	}
	return nil
}

func UniqueAlias(channel *bcgo.Channel, cache bcgo.Cache, network bcgo.Network, alias string) error {
	return bcgo.Iterate(channel.Name, channel.Head, nil, cache, network, func(hash []byte, block *bcgo.Block) error {
		for _, entry := range block.Entry {
			record := entry.Record
			if record.Creator == alias {
				a := &Alias{}
				err := proto.Unmarshal(record.Payload, a)
				if err != nil {
					return err
				}
				if a.Alias == alias {
					return fmt.Errorf(ERROR_ALIAS_ALREADY_REGISTERED, alias)
				}
			}
		}
		return nil
	})
}

func IterateAliases(channel *bcgo.Channel, cache bcgo.Cache, network bcgo.Network, callback func(*bcgo.BlockEntry, *Alias) error) error {
	return bcgo.Iterate(channel.Name, channel.Head, nil, cache, network, func(hash []byte, block *bcgo.Block) error {
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

func GetAlias(channel *bcgo.Channel, cache bcgo.Cache, network bcgo.Network, publicKey *rsa.PublicKey) (*Alias, error) {
	var result *Alias
	if err := bcgo.Iterate(channel.Name, channel.Head, nil, cache, network, func(hash []byte, block *bcgo.Block) error {
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
				return bcgo.StopIterationError{}
			}
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
	if result == nil {
		return nil, errors.New(ERROR_ALIAS_NOT_FOUND)
	}
	return result, nil
}

func GetPublicKey(channel *bcgo.Channel, cache bcgo.Cache, network bcgo.Network, alias string) (*rsa.PublicKey, error) {
	var result *rsa.PublicKey
	if err := bcgo.Iterate(channel.Name, channel.Head, nil, cache, network, func(hash []byte, block *bcgo.Block) error {
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
				return bcgo.StopIterationError{}
			}
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
	if result == nil {
		return nil, errors.New(ERROR_PUBLIC_KEY_NOT_FOUND)
	}
	return result, nil
}

func GetPublicKeys(channel *bcgo.Channel, cache bcgo.Cache, network bcgo.Network, addresses []string) map[string]*rsa.PublicKey {
	acl := make(map[string]*rsa.PublicKey)
	if len(addresses) > 0 {
		alias := &Alias{}
		bcgo.Iterate(channel.Name, channel.Head, nil, cache, network, func(hash []byte, block *bcgo.Block) error {
			for _, entry := range block.Entry {
				err := proto.Unmarshal(entry.Record.Payload, alias)
				if err != nil {
					return err
				}
				for _, address := range addresses {
					if alias.Alias == address {
						publicKey, err := cryptogo.ParseRSAPublicKey(alias.PublicKey, alias.PublicFormat)
						if err != nil {
							return err
						}
						acl[address] = publicKey
					}
				}
			}
			return nil
		})
	}
	return acl
}

func GetRecord(channel *bcgo.Channel, cache bcgo.Cache, network bcgo.Network, alias string) (*bcgo.Record, *Alias, error) {
	var recordResult *bcgo.Record
	var aliasResult *Alias
	if err := bcgo.Iterate(channel.Name, channel.Head, nil, cache, network, func(hash []byte, block *bcgo.Block) error {
		for _, entry := range block.Entry {
			record := entry.Record
			if record.Creator == alias {
				recordResult = record
				aliasResult = &Alias{}
				err := proto.Unmarshal(record.Payload, aliasResult)
				if err != nil {
					return err
				}
				return bcgo.StopIterationError{}
			}
		}
		return nil
	}); err != nil {
		switch err.(type) {
		case bcgo.StopIterationError:
			// Do nothing
			break
		default:
			return nil, nil, err
		}
	}
	if recordResult == nil || aliasResult == nil {
		return nil, nil, errors.New(ERROR_ALIAS_NOT_FOUND)
	}
	return recordResult, aliasResult, nil
}

type AliasValidator struct {
}

func (a *AliasValidator) Validate(channel *bcgo.Channel, cache bcgo.Cache, network bcgo.Network, hash []byte, block *bcgo.Block) error {
	register := make(map[string]bool)
	return bcgo.Iterate(channel.Name, hash, block, cache, network, func(h []byte, b *bcgo.Block) error {
		for _, entry := range b.Entry {
			record := entry.Record
			if len(record.Access) != 0 {
				return fmt.Errorf(ERROR_ALIAS_NOT_PUBLIC)
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
				return fmt.Errorf(ERROR_ALIAS_ALREADY_REGISTERED, a.Alias)
			}
			register[a.Alias] = true
		}
		return nil
	})
}

func Register(node *bcgo.Node, listener bcgo.MiningListener) error {
	// Open Alias Channel
	aliases := OpenAliasChannel()
	if err := aliases.Refresh(node.Cache, node.Network); err != nil {
		log.Println(err)
	}
	if err := ValidateAlias(node.Alias); err != nil {
		return err
	}
	// Check Alias is unique
	if err := UniqueAlias(aliases, node.Cache, node.Network, node.Alias); err != nil {
		return err
	}
	// Register Alias
	if err := RegisterAlias(bcgo.GetBCWebsite(), node.Alias, node.Key); err != nil {
		log.Println("Could not register alias remotely:", err)
		log.Println("Registering locally")
		// Create record
		record, err := CreateSignedAliasRecord(node.Alias, node.Key)
		if err != nil {
			return err
		}

		// Write record to cache
		if _, err := bcgo.WriteRecord(ALIAS, node.Cache, record); err != nil {
			return err
		}

		// Mine record into blockchain
		if _, _, err := node.Mine(aliases, ALIAS_THRESHOLD, listener); err != nil {
			return err
		}

		// Push update to peers
		if err := aliases.Push(node.Cache, node.Network); err != nil {
			return err
		}
	}
	return nil
}

func CreateSignedAliasRecord(alias string, privateKey *rsa.PrivateKey) (*bcgo.Record, error) {
	if err := ValidateAlias(alias); err != nil {
		return nil, err
	}

	publicKeyBytes, err := cryptogo.RSAPublicKeyToPKIXBytes(&privateKey.PublicKey)
	if err != nil {
		return nil, err
	}

	publicKeyFormat := cryptogo.PublicKeyFormat_PKIX
	hash, err := cryptogo.HashProtobuf(&Alias{
		Alias:        alias,
		PublicKey:    publicKeyBytes,
		PublicFormat: publicKeyFormat,
	})
	if err != nil {
		return nil, err
	}

	signatureAlgorithm := cryptogo.SignatureAlgorithm_SHA512WITHRSA_PSS
	signature, err := cryptogo.CreateSignature(privateKey, hash, signatureAlgorithm)
	if err != nil {
		return nil, err
	}

	return CreateAliasRecord(alias, publicKeyBytes, publicKeyFormat, signature, signatureAlgorithm)
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

func RegisterAlias(host, alias string, key *rsa.PrivateKey) error {
	if err := ValidateAlias(alias); err != nil {
		return err
	}

	publicKeyBytes, err := cryptogo.RSAPublicKeyToPKIXBytes(&key.PublicKey)
	if err != nil {
		return err
	}

	data, err := proto.Marshal(&Alias{
		Alias:        alias,
		PublicKey:    publicKeyBytes,
		PublicFormat: cryptogo.PublicKeyFormat_PKIX,
	})
	if err != nil {
		return err
	}

	signatureAlgorithm := cryptogo.SignatureAlgorithm_SHA512WITHRSA_PSS

	signature, err := cryptogo.CreateSignature(key, cryptogo.Hash(data), signatureAlgorithm)
	if err != nil {
		return err
	}

	response, err := http.PostForm(host+"/alias-register", url.Values{
		"alias":              {alias},
		"publicKey":          {base64.RawURLEncoding.EncodeToString(publicKeyBytes)},
		"publicKeyFormat":    {"PKIX"},
		"signature":          {base64.RawURLEncoding.EncodeToString(signature)},
		"signatureAlgorithm": {signatureAlgorithm.String()},
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
