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

package bcgo

import "aletheiaware.com/cryptogo"

type Identity interface {
	Alias() string
	PublicKey() (cryptogo.PublicKeyFormat, []byte, error)
	// Encrypt takes a Plaintext Payload
	// A new AES 256bit Symmetric Key is generated, used to encrypt the Payload.
	// Encrypt returns the Encryption Algorithm, Encrypted Payload, Key used, or an error.
	Encrypt([]byte) (cryptogo.EncryptionAlgorithm, []byte, []byte, error)
	// Encrypt takes a Plaintext key
	// The given key is encrypted with this Identity's Public Key
	// EncryptKey returns the Encryption Algorithm and Encrypted Key, or an error.
	EncryptKey([]byte) (cryptogo.EncryptionAlgorithm, []byte, error)
	// Verify takes a Signature Algorithm, Payload, and Signature.
	// Verify returns an error if the signature cannot be verified.
	Verify(cryptogo.SignatureAlgorithm, []byte, []byte) error
}
