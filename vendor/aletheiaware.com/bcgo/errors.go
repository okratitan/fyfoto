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

import "fmt"

// ErrBlockHashIncorrect is returned when the given hash does not match the hash of the given block.
type ErrBlockHashIncorrect struct {
}

func (e ErrBlockHashIncorrect) Error() string {
	return "Hash doesn't match block hash"
}

// ErrChainInvalid is returned when a block fails validation for some reason.
type ErrChainInvalid struct {
	Reason string
}

func (e ErrChainInvalid) Error() string {
	return fmt.Sprintf("Chain invalid: %s", e.Reason)
}

// ErrChainTooShort is returned when a new chain is shorter the channel's current head.
type ErrChainTooShort struct {
	LengthA, LengthB uint64
}

func (e ErrChainTooShort) Error() string {
	return fmt.Sprintf("Chain too short to replace current head: %d vs %d", e.LengthA, e.LengthB)
}

// ErrChannelNameIncorrect is returned when a block channel name doesn't match the channel.
type ErrChannelNameIncorrect struct {
	Expected, Actual string
}

func (e ErrChannelNameIncorrect) Error() string {
	return fmt.Sprintf("Name doesn't match channel name: %s vs %s", e.Expected, e.Actual)
}

// ErrChannelNameInvalid is returned when a channel name is too long or includes unsupported characters.
type ErrChannelNameInvalid struct {
	Reason string
}

func (e ErrChannelNameInvalid) Error() string {
	return fmt.Sprintf("Name invalid: %s", e.Reason)
}

// ErrChannelOutOfDate is returned when a broadcast fails due to the network
// having a more up-to-date version of the channel.
type ErrChannelOutOfDate struct {
	Channel string
}

func (e ErrChannelOutOfDate) Error() string {
	return fmt.Sprintf("Channel out of date: %s", e.Channel)
}

// ErrBlockTooLarge is returned when the Block exceeds the size limit.
type ErrBlockTooLarge struct {
	Size, Max uint64
}

func (e ErrBlockTooLarge) Error() string {
	return fmt.Sprintf("Block too large: %s max: %s", BinarySizeToString(e.Size), BinarySizeToString(e.Max))
}

// ErrPayloadTooLarge is returned when the Payload exceeds the size limit.
type ErrPayloadTooLarge struct {
	Size, Max uint64
}

func (e ErrPayloadTooLarge) Error() string {
	return fmt.Sprintf("Payload too large: %s max: %s", BinarySizeToString(e.Size), BinarySizeToString(e.Max))
}

// ErrNoSuchChannel is returned when a channel with the given name cannot be found.
type ErrNoSuchChannel struct {
	Channel string
}

func (e ErrNoSuchChannel) Error() string {
	return fmt.Sprintf("No such channel: %s", e.Channel)
}

// ErrNoSuchBlock is returned when a block with the given hash cannot be found.
type ErrNoSuchBlock struct {
	Hash string
}

func (e ErrNoSuchBlock) Error() string {
	return fmt.Sprintf("No such block: %s", e.Hash)
}

// ErrNoSuchHead is returned when a head with the given name cannot be found.
type ErrNoSuchHead struct {
	Channel string
}

func (e ErrNoSuchHead) Error() string {
	return fmt.Sprintf("No such head: %s", e.Channel)
}

// ErrNoSuchMapping is returned when a mapping with the given hash cannot be found.
type ErrNoSuchMapping struct {
	Hash string
}

func (e ErrNoSuchMapping) Error() string {
	return fmt.Sprintf("No such record to block mapping: %s", e.Hash)
}

// ErrNoEntriesToMine is returned when a mining operation fails due to a lack of entries.
type ErrNoEntriesToMine struct {
	Channel string
}

func (e ErrNoEntriesToMine) Error() string {
	return fmt.Sprintf("No entries to mine for channel: %s", e.Channel)
}

// ErrNonceWrapAround is returned when a mining operation fails after attempting every possible nonce.
type ErrNonceWrapAround struct {
}

func (e ErrNonceWrapAround) Error() string {
	return "Nonce wrapped around before reaching threshold"
}

// ErrStopIteration is used by callbacks to indicate that their iteration
// through a channel should stop, typically because the result has been found.
type ErrStopIteration struct {
}

func (e ErrStopIteration) Error() string {
	return "Stop Iteration"
}
