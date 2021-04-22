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

package account

import (
	"aletheiaware.com/bcgo"
	"aletheiaware.com/bcgo/identity"
	"aletheiaware.com/cryptogo"
	"crypto/rand"
	"crypto/rsa"
)

type rsaAccount struct {
	bcgo.Identity
	key *rsa.PrivateKey
}

func GenerateRSA(alias string) (bcgo.Account, error) {
	key, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, err
	}
	return NewRSA(alias, key), nil
}

func LoadRSA(directory string) (bcgo.Account, error) {
	// Get alias
	alias, err := bcgo.Alias()
	if err != nil {
		return nil, err
	}
	keystore, err := bcgo.KeyDirectory(directory)
	if err != nil {
		return nil, err
	}
	// Get private key
	key, err := cryptogo.LoadRSAPrivateKey(keystore, alias)
	if err != nil {
		return nil, err
	}
	return NewRSA(alias, key), nil
}

func NewRSA(alias string, key *rsa.PrivateKey) bcgo.Account {
	return &rsaAccount{
		Identity: identity.NewRSA(alias, &key.PublicKey),
		key:      key,
	}
}

func (a *rsaAccount) DecryptKey(algorithm cryptogo.EncryptionAlgorithm, key []byte) ([]byte, error) {
	return cryptogo.DecryptKey(algorithm, key, a.key)
}

func (a *rsaAccount) Sign(data []byte) ([]byte, cryptogo.SignatureAlgorithm, error) {
	// Hash payload
	hashed := cryptogo.Hash(data)

	algorithm := cryptogo.SignatureAlgorithm_SHA512WITHRSA_PSS

	// Sign hash of encrypted payload
	signature, err := cryptogo.CreateSignature(a.key, hashed, algorithm)
	if err != nil {
		return nil, cryptogo.SignatureAlgorithm_UNKNOWN_SIGNATURE, err
	}

	return signature, algorithm, nil
}
