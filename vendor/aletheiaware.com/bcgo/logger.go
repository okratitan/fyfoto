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

import (
	"encoding/base64"
	"log"
)

type LoggingMiningListener struct {
}

func (l *LoggingMiningListener) OnMiningStarted(channel Channel, size uint64) {
	log.Printf("Mining %s %s\n", channel.Name(), BinarySizeToString(size))
}

func (l *LoggingMiningListener) OnNewMaxOnes(channel Channel, nonce, ones uint64) {
	log.Printf("Mining %s %d %d/512\n", channel.Name(), nonce, ones)
}

func (l *LoggingMiningListener) OnMiningThresholdReached(channel Channel, hash []byte, block *Block) {
	log.Printf("Mined %s %s %s\n", channel.Name(), TimestampToString(block.Timestamp), base64.RawURLEncoding.EncodeToString(hash))
}
