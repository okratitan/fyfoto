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

package bcgo

const (
	ERROR_CHANNEL_OUT_OF_DATE = "Channel out of date"
)

type Network interface {
	// Requests the head hash of the given channel
	GetHead(channel string) (*Reference, error)
	// Requests the block from the given reference
	GetBlock(reference *Reference) (*Block, error)
	// Broadcasts the channel update to the network
	Broadcast(channel *Channel, cache Cache, hash []byte, block *Block) error
}
