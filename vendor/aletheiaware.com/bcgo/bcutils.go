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

// Package containing utilities for BC in Go
package bcgo

import (
	"aletheiaware.com/cryptogo"
	"bufio"
	"errors"
	"fmt"
	"github.com/golang/protobuf/proto"
	"io"
	"io/ioutil"
	"log"
	"math/bits"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	BC_HOST      = "bc.aletheiaware.com"
	BC_HOST_TEST = "test-bc.aletheiaware.com"

	BETA_FLAG = "BETA"
	LIVE_FLAG = "LIVE"
)

func Ones(data []byte) uint64 {
	var count uint64
	for _, x := range data {
		count += uint64(bits.OnesCount(uint(x)))
	}
	return count
}

func BinarySizeToString(size uint64) string {
	if size == 1 {
		return "1Byte"
	}
	if size <= 1024 {
		return fmt.Sprintf("%dBytes", size)
	}
	var unit string
	s := float64(size)
	if s >= 1024 {
		s = s / 1024
		unit = "KiB"
	}
	if s >= 1024 {
		s = s / 1024
		unit = "MiB"
	}
	if s >= 1024 {
		s = s / 1024
		unit = "GiB"
	}
	if s >= 1024 {
		s = s / 1024
		unit = "TiB"
	}
	if s >= 1024 {
		s = s / 1024
		unit = "PiB"
	}
	return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.2f", s), "0"), ".") + unit
}

func DecimalSizeToString(size uint64) string {
	if size == 1 {
		return "1Byte"
	}
	if size <= 1000 {
		return fmt.Sprintf("%dBytes", size)
	}
	var unit string
	s := float64(size)
	if s >= 1000 {
		s = s / 1000
		unit = "KB"
	}
	if s >= 1000 {
		s = s / 1000
		unit = "MB"
	}
	if s >= 1000 {
		s = s / 1000
		unit = "GB"
	}
	if s >= 1000 {
		s = s / 1000
		unit = "TB"
	}
	if s >= 1000 {
		s = s / 1000
		unit = "PB"
	}
	return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.2f", s), "0"), ".") + unit
}

func Timestamp() uint64 {
	return uint64(time.Now().UnixNano())
}

func TimestampToString(timestamp uint64) string {
	return time.Unix(0, int64(timestamp)).UTC().Format("2006-01-02 15:04:05")
}

func MoneyToString(currency string, amount int64) string {
	if amount == 0 {
		return "Free"
	}
	s := "?"
	switch currency {
	case "usd":
		s = fmt.Sprintf("$%.2f", float64(amount)/100.0)
	}
	return strings.TrimRight(strings.TrimRight(s, "0"), ".")
}

func IsLive() bool {
	return BooleanFlag(LIVE_FLAG)
}

func IsBeta() bool {
	return BooleanFlag(BETA_FLAG)
}

func BooleanFlag(name string) bool {
	flag, ok := os.LookupEnv(name)
	if !ok {
		return false
	}
	b, err := strconv.ParseBool(flag)
	if err != nil {
		return false
	}
	return b
}

func BCHost() string {
	if IsLive() {
		return BC_HOST
	}
	return BC_HOST_TEST
}

func BCWebsite() string {
	return "https://" + BCHost()
}

func Alias() (string, error) {
	alias, ok := os.LookupEnv("ALIAS")
	if !ok {
		u, err := user.Current()
		if err != nil {
			return "", err
		}
		alias = u.Username
	}
	return alias, nil
}

func RootDirectory() (string, error) {
	root, ok := os.LookupEnv("ROOT_DIRECTORY")
	if !ok {
		u, err := user.Current()
		if err != nil {
			return "", err
		}
		root = filepath.Join(u.HomeDir, "bc")
		if _, err := os.Stat(root); os.IsNotExist(err) {
			root = RootDirectoryForUser(u)
		}
	}
	return root, nil
}

func KeyDirectory(directory string) (string, error) {
	keystore, ok := os.LookupEnv("KEYS_DIRECTORY")
	if !ok {
		keystore = filepath.Join(directory, "keys")
	}
	if err := os.MkdirAll(keystore, os.ModePerm); err != nil {
		return "", err
	}
	return keystore, nil
}

func CacheDirectory(directory string) (string, error) {
	cache, ok := os.LookupEnv("CACHE_DIRECTORY")
	if !ok {
		cache = filepath.Join(directory, "cache")
	}
	return cache, nil
}

func CertificateDirectory(directory string) (string, error) {
	certs, ok := os.LookupEnv("CERTIFICATE_DIRECTORY")
	if !ok {
		certs = filepath.Join(directory, "certificates")
	}
	return certs, nil
}

func LoadConfig() error {
	// Load Local Config
	directory, err := os.Getwd()
	if err != nil {
		return err
	}
	if err := ReadConfig(directory); err != nil {
		return err
	}
	// Load Root Config
	directory, err = RootDirectory()
	if err != nil {
		return err
	}
	if err := ReadConfig(directory); err != nil {
		return err
	}
	return nil
}

func ReadConfig(directory string) error {
	path := filepath.Join(directory, "config")
	if _, err := os.Stat(path); err == nil {
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			paths := strings.SplitN(scanner.Text(), "=", 2)
			if len(paths) > 1 {
				key, value := paths[0], paths[1]
				_, ok := os.LookupEnv(key)
				if !ok {
					os.Setenv(key, value)
				}
			}
		}
	}
	return nil
}

func SetupLogging(directory string) (*os.File, error) {
	store, ok := os.LookupEnv("LOG_DIRECTORY")
	if !ok {
		store = filepath.Join(directory, "logs")
	}
	if err := os.MkdirAll(store, os.ModePerm); err != nil {
		return nil, err
	}
	logFile, err := os.OpenFile(filepath.Join(store, time.Now().UTC().Format(time.RFC3339)), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		return nil, err
	}
	log.SetOutput(io.MultiWriter(os.Stdout, logFile))
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	return logFile, nil
}

func Peers(directory string) ([]string, error) {
	env, ok := os.LookupEnv("PEERS")
	if ok {
		return SplitRemoveEmpty(env, ","), nil
	} else {
		filename := "test-peers"
		if IsLive() {
			filename = "peers"
		}
		filepath := filepath.Join(directory, filename)
		if _, err := os.Stat(filepath); os.IsNotExist(err) {
			return []string{}, nil
		}

		data, err := ioutil.ReadFile(filepath)
		if err != nil {
			return nil, err
		}

		return SplitRemoveEmpty(string(data), "\n"), nil
	}
}

func SplitRemoveEmpty(s, sep string) []string {
	var result []string
	for _, s := range strings.Split(s, sep) {
		s = strings.TrimSpace(s)
		if s != "" {
			result = append(result, s)
		}
	}
	return result
}

func AddPeer(directory, peer string) error {
	filename := "test-peers"
	if IsLive() {
		filename = "peers"
	}
	file, err := os.OpenFile(filepath.Join(directory, filename), os.O_APPEND|os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return err
	}
	defer file.Close()
	if _, err := file.WriteString(peer + "\n"); err != nil {
		return err
	}
	return nil
}

// Chunk the data from reader into individual records with their own secret key and access list
func CreateRecords(creator Account, access []Identity, references []*Reference, reader io.Reader, callback func([]byte, *Record) error) (int, error) {
	payload := make([]byte, MAX_PAYLOAD_SIZE_BYTES)
	size := 0
	for {
		count, err := reader.Read(payload)
		if err != nil {
			if err == io.EOF {
				// Ignore EOFs
				break
			} else {
				return 0, err
			}
		}
		size = size + count
		key, record, err := CreateRecord(Timestamp(), creator, access, references, payload[:count])
		if err != nil {
			return 0, err
		}
		if err := callback(key, record); err != nil {
			return 0, err
		}
	}
	return size, nil
}

func CreateRecord(timestamp uint64, creator Account, access []Identity, references []*Reference, payload []byte) ([]byte, *Record, error) {
	size := uint64(len(payload))
	if size > MAX_PAYLOAD_SIZE_BYTES {
		return nil, nil, errors.New("Payload too large: " + BinarySizeToString(size) + " max: " + BinarySizeToString(MAX_PAYLOAD_SIZE_BYTES))
	}

	// Create record
	record := &Record{
		Timestamp: timestamp,
		Creator:   creator.Alias(),
		Reference: references,
	}

	// Encrypt payload if access map is not empty
	var key []byte
	var err error
	if len(access) > 0 {
		key, err = cryptogo.GenerateRandomKey(cryptogo.AES_256_KEY_SIZE_BYTES)
		if err != nil {
			return nil, nil, err
		}

		payload, err = cryptogo.EncryptAESGCM(key, payload)
		if err != nil {
			return nil, nil, err
		}

		// Grant access to each public key
		for _, a := range access {
			encrypted, algorithm, err := a.EncryptKey(key)
			if err != nil {
				return nil, nil, err
			}
			record.Access = append(record.Access, &Record_Access{
				Alias:               a.Alias(),
				SecretKey:           encrypted,
				EncryptionAlgorithm: algorithm,
			})
		}
		record.EncryptionAlgorithm = cryptogo.EncryptionAlgorithm_AES_256_GCM_NOPADDING
	} else {
		record.EncryptionAlgorithm = cryptogo.EncryptionAlgorithm_UNKNOWN_ENCRYPTION
	}

	signature, algorithm, err := creator.Sign(payload)
	if err != nil {
		return nil, nil, err
	}

	// Set payload and signature
	record.Payload = payload
	record.Signature = signature
	record.SignatureAlgorithm = algorithm

	if l, ok := os.LookupEnv(LIVE_FLAG); ok {
		record.Meta = map[string]string{
			LIVE_FLAG: l,
		}
	}

	return key, record, nil
}

func ReadDelimitedProtobuf(reader *bufio.Reader, destination proto.Message) error {
	data := make([]byte, 32)
	n, err := reader.Read(data[:])
	if err != nil {
		return err
	}
	if n <= 0 {
		return errors.New("Could not read data")
	}
	size, s := proto.DecodeVarint(data[:])
	if s <= 0 {
		return errors.New("Could not read size")
	}
	if size > MAX_BLOCK_SIZE_BYTES {
		return fmt.Errorf("Protobuf too large: %d max: %d", size, MAX_BLOCK_SIZE_BYTES)
	}

	// Calculate data received
	count := uint64(n - s)
	if count >= size {
		// All data in data[s:n]
		if err = proto.Unmarshal(data[s:s+int(size)], destination); err != nil {
			return err
		}
	} else {
		// More data in reader
		// Create new larger buffer
		buffer := make([]byte, size)
		// Copy data into new buffer
		copy(buffer[:count], data[s:n])
		// Read addition bytes
		for count < size {
			n, err := reader.Read(buffer[count:])
			if err != nil {
				if err == io.EOF {
					// Ignore EOFs, keep trying to read until count == size
				} else {
					return err
				}
			}
			if n <= 0 {
				return errors.New("Could not read data")
			}
			count = count + uint64(n)
		}

		if err = proto.Unmarshal(buffer, destination); err != nil {
			return err
		}
	}
	return nil
}

func WriteDelimitedProtobuf(writer *bufio.Writer, source proto.Message) error {
	size := uint64(proto.Size(source))
	if size > MAX_BLOCK_SIZE_BYTES {
		return errors.New("Protobuf too large: " + BinarySizeToString(size) + " max: " + BinarySizeToString(MAX_BLOCK_SIZE_BYTES))
	}

	data, err := proto.Marshal(source)
	if err != nil {
		return err
	}
	// Write request size varint
	if _, err := writer.Write(proto.EncodeVarint(size)); err != nil {
		return err
	}
	// Write request data
	if _, err = writer.Write(data); err != nil {
		return err
	}
	// Flush writer
	return writer.Flush()
}
