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

func (a *rsaIdentity) PublicKey() (cryptogo.PublicKeyFormat, []byte, error) {
	bytes, err := cryptogo.RSAPublicKeyToPKIXBytes(a.key)
	if err != nil {
		return cryptogo.PublicKeyFormat_UNKNOWN_PUBLIC_KEY_FORMAT, nil, err
	}
	return cryptogo.PublicKeyFormat_PKIX, bytes, nil
}

func (a *rsaIdentity) Encrypt(payload []byte) (cryptogo.EncryptionAlgorithm, []byte, []byte, error) {
	key, err := cryptogo.GenerateRandomKey(cryptogo.AES_256_KEY_SIZE_BYTES)
	if err != nil {
		return cryptogo.EncryptionAlgorithm_UNKNOWN_ENCRYPTION, nil, nil, err
	}

	encryptedPayload, err := cryptogo.EncryptAESGCM(key, payload)
	if err != nil {
		return cryptogo.EncryptionAlgorithm_UNKNOWN_ENCRYPTION, nil, nil, err
	}

	return cryptogo.EncryptionAlgorithm_AES_256_GCM_NOPADDING, encryptedPayload, key, nil
}

func (a *rsaIdentity) EncryptKey(key []byte) (cryptogo.EncryptionAlgorithm, []byte, error) {
	encrypted, err := rsa.EncryptOAEP(sha512.New(), rand.Reader, a.key, key, nil)
	if err != nil {
		return cryptogo.EncryptionAlgorithm_UNKNOWN_ENCRYPTION, nil, err
	}
	return cryptogo.EncryptionAlgorithm_RSA_ECB_OAEPPADDING, encrypted, nil
}

func (a *rsaIdentity) Verify(algorithm cryptogo.SignatureAlgorithm, data, signature []byte) error {
	return cryptogo.VerifySignature(algorithm, a.key, data, signature)
}
