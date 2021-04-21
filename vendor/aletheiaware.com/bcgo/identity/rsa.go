/*
 * Copyright 2021 Aletheia Ware LLC
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

package identity

import (
	"aletheiaware.com/bcgo"
	"aletheiaware.com/cryptogo"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha512"
)

type rsaIdentity struct {
	alias string
	key   *rsa.PublicKey
}

func NewRSA(alias string, key *rsa.PublicKey) bcgo.Identity {
	return &rsaIdentity{
		alias: alias,
		key:   key,
	}
}

func (a *rsaIdentity) Alias() string {
	return a.alias
}

func (a *rsaIdentity) PublicKey() ([]byte, cryptogo.PublicKeyFormat, error) {
	bytes, err := cryptogo.RSAPublicKeyToPKIXBytes(a.key)
	if err != nil {
		return nil, cryptogo.PublicKeyFormat_UNKNOWN_PUBLIC_KEY_FORMAT, err
	}
	return bytes, cryptogo.PublicKeyFormat_PKIX, nil
}

func (a *rsaIdentity) Encrypt(data []byte) ([]byte, cryptogo.EncryptionAlgorithm, error) {
	encrypted, err := rsa.EncryptOAEP(sha512.New(), rand.Reader, a.key, data, nil)
	if err != nil {
		return nil, cryptogo.EncryptionAlgorithm_UNKNOWN_ENCRYPTION, err
	}
	return encrypted, cryptogo.EncryptionAlgorithm_RSA_ECB_OAEPPADDING, nil
}

func (a *rsaIdentity) EncryptKey(key []byte) (*bcgo.Record_Access, error) {
	encrypted, algorithm, err := a.Encrypt(key)
	if err != nil {
		return nil, err
	}
	return &bcgo.Record_Access{
		Alias:               a.alias,
		SecretKey:           encrypted,
		EncryptionAlgorithm: algorithm,
	}, nil
}

func (a *rsaIdentity) Verify(data, signature []byte, algorithm cryptogo.SignatureAlgorithm) error {
	return cryptogo.VerifySignature(a.key, data, signature, algorithm)
}
