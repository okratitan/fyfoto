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

type Account interface {
	Identity
	// Decrypt takes an Encryption Algorithm, Encrypted Payload, and Symmetric Key.
	// The Key is used to decrypt the Payload.
	// Decrypt returns the Decrypted Payload, or an error.
	Decrypt(cryptogo.EncryptionAlgorithm, []byte, []byte) ([]byte, error)
	// DecryptKey takes an Encrytion Algorithm and Encrypted Key.
	// The Key is decrypted with this Account's Private Key.
	// DecryptKey returns the Decrypted Key, or an error
	DecryptKey(cryptogo.EncryptionAlgorithm, []byte) ([]byte, error)
	// Sign takes a Payload.
	// Sign returns the Signature, Algorithm, or an error.
	Sign([]byte) (cryptogo.SignatureAlgorithm, []byte, error)
}
