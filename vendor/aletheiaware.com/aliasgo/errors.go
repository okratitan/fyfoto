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

package aliasgo

import "fmt"

// ErrAliasAlreadyRegistered is returned if the alias being registered is already taken.
type ErrAliasAlreadyRegistered struct {
	Alias string
}

func (e ErrAliasAlreadyRegistered) Error() string {
	return fmt.Sprintf("Alias Already Registered: %s", e.Alias)
}

// ErrAliasInvalid is returned when the alias includes unsupported characters.
type ErrAliasInvalid struct {
	Alias string
}

func (e ErrAliasInvalid) Error() string {
	return fmt.Sprintf("Alias Invalid: %s", e.Alias)
}

// ErrAliasNotFound is returned if the alias cannot be found.
type ErrAliasNotFound struct {
}

func (e ErrAliasNotFound) Error() string {
	return fmt.Sprintf("Could Not Find Alias For Public Key")
}

// ErrAliasNotPublic is returned when registering an encrypted alias.
type ErrAliasNotPublic struct {
}

func (e ErrAliasNotPublic) Error() string {
	return fmt.Sprintf("Cannot Register Private Alias")
}

// ErrAliasTooLong is returned when the alias is too long.
type ErrAliasTooLong struct {
	Size, Max int
}

func (e ErrAliasTooLong) Error() string {
	return fmt.Sprintf("Alias Too Long: %d Maximum: %d", e.Size, e.Max)
}

// ErrAliasTooShort is returned when the alias is too short.
type ErrAliasTooShort struct {
	Size, Min int
}

func (e ErrAliasTooShort) Error() string {
	return fmt.Sprintf("Alias Too Short: %d Minimum: %d", e.Size, e.Min)
}

// ErrPublicKeyNotFound is return if the public key cannot be found.
type ErrPublicKeyNotFound struct {
	Alias string
}

func (e ErrPublicKeyNotFound) Error() string {
	return fmt.Sprintf("Could Not Find Public Key For Alias: %s", e.Alias)
}
