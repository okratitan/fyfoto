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

package cache

import (
	"aletheiaware.com/bcgo"
	"encoding/base64"
	"github.com/golang/protobuf/proto"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"strings"
)

type FileSystem struct {
	Directory string
}

func NewFileSystem(directory string) (*FileSystem, error) {
	// Create Block Cache
	if err := os.MkdirAll(path.Join(directory, "block"), os.ModePerm); err != nil {
		return nil, err
	}
	// Create Channel Cache
	if err := os.MkdirAll(path.Join(directory, "channel"), os.ModePerm); err != nil {
		return nil, err
	}
	return &FileSystem{
		Directory: directory,
	}, nil
}

func (f *FileSystem) Block(hash []byte) (*bcgo.Block, error) {
	// Read from file
	data, err := ioutil.ReadFile(path.Join(f.Directory, "block", base64.RawURLEncoding.EncodeToString(hash)))
	if err != nil {
		return nil, err
	}
	// Unmarshal into block
	block := &bcgo.Block{}
	if err = proto.Unmarshal(data[:], block); err != nil {
		return nil, err
	}
	return block, nil
}

func (f *FileSystem) BlockEntries(channel string, timestamp uint64) ([]*bcgo.BlockEntry, error) {
	directory := path.Join(f.Directory, "entry", base64.RawURLEncoding.EncodeToString([]byte(channel)))
	// Read directory
	files, err := ioutil.ReadDir(directory)
	if err != nil {
		return nil, err
	}
	var entries []*bcgo.BlockEntry
	for _, f := range files {
		t, err := strconv.ParseUint(f.Name(), 10, 64)
		if err != nil {
			return nil, err
		}
		if t >= timestamp {
			// Read from file
			data, err := ioutil.ReadFile(path.Join(directory, f.Name()))
			if err != nil {
				return nil, err
			}
			// Unmarshal into entry
			entry := &bcgo.BlockEntry{}
			if err = proto.Unmarshal(data[:], entry); err != nil {
				return nil, err
			}
			entries = append(entries, entry)
		}
	}
	return entries, nil
}

func (f *FileSystem) BlockContainingRecord(channel string, hash []byte) (*bcgo.Block, error) {
	data, err := ioutil.ReadFile(path.Join(f.Directory, "mapping", base64.RawURLEncoding.EncodeToString([]byte(channel)), base64.RawURLEncoding.EncodeToString(hash)))
	if err != nil {
		return nil, err
	}
	return f.Block(data)
}

func (f *FileSystem) Head(channel string) (*bcgo.Reference, error) {
	// Read from file
	data, err := ioutil.ReadFile(path.Join(f.Directory, "channel", base64.RawURLEncoding.EncodeToString([]byte(channel))))
	if err != nil {
		return nil, err
	}
	// Unmarshal into reference
	reference := &bcgo.Reference{}
	if err = proto.Unmarshal(data[:], reference); err != nil {
		return nil, err
	}
	return reference, nil
}

func (f *FileSystem) PutBlock(hash []byte, block *bcgo.Block) error {
	directory := path.Join(f.Directory, "mapping", base64.RawURLEncoding.EncodeToString([]byte(block.ChannelName)))
	// Create mapping directory
	if err := os.MkdirAll(directory, os.ModePerm); err != nil {
		return err
	}
	// Add record -> block mapping
	for _, e := range block.Entry {
		if err := ioutil.WriteFile(path.Join(directory, base64.RawURLEncoding.EncodeToString(e.RecordHash)), hash, os.ModePerm); err != nil {
			return err
		}
	}
	// Marshal into byte array
	data, err := proto.Marshal(block)
	if err != nil {
		return err
	}
	// Write to file
	return ioutil.WriteFile(path.Join(f.Directory, "block", base64.RawURLEncoding.EncodeToString(hash)), data, os.ModePerm)
}

func (f *FileSystem) PutBlockEntry(channel string, entry *bcgo.BlockEntry) error {
	// Marshal into byte array
	data, err := proto.Marshal(entry)
	if err != nil {
		return err
	}
	directory := path.Join(f.Directory, "entry", base64.RawURLEncoding.EncodeToString([]byte(channel)))
	// Create directory
	if err := os.MkdirAll(directory, os.ModePerm); err != nil {
		return err
	}
	// Write to file
	return ioutil.WriteFile(path.Join(directory, strconv.FormatUint(entry.Record.Timestamp, 10)), data, os.ModePerm)
}

func (f *FileSystem) PutHead(channel string, reference *bcgo.Reference) error {
	// Marshal into byte array
	data, err := proto.Marshal(reference)
	if err != nil {
		return err
	}
	// Write to file
	return ioutil.WriteFile(path.Join(f.Directory, "channel", base64.RawURLEncoding.EncodeToString([]byte(channel))), data, os.ModePerm)
}

// func (f *FileSystem) DeleteBlock(hash []byte) error {
// 	// Delete file
// 	return os.Remove(path.Join(f.Directory, "block", base64.RawURLEncoding.EncodeToString(hash)))
// }

func (f *FileSystem) MeasureStorageUsage(prefix string) (map[string]uint64, error) {
	usage := make(map[string]uint64)
	files, err := ioutil.ReadDir(path.Join(f.Directory, "block"))
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		hash, err := base64.RawURLEncoding.DecodeString(file.Name())
		if err != nil {
			return nil, err
		}
		block, err := f.Block(hash)
		if err != nil {
			return nil, err
		}
		if strings.HasPrefix(block.ChannelName, prefix) {
			for _, entry := range block.Entry {
				creator := entry.Record.Creator
				u, ok := usage[creator]
				if !ok {
					u = 0
				}
				u += uint64(proto.Size(entry))
				usage[creator] = u
			}
		}
	}
	return usage, nil
}
