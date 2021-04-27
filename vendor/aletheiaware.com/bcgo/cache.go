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

type Cache interface {
	Head(channel string) (*Reference, error)
	Block(hash []byte) (*Block, error)
	BlockEntries(channel string, timestamp uint64) ([]*BlockEntry, error)
	BlockContainingRecord(channel string, hash []byte) (*Block, error)
	PutHead(channel string, reference *Reference) error
	PutBlock(hash []byte, block *Block) error
	PutBlockEntry(channel string, entry *BlockEntry) error
	//DeleteBlock(hash []byte) error
}
