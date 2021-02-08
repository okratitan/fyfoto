/*
 * Copyright 2020 Aletheia Ware LLC
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

package financego

import (
	"aletheiaware.com/aliasgo"
	"aletheiaware.com/bcgo"
	"crypto/rsa"
	"github.com/golang/protobuf/proto"
	"log"
)

func Register(merchant *bcgo.Node, processor Processor, aliases, registrations *bcgo.Channel, threshold uint64, listener bcgo.MiningListener) func(string, string, string) (string, *bcgo.Reference, error) {
	return func(customerAlias, customerEmail, customerToken string) (string, *bcgo.Reference, error) {
		cache := merchant.Cache
		network := merchant.Network

		if err := aliases.Refresh(cache, network); err != nil {
			log.Println(err)
		}

		// Get rsa.PublicKey for Alias
		publicKey, err := aliasgo.GetPublicKey(aliases, cache, network, customerAlias)
		if err != nil {
			return "", nil, err
		}

		// Create list of access (user + server)
		acl := map[string]*rsa.PublicKey{
			customerAlias:  publicKey,
			merchant.Alias: &merchant.Key.PublicKey,
		}
		log.Println("Access", acl)

		registration, err := processor.NewRegistration(merchant.Alias, customerAlias, customerEmail, customerToken, customerAlias+" "+merchant.Alias)
		if err != nil {
			return "", nil, err
		}

		registrationData, err := proto.Marshal(registration)
		if err != nil {
			return "", nil, err
		}

		if err := registrations.Refresh(cache, network); err != nil {
			log.Println(err)
		}

		_, err = merchant.Write(bcgo.Timestamp(), registrations, acl, nil, registrationData)
		if err != nil {
			return "", nil, err
		}

		registrationHash, registrationBlock, err := merchant.Mine(registrations, threshold, listener)
		if err != nil {
			return "", nil, err
		}

		if network != nil {
			if err := registrations.Push(cache, network); err != nil {
				log.Println(err)
			}
		}

		registrationReference := &bcgo.Reference{
			Timestamp:   registrationBlock.Timestamp,
			ChannelName: registrationBlock.ChannelName,
			BlockHash:   registrationHash,
		}

		log.Println("RegistrationReference", registrationReference)

		return registration.CustomerId, registrationReference, nil
	}
}

func Subscribe(merchant *bcgo.Node, processor Processor, aliases, subscriptions *bcgo.Channel, threshold uint64, listener bcgo.MiningListener, productId, planId string) func(string, string) (string, *bcgo.Reference, error) {
	return func(customerAlias, customerID string) (string, *bcgo.Reference, error) {
		cache := merchant.Cache
		network := merchant.Network

		if err := aliases.Refresh(cache, network); err != nil {
			log.Println(err)
		}

		// Get rsa.PublicKey for Alias
		publicKey, err := aliasgo.GetPublicKey(aliases, cache, network, customerAlias)
		if err != nil {
			return "", nil, err
		}

		// Create list of access (user + server)
		acl := map[string]*rsa.PublicKey{
			customerAlias:  publicKey,
			merchant.Alias: &merchant.Key.PublicKey,
		}
		log.Println("Access", acl)

		subscription, err := processor.NewSubscription(merchant.Alias, customerAlias, customerID, "", productId, planId)
		if err != nil {
			return "", nil, err
		}

		subscriptionData, err := proto.Marshal(subscription)
		if err != nil {
			return "", nil, err
		}

		if err := subscriptions.Refresh(cache, network); err != nil {
			log.Println(err)
		}

		_, err = merchant.Write(bcgo.Timestamp(), subscriptions, acl, nil, subscriptionData)
		if err != nil {
			return "", nil, err
		}

		subscriptionHash, subscriptionBlock, err := merchant.Mine(subscriptions, threshold, listener)
		if err != nil {
			return "", nil, err
		}

		if network != nil {
			if err := subscriptions.Push(cache, network); err != nil {
				log.Println(err)
			}
		}

		subscriptionReference := &bcgo.Reference{
			Timestamp:   subscriptionBlock.Timestamp,
			ChannelName: subscriptionBlock.ChannelName,
			BlockHash:   subscriptionHash,
		}

		log.Println("SubscriptionReference", subscriptionReference)

		return subscription.SubscriptionItemId, subscriptionReference, nil
	}
}
