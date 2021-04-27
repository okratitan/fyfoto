/*
 * Copyright 2019-21 Aletheia Ware LLC
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

package network

import (
	"aletheiaware.com/bcgo"
	"bufio"
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net"
	"sort"
	"strconv"
	"sync"
	"time"
)

const (
	MAX_TCP_ERRORS = 10
	TIMEOUT        = 3 * time.Minute
	PORT_CONNECT   = 22022
	PORT_GET_BLOCK = 22222
	PORT_GET_HEAD  = 22322
	PORT_BROADCAST = 23232
)

type TCP struct {
	DialTimeout time.Duration
	GetTimeout  time.Duration
	lock        sync.RWMutex
	peers       map[string]int
}

func NewTCP(peers ...string) *TCP {
	t := &TCP{
		DialTimeout: TIMEOUT,
		GetTimeout:  TIMEOUT,
		peers:       make(map[string]int),
	}
	for _, p := range peers {
		t.peers[p] = 0
	}
	return t
}

func (t *TCP) AddPeer(peer string) {
	t.lock.Lock()
	defer t.lock.Unlock()
	t.ensurePeerMap()
	t.peers[peer] = 0
}

func (t *TCP) SetPeers(peers ...string) {
	t.lock.Lock()
	defer t.lock.Unlock()
	ps := make(map[string]int, len(peers))
	if t.peers != nil {
		// Copy errors into new map
		for _, p := range peers {
			e, ok := t.peers[p]
			if !ok {
				e = 0
			}
			ps[p] = e
		}
	}
	t.peers = ps
}

func (t *TCP) Connect(peer string, data []byte) error {
	address := net.JoinHostPort(peer, strconv.Itoa(PORT_CONNECT))
	dialer := &net.Dialer{Timeout: t.DialTimeout}
	connection, err := dialer.Dial("tcp", address)
	if err != nil {
		return fmt.Errorf("%s: %v", peer, err)
	}
	defer connection.Close()
	writer := bufio.NewWriter(connection)
	if _, err := writer.Write(data); err != nil {
		return fmt.Errorf("%s: %v", peer, err)
	}
	if err := writer.Flush(); err != nil {
		return fmt.Errorf("%s: %v", peer, err)
	}
	t.AddPeer(peer)
	return nil
}

func (t *TCP) Head(channel string) (*bcgo.Reference, error) {
	for _, peer := range t.bestPeers() {
		log.Println("Requesting", channel, "from", peer)
		address := net.JoinHostPort(peer, strconv.Itoa(PORT_GET_HEAD))
		dialer := &net.Dialer{Timeout: t.DialTimeout}
		connection, err := dialer.Dial("tcp", address)
		if err != nil {
			t.error(peer, err)
			continue
		}
		defer connection.Close()
		writer := bufio.NewWriter(connection)
		if err := bcgo.WriteDelimitedProtobuf(writer, &bcgo.Reference{
			ChannelName: channel,
		}); err != nil {
			t.error(peer, err)
			continue
		}
		reader := bufio.NewReader(connection)
		reference := &bcgo.Reference{}
		if err := bcgo.ReadDelimitedProtobuf(reader, reference); err != nil {
			if err != io.EOF {
				t.error(peer, err)
			}
			continue
		} else {
			log.Println("Received", bcgo.TimestampToString(reference.Timestamp), reference.ChannelName, base64.RawURLEncoding.EncodeToString(reference.BlockHash), "from", peer)
			return reference, nil
		}
	}
	return nil, fmt.Errorf("Could not get %s from peers", channel)
}

func (t *TCP) Block(reference *bcgo.Reference) (*bcgo.Block, error) {
	for _, peer := range t.bestPeers() {
		log.Println("Requesting", reference.ChannelName, base64.RawURLEncoding.EncodeToString(reference.BlockHash), base64.RawURLEncoding.EncodeToString(reference.RecordHash), "from", peer)
		address := net.JoinHostPort(peer, strconv.Itoa(PORT_GET_BLOCK))
		dialer := &net.Dialer{Timeout: t.DialTimeout}
		connection, err := dialer.Dial("tcp", address)
		if err != nil {
			t.error(peer, err)
			continue
		}
		defer connection.Close()
		writer := bufio.NewWriter(connection)
		if err := bcgo.WriteDelimitedProtobuf(writer, reference); err != nil {
			t.error(peer, err)
			continue
		}
		reader := bufio.NewReader(connection)
		block := &bcgo.Block{}
		if err := bcgo.ReadDelimitedProtobuf(reader, block); err != nil {
			if err != io.EOF {
				t.error(peer, err)
			}
			continue
		} else {
			log.Println("Received", bcgo.TimestampToString(block.Timestamp), block.ChannelName, "from", peer)
			return block, nil
		}
	}
	return nil, fmt.Errorf("Could not get %s block from peers", reference.ChannelName)
}

func (t *TCP) Broadcast(channel bcgo.Channel, cache bcgo.Cache, hash []byte, block *bcgo.Block) error {
	var last error
	for _, peer := range t.bestPeers() {
		last = nil
		log.Println("Broadcasting", channel, base64.RawURLEncoding.EncodeToString(hash), "to", peer)
		address := net.JoinHostPort(peer, strconv.Itoa(PORT_BROADCAST))
		dialer := &net.Dialer{Timeout: t.DialTimeout}
		connection, err := dialer.Dial("tcp", address)
		if err != nil {
			last = err
			t.error(peer, err)
			continue
		}
		defer connection.Close()
		writer := bufio.NewWriter(connection)
		reader := bufio.NewReader(connection)

		for {
			if err := bcgo.WriteDelimitedProtobuf(writer, block); err != nil {
				return err
			}
			reference := &bcgo.Reference{}
			if err := bcgo.ReadDelimitedProtobuf(reader, reference); err != nil {
				if err == io.EOF {
					// Ignore
					break
				}
				return err
			}

			remote := reference.BlockHash
			if bytes.Equal(hash, remote) {
				// Broadcast accepted
				log.Println("Broadcast to", peer, "succeeded")
				break
			} else {
				// Broadcast rejected
				referencedBlock, err := bcgo.LoadBlock(channel.Name(), cache, t, remote)
				if err != nil {
					return err
				}

				if referencedBlock.Length == block.Length {
					// Option A: remote points to a different chain of the same length, next chain to get a block mined on top wins
					log.Println("Broadcast to", peer, "failed: Option A")
					break
				} else if referencedBlock.Length > block.Length {
					// Option B: remote points to a longer chain
					log.Println("Broadcast to", peer, "failed: Option B")
					go func() {
						if err := channel.Pull(cache, t); err != nil {
							log.Println(err)
						}
					}()
					return bcgo.ErrChannelOutOfDate{Channel: channel.Name()}
					// TODO re-mine all dropped records into new blocks on top of new head
				} else {
					// Option C: remote points to a shorter chain, and cannot update because the chain cannot be verified or the host is missing some blocks
					log.Println("Broadcast to", peer, "failed: Option C")
					block = referencedBlock
				}
			}
		}
	}
	return last
}

// Returns the peer associated with the given address
func (t *TCP) PeerForAddress(address string) string {
	host, _, err := net.SplitHostPort(address)
	if err != nil {
		log.Println(address, err)
		return ""
	}
	hIP := net.ParseIP(host)
	peers := t.Peers()
	for _, p := range peers {
		if p == host {
			return p
		}
		if hIP != nil {
			// DNS lookup peer to get IP addresses
			ips, err := net.LookupIP(p)
			if err != nil {
				continue
			}
			for _, ip := range ips {
				if ip.Equal(hIP) {
					return p
				}
			}
		}
	}
	return ""
}

func (t *TCP) Peers() (peers []string) {
	t.lock.RLock()
	defer t.lock.RUnlock()
	if t.peers != nil {
		for p := range t.peers {
			if p == "" {
				continue
			}
			peers = append(peers, p)
		}
	}
	return
}

// Returns a slice of peers sorted by ascending error rate
func (t *TCP) bestPeers() []string {
	peers := t.Peers()
	if len(peers) == 0 {
		return peers
	}
	if t.peers != nil {
		t.lock.RLock()
		sort.Slice(peers, func(i, j int) bool {
			return t.peers[peers[i]] < t.peers[peers[j]]
		})
		t.lock.RUnlock()
	}
	return peers[:1+len(peers)/2] // return first half of peers ie least erroneous
}

func (t *TCP) ensurePeerMap() {
	if t.peers == nil {
		t.peers = make(map[string]int)
	}
}

func (t *TCP) error(peer string, err error) {
	t.lock.Lock()
	defer t.lock.Unlock()
	log.Println("Error:", peer, err)
	t.ensurePeerMap()
	count := t.peers[peer] + 1
	t.peers[peer] = count
	if count > MAX_TCP_ERRORS {
		log.Println(peer, "Exceeded MAX_TCP_ERRORS")
		delete(t.peers, peer)
	}
}
