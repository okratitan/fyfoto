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

package cryptogo

import "fmt"

// ErrExportFailed is returned when the key cannot be exported.
type ErrExportFailed struct {
	StatusCode int
	Status     string
}

func (e ErrExportFailed) Error() string {
	return fmt.Sprintf("Export error: %d %s", e.StatusCode, e.Status)
}

// ErrPasswordTooShort is returned when the password doesn't have enough characters.
type ErrPasswordTooShort struct {
	Size, Min int
}

func (e ErrPasswordTooShort) Error() string {
	return fmt.Sprintf("Password Too Short: %d Minimum: %d", e.Size, e.Min)
}

// ErrPasswordsDoNotMatch is returned when the password doesn't match the confirmation.
type ErrPasswordsDoNotMatch struct {
}

func (e ErrPasswordsDoNotMatch) Error() string {
	return fmt.Sprintf("Passwords Do Not Match")
}

// ErrUnsupportedEncryption is returned when the algorithm used is not supported.
type ErrUnsupportedEncryption struct {
	Algorithm string
}

func (e ErrUnsupportedEncryption) Error() string {
	return fmt.Sprintf("Unsupported encryption: %s", e.Algorithm)
}

// ErrUnsupportedPublicKeyType is returned when the type used is not supported.
type ErrUnsupportedPublicKeyType struct {
	Type string
}

func (e ErrUnsupportedPublicKeyType) Error() string {
	return fmt.Sprintf("Unsupported Public Key Type: %s", e.Type)
}

// ErrUnsupportedPrivateKeyType is returned when the type used is not supported.
type ErrUnsupportedPrivateKeyType struct {
	Type string
}

func (e ErrUnsupportedPrivateKeyType) Error() string {
	return fmt.Sprintf("Unsupported Private Key Type: %s", e.Type)
}

// ErrUnsupportedPublicKeyFormat is returned when the format used is not supported.
type ErrUnsupportedPublicKeyFormat struct {
	Format string
}

func (e ErrUnsupportedPublicKeyFormat) Error() string {
	return fmt.Sprintf("Unsupported Public Key Format: %s", e.Format)
}

// ErrUnsupportedPrivateKeyFormat is returned when the format used is not supported.
type ErrUnsupportedPrivateKeyFormat struct {
	Format string
}

func (e ErrUnsupportedPrivateKeyFormat) Error() string {
	return fmt.Sprintf("Unsupported Private Key Format: %s", e.Format)
}

// ErrUnsupportedSignature is returned when the algorithm used is not supported.
type ErrUnsupportedSignature struct {
	Algorithm string
}

func (e ErrUnsupportedSignature) Error() string {
	return fmt.Sprintf("Unsupported Signature Algorithm: %s", e.Algorithm)
}
