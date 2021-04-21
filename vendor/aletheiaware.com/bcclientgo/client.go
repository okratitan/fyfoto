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

package bcclientgo

import (
	"aletheiaware.com/aliasgo"
	"aletheiaware.com/bcgo"
	"aletheiaware.com/bcgo/account"
	"aletheiaware.com/bcgo/cache"
	"aletheiaware.com/bcgo/channel"
	"aletheiaware.com/bcgo/network"
	"aletheiaware.com/bcgo/node"
	"aletheiaware.com/cryptogo"
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"reflect"
)

type BCClient interface {
	Root() (string, error)
	Peers() []string
	Cache() (bcgo.Cache, error)
	Network() (bcgo.Network, error)
	Node() (bcgo.Node, error)
	Init(bcgo.MiningListener) (bcgo.Node, error)
	IsSignedIn() bool
	PublicKey(string) ([]byte, cryptogo.PublicKeyFormat, error)
	Head(string) ([]byte, error)
	Chain(string, func([]byte, *bcgo.Block) error) error
	Block(string, []byte) (*bcgo.Block, error)
	Record(string, []byte) (*bcgo.Record, error)
	Read(string, []byte, []byte, io.Writer) error
	ReadKey(string, []byte, []byte, io.Writer) error
	ReadPayload(string, []byte, []byte, io.Writer) error
	Write(string, []string, io.Reader) (int, error)
	Mine(string, uint64, bcgo.MiningListener) ([]byte, error)
	Pull(string) error
	Push(string) error
	Purge() error
	ImportKeys(string, string, string) error
	ExportKeys(string, string, []byte) (string, error)
	SetRoot(string)
	SetPeers(...string)
	SetCache(bcgo.Cache)
	SetNetwork(bcgo.Network)
	SetNode(bcgo.Node)
}

type bcClient struct {
	root    string
	peers   []string
	cache   bcgo.Cache
	network bcgo.Network
	node    bcgo.Node
}

func NewBCClient(peers ...string) BCClient {
	/* TODO try loading from file system
		rootDir, err := c.Root()
		if err == nil {
			bcgo.Peers(rootDir)
	}
	*/
	if len(peers) == 0 {
		peers = append(peers,
			bcgo.BCHost(), // Add BC host as peer
		)
	}
	return &bcClient{
		peers: peers,
	}
}

func (c *bcClient) Peers() []string {
	return c.peers
}

func (c *bcClient) Cache() (bcgo.Cache, error) {
	if c.cache == nil {
		rootDir, err := c.Root()
		if err != nil {
			return nil, err
		}
		cacheDir, err := bcgo.CacheDirectory(rootDir)
		if err != nil {
			return nil, fmt.Errorf("Could not get cache directory: %s", err.Error())
		}
		cache, err := cache.NewFileSystem(cacheDir)
		if err != nil {
			return nil, fmt.Errorf("Could not create file cache: %s", err.Error())
		}
		c.cache = cache
	}
	return c.cache, nil
}

func (c *bcClient) Network() (bcgo.Network, error) {
	if c.network == nil || reflect.ValueOf(c.network).IsNil() {
		c.network = network.NewTCP(c.Peers()...)
	}
	return c.network, nil
}

func (c *bcClient) Node() (bcgo.Node, error) {
	if !c.IsSignedIn() {
		rootDir, err := c.Root()
		if err != nil {
			return nil, err
		}
		cache, err := c.Cache()
		if err != nil {
			return nil, err
		}
		network, err := c.Network()
		if err != nil {
			return nil, err
		}
		account, err := account.LoadRSA(rootDir)
		if err != nil {
			return nil, err
		}
		c.node = node.New(account, cache, network)
	}
	return c.node, nil
}

func (c *bcClient) SetRoot(root string) {
	c.root = root
}

func (c *bcClient) SetPeers(peers ...string) {
	c.peers = peers
	if c.network == nil || reflect.ValueOf(c.network).IsNil() {
		return
	}
	if n, ok := c.network.(*network.TCP); ok {
		n.SetPeers(peers...)
	}
}

func (c *bcClient) SetCache(cache bcgo.Cache) {
	c.cache = cache
}

func (c *bcClient) SetNetwork(network bcgo.Network) {
	c.network = network
}

func (c *bcClient) SetNode(node bcgo.Node) {
	c.node = node
}

func (c *bcClient) Init(listener bcgo.MiningListener) (bcgo.Node, error) {
	// TODO check Alias is unique

	// Create Node
	node, err := c.Node()
	if err != nil {
		return nil, err
	}

	// Register Alias
	if err := aliasgo.Register(node, listener); err != nil {
		return nil, err
	}

	return node, nil
}

func (c *bcClient) IsSignedIn() bool {
	return c.node != nil && !reflect.ValueOf(c.node).IsNil()
}

func (c *bcClient) PublicKey(alias string) ([]byte, cryptogo.PublicKeyFormat, error) {
	cache, err := c.Cache()
	if err != nil {
		return nil, cryptogo.PublicKeyFormat_UNKNOWN_PUBLIC_KEY_FORMAT, err
	}
	network, err := c.Network()
	if err != nil {
		return nil, cryptogo.PublicKeyFormat_UNKNOWN_PUBLIC_KEY_FORMAT, err
	}
	// Open Alias Channel
	aliases := aliasgo.OpenAliasChannel()
	if err := aliases.Refresh(cache, network); err != nil {
		log.Println(err)
	}
	// Get Public Key for Alias
	publicKey, err := aliasgo.PublicKeyForAlias(aliases, cache, network, alias)
	if err != nil {
		return nil, cryptogo.PublicKeyFormat_UNKNOWN_PUBLIC_KEY_FORMAT, err
	}
	publicKeyBytes, err := cryptogo.RSAPublicKeyToPKIXBytes(publicKey)
	if err != nil {
		return nil, cryptogo.PublicKeyFormat_UNKNOWN_PUBLIC_KEY_FORMAT, err
	}
	return publicKeyBytes, cryptogo.PublicKeyFormat_PKIX, nil
}

func (c *bcClient) Head(name string) ([]byte, error) {
	cache, err := c.Cache()
	if err != nil {
		return nil, err
	}
	network, err := c.Network()
	if err != nil {
		return nil, err
	}
	ch := channel.New(name)
	if err := ch.Load(cache, network); err != nil {
		return nil, err
	}
	return ch.Head(), nil
}

func (c *bcClient) Chain(name string, callback func([]byte, *bcgo.Block) error) error {
	cache, err := c.Cache()
	if err != nil {
		return err
	}
	network, err := c.Network()
	if err != nil {
		return err
	}
	ch := channel.New(name)
	if err := ch.Load(cache, network); err != nil {
		return err
	}
	return bcgo.Iterate(name, ch.Head(), nil, cache, network, callback)
}

func (c *bcClient) Block(name string, hash []byte) (*bcgo.Block, error) {
	cache, err := c.Cache()
	if err != nil {
		return nil, err
	}
	network, err := c.Network()
	if err != nil {
		return nil, err
	}
	block, err := bcgo.LoadBlock(name, cache, network, hash)
	if err != nil {
		return nil, err
	}
	return block, nil
}

func (c *bcClient) Record(name string, hash []byte) (*bcgo.Record, error) {
	cache, err := c.Cache()
	if err != nil {
		return nil, err
	}
	network, err := c.Network()
	if err != nil {
		return nil, err
	}
	block, err := bcgo.LoadBlockContainingRecord(name, cache, network, hash)
	if err != nil {
		return nil, err
	}
	for _, entry := range block.Entry {
		if bytes.Equal(entry.RecordHash, hash) {
			return entry.Record, nil
		}
	}
	return nil, errors.New("Could not get block containing record")
}

func (c *bcClient) Read(name string, blockHash, recordHash []byte, output io.Writer) error {
	cache, err := c.Cache()
	if err != nil {
		return err
	}
	network, err := c.Network()
	if err != nil {
		return err
	}
	node, err := c.Node()
	if err != nil {
		return err
	}

	ch := channel.New(name)

	if err := ch.Load(cache, network); err != nil {
		log.Println(err)
	}

	if blockHash == nil {
		blockHash = ch.Head()
	}

	return bcgo.Read(name, blockHash, nil, cache, network, node.Account(), recordHash, func(entry *bcgo.BlockEntry, key, payload []byte) error {
		bcgo.PrintBlockEntry(output, "", entry)
		return nil
	})
}

func (c *bcClient) ReadKey(name string, blockHash, recordHash []byte, output io.Writer) error {
	cache, err := c.Cache()
	if err != nil {
		return err
	}
	network, err := c.Network()
	if err != nil {
		return err
	}
	node, err := c.Node()
	if err != nil {
		return err
	}

	ch := channel.New(name)

	if err := ch.Load(cache, network); err != nil {
		log.Println(err)
	}

	if blockHash == nil {
		blockHash = ch.Head()
	}

	return bcgo.ReadKey(name, blockHash, nil, cache, network, node.Account(), recordHash, func(key []byte) error {
		output.Write(key)
		return nil
	})
}

func (c *bcClient) ReadPayload(name string, blockHash, recordHash []byte, output io.Writer) error {
	cache, err := c.Cache()
	if err != nil {
		return err
	}
	network, err := c.Network()
	if err != nil {
		return err
	}
	node, err := c.Node()
	if err != nil {
		return err
	}

	ch := channel.New(name)

	if err := ch.Load(cache, network); err != nil {
		log.Println(err)
	}

	if blockHash == nil {
		blockHash = ch.Head()
	}

	return bcgo.Read(name, blockHash, nil, cache, network, node.Account(), recordHash, func(entry *bcgo.BlockEntry, key, payload []byte) error {
		output.Write(payload)
		return nil
	})
}

func (c *bcClient) Write(name string, accesses []string, input io.Reader) (int, error) {
	cache, err := c.Cache()
	if err != nil {
		return 0, err
	}
	network, err := c.Network()
	if err != nil {
		return 0, err
	}
	var access []bcgo.Identity

	if len(accesses) > 0 {
		// Open Alias Channel
		aliases := aliasgo.OpenAliasChannel()
		if err := aliases.Refresh(cache, network); err != nil {
			log.Println(err)
		}
		access = aliasgo.PublicKeysForAliases(aliases, cache, network, accesses)
	}

	node, err := c.Node()
	if err != nil {
		return 0, err
	}

	size, err := bcgo.CreateRecords(node.Account(), access, nil, input, func(key []byte, record *bcgo.Record) error {
		_, err := bcgo.WriteRecord(name, cache, record)
		return err
	})
	if err != nil {
		return 0, err
	}

	return size, nil
}

func (c *bcClient) Mine(name string, threshold uint64, listener bcgo.MiningListener) ([]byte, error) {
	cache, err := c.Cache()
	if err != nil {
		return nil, err
	}
	network, err := c.Network()
	if err != nil {
		return nil, err
	}
	node, err := c.Node()
	if err != nil {
		return nil, err
	}

	ch := channel.New(name)

	if err := ch.Load(cache, network); err != nil {
		log.Println(err)
	}

	hash, _, err := bcgo.Mine(node, ch, threshold, listener)
	if err != nil {
		return nil, err
	}
	return hash, nil
}

func (c *bcClient) Pull(name string) error {
	cache, err := c.Cache()
	if err != nil {
		return err
	}
	network, err := c.Network()
	if err != nil {
		return err
	}
	return channel.New(name).Pull(cache, network)
}

func (c *bcClient) Push(name string) error {
	cache, err := c.Cache()
	if err != nil {
		return err
	}
	network, err := c.Network()
	if err != nil {
		return err
	}
	ch := channel.New(name)
	if err := ch.Load(cache, nil); err != nil {
		return err
	}
	return ch.Push(cache, network)
}

func (c *bcClient) Purge() error {
	rootDir, err := c.Root()
	if err != nil {
		return err
	}
	// Get cache directory
	dir, err := bcgo.CacheDirectory(rootDir)
	if err != nil {
		return err
	}
	return os.RemoveAll(dir)
}

func (c *bcClient) ImportKeys(peer, alias, accessCode string) error {
	rootDir, err := c.Root()
	if err != nil {
		return err
	}
	// Get KeyStore
	keystore, err := bcgo.KeyDirectory(rootDir)
	if err != nil {
		return err
	}
	return cryptogo.ImportKeys(peer, keystore, alias, accessCode)
}

func (c *bcClient) ExportKeys(peer, alias string, password []byte) (string, error) {
	rootDir, err := c.Root()
	if err != nil {
		return "", err
	}
	// Get KeyStore
	keystore, err := bcgo.KeyDirectory(rootDir)
	if err != nil {
		return "", err
	}
	return cryptogo.ExportKeys(peer, keystore, alias, password)
}

func PrintNode(output io.Writer, node bcgo.Node) error {
	fmt.Fprintln(output, node.Account().Alias())
	bytes, format, err := node.Account().PublicKey()
	if err != nil {
		return err
	}
	fmt.Fprintln(output, base64.RawURLEncoding.EncodeToString(bytes), format)
	return nil
}

func rootDir() (string, error) {
	rootDir, err := bcgo.RootDirectory()
	if err != nil {
		return "", fmt.Errorf("Could not get root directory: %s\n", err.Error())
	}
	if err := bcgo.ReadConfig(rootDir); err != nil {
		log.Println("Error reading config:", err)
	}
	return rootDir, nil
}
