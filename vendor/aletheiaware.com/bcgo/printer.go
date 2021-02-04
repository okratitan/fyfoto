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

import (
	"encoding/base64"
	"fmt"
	"io"
)

type PrintingMiningListener struct {
	Output io.Writer
}

func (p *PrintingMiningListener) OnMiningStarted(channel *Channel, size uint64) {
	fmt.Fprintf(p.Output, "Mining %s %s\n", channel.Name, BinarySizeToString(size))
}

func (p *PrintingMiningListener) OnNewMaxOnes(channel *Channel, nonce, ones uint64) {
	fmt.Fprintf(p.Output, "Mining %s %d %d/512\n", channel.Name, nonce, ones)
}

func (p *PrintingMiningListener) OnMiningThresholdReached(channel *Channel, hash []byte, block *Block) {
	fmt.Fprintf(p.Output, "Mined %s %s %s\n", channel.Name, TimestampToString(block.Timestamp), base64.RawURLEncoding.EncodeToString(hash))
}

/*
func GetAndPrintURL(url string) {
	response, err := http.Get(url)
	if err != nil {
		log.Println(err)
		return
	}
	log.Println(response)
	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Println(err)
		return
	}
	log.Println(string(data))
}
*/

func PrintReference(output io.Writer, prefix string, reference *Reference) {
	fmt.Fprintf(output, "%sTimestamp: %d\n", prefix, reference.Timestamp)
	fmt.Fprintf(output, "%sChannelName: %s\n", prefix, reference.ChannelName)
	fmt.Fprintf(output, "%sBlockHash: %s\n", prefix, base64.RawURLEncoding.EncodeToString(reference.BlockHash))
	fmt.Fprintf(output, "%sRecordHash: %s\n", prefix, base64.RawURLEncoding.EncodeToString(reference.RecordHash))
}

func PrintBlock(output io.Writer, prefix string, hash []byte, block *Block) {
	fmt.Fprintf(output, "%sHash: %s\n", prefix, base64.RawURLEncoding.EncodeToString(hash))
	fmt.Fprintf(output, "%sTimestamp: %d\n", prefix, block.Timestamp)
	fmt.Fprintf(output, "%sChannelName: %s\n", prefix, block.ChannelName)
	fmt.Fprintf(output, "%sLength: %d\n", prefix, block.Length)
	fmt.Fprintf(output, "%sPrevious: %s\n", prefix, base64.RawURLEncoding.EncodeToString(block.Previous))
	fmt.Fprintf(output, "%sMiner: %s\n", prefix, block.Miner)
	fmt.Fprintf(output, "%sNonce: %d\n", prefix, block.Nonce)
	fmt.Fprintf(output, "%sEntries: %d\n", prefix, len(block.Entry))
	for i, entry := range block.Entry {
		fmt.Fprintf(output, "%sEntry: %d\n", prefix, i)
		PrintBlockEntry(output, prefix+"\t", entry)
	}
}

func PrintBlockEntry(output io.Writer, prefix string, entry *BlockEntry) {
	PrintRecord(output, prefix, entry.RecordHash, entry.Record)
}

func PrintRecord(output io.Writer, prefix string, hash []byte, record *Record) {
	fmt.Fprintf(output, "%sHash: %s\n", prefix, base64.RawURLEncoding.EncodeToString(hash))
	if record == nil {
		fmt.Fprintf(output, "%sRecord: <nil>\n", prefix)
	} else {
		fmt.Fprintf(output, "%sTimestamp: %d\n", prefix, record.Timestamp)
		fmt.Fprintf(output, "%sCreator: %s\n", prefix, record.Creator)
		for j, access := range record.Access {
			fmt.Fprintf(output, "%sAccess: %d\n", prefix, j)
			fmt.Fprintf(output, "%s\tAlias: %s\n", prefix, access.Alias)
			fmt.Fprintf(output, "%s\tSecretKey: %s\n", prefix, base64.RawURLEncoding.EncodeToString(access.SecretKey))
			fmt.Fprintf(output, "%s\tKeyEncryptionAlgorithm: %s\n", prefix, access.EncryptionAlgorithm)
		}
		fmt.Fprintf(output, "%sPayload: %s\n", prefix, base64.RawURLEncoding.EncodeToString(record.Payload))
		fmt.Fprintf(output, "%sCompressionAlgorithm: %s\n", prefix, record.CompressionAlgorithm)
		fmt.Fprintf(output, "%sEncryptionAlgorithm: %s\n", prefix, record.EncryptionAlgorithm)
		fmt.Fprintf(output, "%sSignature: %s\n", prefix, base64.RawURLEncoding.EncodeToString(record.Signature))
		fmt.Fprintf(output, "%sSignatureAlgorithm: %s\n", prefix, record.SignatureAlgorithm)
		fmt.Fprintf(output, "%sReferences: %d\n", prefix, len(record.Reference))
		for k, reference := range record.Reference {
			fmt.Fprintf(output, "%sReference: %d\n", prefix, k)
			PrintReference(output, prefix+"\t", reference)
		}
	}
}
