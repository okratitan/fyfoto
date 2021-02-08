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
	"aletheiaware.com/cryptogo"
	"bytes"
	"crypto/rsa"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"reflect"
)

type BCClient struct {
	Root    string
	Peers   []string
	Cache   bcgo.Cache
	Network bcgo.Network
	Node    *bcgo.Node
}

func NewBCClient(peers ...string) *BCClient {
	/* TODO try loading from file system
		rootDir, err := c.GetRoot()
		if err == nil {
			bcgo.GetPeers(rootDir)
	}
	*/
	if len(peers) == 0 {
		peers = append(peers,
			bcgo.GetBCHost(), // Add BC host as peer
		)
	}
	return &BCClient{
		Peers: peers,
	}
}

func (c *BCClient) GetPeers() ([]string, error) {
	return c.Peers, nil
}

func (c *BCClient) GetCache() (bcgo.Cache, error) {
	if c.Cache == nil {
		rootDir, err := c.GetRoot()
		if err != nil {
			return nil, err
		}
		cacheDir, err := bcgo.GetCacheDirectory(rootDir)
		if err != nil {
			return nil, fmt.Errorf("Could not get cache directory: %s", err.Error())
		}
		cache, err := bcgo.NewFileCache(cacheDir)
		if err != nil {
			return nil, fmt.Errorf("Could not create file cache: %s", err.Error())
		}
		c.Cache = cache
	}
	return c.Cache, nil
}

func (c *BCClient) GetNetwork() (bcgo.Network, error) {
	if c.Network == nil || reflect.ValueOf(c.Network).IsNil() {
		peers, err := c.GetPeers()
		if err != nil {
			return nil, err
		}
		c.Network = bcgo.NewTCPNetwork(peers...)
	}
	return c.Network, nil
}

func (c *BCClient) GetNode() (*bcgo.Node, error) {
	if c.Node == nil {
		rootDir, err := c.GetRoot()
		if err != nil {
			return nil, err
		}
		cache, err := c.GetCache()
		if err != nil {
			return nil, err
		}
		network, err := c.GetNetwork()
		if err != nil {
			return nil, err
		}
		node, err := bcgo.NewNode(rootDir, cache, network)
		if err != nil {
			return nil, err
		}
		c.Node = node
	}
	return c.Node, nil
}

func (c *BCClient) SetRoot(root string) {
	c.Root = root
}

func (c *BCClient) SetPeers(peers ...string) {
	c.Peers = peers
	if c.Network == nil || reflect.ValueOf(c.Network).IsNil() {
		return
	}
	if n, ok := c.Network.(*bcgo.TCPNetwork); ok {
		n.SetPeers(peers...)
	}
}

func (c *BCClient) SetCache(cache bcgo.Cache) {
	c.Cache = cache
}

func (c *BCClient) SetNetwork(network bcgo.Network) {
	c.Network = network
}

func (c *BCClient) SetNode(node *bcgo.Node) {
	c.Node = node
}

func (c *BCClient) Init(listener bcgo.MiningListener) (*bcgo.Node, error) {
	// Create Node
	node, err := c.GetNode()
	if err != nil {
		return nil, err
	}

	// Register Alias
	if err := aliasgo.Register(node, listener); err != nil {
		return nil, err
	}

	return node, nil
}

func (c *BCClient) Alias(alias string) (string, error) {
	cache, err := c.GetCache()
	if err != nil {
		return "", err
	}
	network, err := c.GetNetwork()
	if err != nil {
		return "", err
	}
	// Open Alias Channel
	aliases := aliasgo.OpenAliasChannel()
	if err := aliases.Refresh(cache, network); err != nil {
		log.Println(err)
	}
	// Get Public Key for Alias
	publicKey, err := aliasgo.GetPublicKey(aliases, cache, network, alias)
	if err != nil {
		return "", err
	}
	publicKeyBytes, err := cryptogo.RSAPublicKeyToPKIXBytes(publicKey)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(publicKeyBytes), nil
}

func (c *BCClient) Head(channel string) ([]byte, error) {
	cache, err := c.GetCache()
	if err != nil {
		return nil, err
	}
	network, err := c.GetNetwork()
	if err != nil {
		return nil, err
	}
	ch := bcgo.NewChannel(channel)
	if err := ch.LoadHead(cache, network); err != nil {
		return nil, err
	}
	return ch.Head, nil
}

func (c *BCClient) Chain(channel string, callback func([]byte, *bcgo.Block) error) error {
	cache, err := c.GetCache()
	if err != nil {
		return err
	}
	network, err := c.GetNetwork()
	if err != nil {
		return err
	}
	ch := bcgo.NewChannel(channel)
	if err := ch.LoadHead(cache, network); err != nil {
		return err
	}
	return bcgo.Iterate(channel, ch.Head, nil, cache, network, callback)
}

func (c *BCClient) Block(channel string, hash []byte) (*bcgo.Block, error) {
	cache, err := c.GetCache()
	if err != nil {
		return nil, err
	}
	network, err := c.GetNetwork()
	if err != nil {
		return nil, err
	}
	block, err := bcgo.GetBlock(channel, cache, network, hash)
	if err != nil {
		return nil, err
	}
	return block, nil
}

func (c *BCClient) Record(channel string, hash []byte) (*bcgo.Record, error) {
	cache, err := c.GetCache()
	if err != nil {
		return nil, err
	}
	network, err := c.GetNetwork()
	if err != nil {
		return nil, err
	}
	block, err := bcgo.GetBlockContainingRecord(channel, cache, network, hash)
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

func (c *BCClient) Read(channel string, blockHash, recordHash []byte, output io.Writer) error {
	cache, err := c.GetCache()
	if err != nil {
		return err
	}
	network, err := c.GetNetwork()
	if err != nil {
		return err
	}
	node, err := c.GetNode()
	if err != nil {
		return err
	}

	ch := bcgo.NewChannel(channel)

	if err := ch.LoadHead(cache, network); err != nil {
		log.Println(err)
	}

	if blockHash == nil {
		blockHash = ch.Head
	}

	return bcgo.Read(channel, blockHash, nil, cache, network, node.Alias, node.Key, recordHash, func(entry *bcgo.BlockEntry, key, payload []byte) error {
		bcgo.PrintBlockEntry(output, "", entry)
		return nil
	})
}

func (c *BCClient) ReadKey(channel string, blockHash, recordHash []byte, output io.Writer) error {
	cache, err := c.GetCache()
	if err != nil {
		return err
	}
	network, err := c.GetNetwork()
	if err != nil {
		return err
	}
	node, err := c.GetNode()
	if err != nil {
		return err
	}

	ch := bcgo.NewChannel(channel)

	if err := ch.LoadHead(cache, network); err != nil {
		log.Println(err)
	}

	if blockHash == nil {
		blockHash = ch.Head
	}

	return bcgo.ReadKey(channel, blockHash, nil, cache, network, node.Alias, node.Key, recordHash, func(key []byte) error {
		output.Write(key)
		return nil
	})
}

func (c *BCClient) ReadPayload(channel string, blockHash, recordHash []byte, output io.Writer) error {
	cache, err := c.GetCache()
	if err != nil {
		return err
	}
	network, err := c.GetNetwork()
	if err != nil {
		return err
	}
	node, err := c.GetNode()
	if err != nil {
		return err
	}

	ch := bcgo.NewChannel(channel)

	if err := ch.LoadHead(cache, network); err != nil {
		log.Println(err)
	}

	if blockHash == nil {
		blockHash = ch.Head
	}

	return bcgo.Read(channel, blockHash, nil, cache, network, node.Alias, node.Key, recordHash, func(entry *bcgo.BlockEntry, key, payload []byte) error {
		output.Write(payload)
		return nil
	})
}

func (c *BCClient) Write(channel string, accesses []string, input io.Reader) (int, error) {
	cache, err := c.GetCache()
	if err != nil {
		return 0, err
	}
	network, err := c.GetNetwork()
	if err != nil {
		return 0, err
	}
	var acl map[string]*rsa.PublicKey

	if len(accesses) > 0 {
		// Open Alias Channel
		aliases := aliasgo.OpenAliasChannel()
		if err := aliases.Refresh(cache, network); err != nil {
			log.Println(err)
		}
		acl = aliasgo.GetPublicKeys(aliases, cache, network, accesses)
	}

	node, err := c.GetNode()
	if err != nil {
		return 0, err
	}

	size, err := bcgo.CreateRecords(node.Alias, node.Key, acl, nil, input, func(key []byte, record *bcgo.Record) error {
		_, err := bcgo.WriteRecord(channel, cache, record)
		return err
	})
	if err != nil {
		return 0, err
	}

	return size, nil
}

func (c *BCClient) Mine(channel string, threshold uint64, listener bcgo.MiningListener) ([]byte, error) {
	cache, err := c.GetCache()
	if err != nil {
		return nil, err
	}
	network, err := c.GetNetwork()
	if err != nil {
		return nil, err
	}
	node, err := c.GetNode()
	if err != nil {
		return nil, err
	}

	ch := bcgo.NewChannel(channel)

	if err := ch.LoadHead(cache, network); err != nil {
		log.Println(err)
	}

	hash, _, err := node.Mine(ch, threshold, listener)
	if err != nil {
		return nil, err
	}
	return hash, nil
}

func (c *BCClient) Pull(channel string) error {
	cache, err := c.GetCache()
	if err != nil {
		return err
	}
	network, err := c.GetNetwork()
	if err != nil {
		return err
	}
	ch := bcgo.NewChannel(channel)
	return ch.Pull(cache, network)
}

func (c *BCClient) Push(channel string) error {
	cache, err := c.GetCache()
	if err != nil {
		return err
	}
	network, err := c.GetNetwork()
	if err != nil {
		return err
	}
	ch := bcgo.NewChannel(channel)
	if err := ch.LoadHead(cache, nil); err != nil {
		return err
	}
	return ch.Push(cache, network)
}

func (c *BCClient) Purge() error {
	rootDir, err := c.GetRoot()
	if err != nil {
		return err
	}
	// Get cache directory
	dir, err := bcgo.GetCacheDirectory(rootDir)
	if err != nil {
		return err
	}
	return os.RemoveAll(dir)
}

func (c *BCClient) ImportKeys(peer, alias, accessCode string) error {
	rootDir, err := c.GetRoot()
	if err != nil {
		return err
	}
	// Get KeyStore
	keystore, err := bcgo.GetKeyDirectory(rootDir)
	if err != nil {
		return err
	}
	return cryptogo.ImportKeys(peer, keystore, alias, accessCode)
}

func (c *BCClient) ExportKeys(peer, alias string, password []byte) (string, error) {
	rootDir, err := c.GetRoot()
	if err != nil {
		return "", err
	}
	// Get KeyStore
	keystore, err := bcgo.GetKeyDirectory(rootDir)
	if err != nil {
		return "", err
	}
	return cryptogo.ExportKeys(peer, keystore, alias, password)
}

func PrintNode(output io.Writer, node *bcgo.Node) error {
	fmt.Fprintln(output, node.Alias)
	publicKeyBytes, err := cryptogo.RSAPublicKeyToPKIXBytes(&node.Key.PublicKey)
	if err != nil {
		return err
	}
	fmt.Fprintln(output, base64.RawURLEncoding.EncodeToString(publicKeyBytes))
	return nil
}

func rootDir() (string, error) {
	rootDir, err := bcgo.GetRootDirectory()
	if err != nil {
		return "", fmt.Errorf("Could not get root directory: %s\n", err.Error())
	}
	if err := bcgo.ReadConfig(rootDir); err != nil {
		log.Println("Error reading config:", err)
	}
	return rootDir, nil
}
