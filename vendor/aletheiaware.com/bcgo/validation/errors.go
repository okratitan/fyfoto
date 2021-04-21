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

package validation

import (
	"fmt"
	"strings"
)

// ErrDifferentLiveFlag is returned when a block's live flag doesn't match the current setting.
type ErrDifferentLiveFlag struct {
	Expected, Actual string
}

func (e ErrDifferentLiveFlag) Error() string {
	return fmt.Sprintf("Different Live Flag; Expected '%s', got '%s'", e.Expected, e.Actual)
}

// ErrDuplicateBlock is returned when a block appears multiple times in a chain.
type ErrDuplicateBlock struct {
	Hash string
}

func (e ErrDuplicateBlock) Error() string {
	return fmt.Sprintf("Duplicate Block: %s", e.Hash)
}

// ErrDuplicateEntry is returned when an entry appears multiple times in a chain.
type ErrDuplicateEntry struct {
	Hash string
}

func (e ErrDuplicateEntry) Error() string {
	return fmt.Sprintf("Duplicate Entry: %s", e.Hash)
}

// ErrHashTooWeak is returned when a block's hash doesn't meet the threshold.
type ErrHashTooWeak struct {
	Expected, Actual uint64
}

func (e ErrHashTooWeak) Error() string {
	return fmt.Sprintf("Hash doesn't meet Proof-of-Work threshold: %d vs %d", e.Expected, e.Actual)
}

// ErrMissingValidatedBlock is returned when a chain doesn't contain a block recorded in periodic validation chain.
type ErrMissingValidatedBlock struct {
	PVC, Channel string
	Missing      []string
}

func (e ErrMissingValidatedBlock) Error() string {
	return fmt.Sprintf("%s Missing Validated Block %s %s", e.PVC, e.Channel, strings.Join(e.Missing, ","))
}
