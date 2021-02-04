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

package financego

import (
	"aletheiaware.com/bcgo"
	"crypto/rsa"
	"github.com/golang/protobuf/proto"
	"log"
)

type Processor interface {
	NewCharge(merchantAlias, customerAlias, paymentId, productId, planId, country, currency string, amount int64, description string) (*Charge, error)
	NewRegistration(merchantAlias, customerAlias, email, paymentId, description string) (*Registration, error)
	NewCustomerCharge(registration *Registration, productId, planId, country, currency string, amount int64, description string) (*Charge, error)
	NewSubscription(merchantAlias, customerAlias, customerId, paymentId, productId, planId string) (*Subscription, error)
	NewUsageRecord(merchantAlias, customerAlias, subscriptionId, subscriptionItemId, productId, planId string, timestamp int64, size int64) (*UsageRecord, error)
}

type ChargeCallback func(*bcgo.BlockEntry, *Charge) error

type RegistrationCallback func(*bcgo.BlockEntry, *Registration) error

type SubscriptionCallback func(*bcgo.BlockEntry, *Subscription) error

type UsageRecordCallback func(*bcgo.BlockEntry, *UsageRecord) error

func GetChargeAsync(charges *bcgo.Channel, cache bcgo.Cache, network bcgo.Network, merchantAlias string, merchantKey *rsa.PrivateKey, customerAlias string, customerKey *rsa.PrivateKey, callback ChargeCallback) error {
	if err := charges.LoadCachedHead(cache); err != nil {
		log.Println(err)
	}
	cb := func(entry *bcgo.BlockEntry, key, data []byte) error {
		// Unmarshal as Charge
		charge := &Charge{}
		err := proto.Unmarshal(data, charge)
		if err != nil {
			return err
		} else if (merchantAlias == "" || charge.MerchantAlias == merchantAlias) && (customerAlias == "" || charge.CustomerAlias == customerAlias) {
			return callback(entry, charge)
		}
		return nil
	}
	// Read as merchant
	if merchantAlias != "" && merchantKey != nil {
		return bcgo.Read(charges.Name, charges.Head, nil, cache, network, merchantAlias, merchantKey, nil, cb)
	}
	// Read as customer
	return bcgo.Read(charges.Name, charges.Head, nil, cache, network, customerAlias, customerKey, nil, cb)
}

func GetChargeSync(charges *bcgo.Channel, cache bcgo.Cache, network bcgo.Network, merchantAlias string, merchantKey *rsa.PrivateKey, customerAlias string, customerKey *rsa.PrivateKey) (*Charge, error) {
	var charge *Charge
	if err := GetChargeAsync(charges, cache, network, merchantAlias, merchantKey, customerAlias, customerKey, func(e *bcgo.BlockEntry, c *Charge) error {
		charge = c
		return bcgo.StopIterationError{}
	}); err != nil {
		switch err.(type) {
		case bcgo.StopIterationError:
			// Do nothing
			break
		default:
			return nil, err
		}
	}
	return charge, nil
}

func GetRegistrationAsync(registrations *bcgo.Channel, cache bcgo.Cache, network bcgo.Network, merchantAlias string, merchantKey *rsa.PrivateKey, customerAlias string, customerKey *rsa.PrivateKey, callback RegistrationCallback) error {
	if err := registrations.LoadCachedHead(cache); err != nil {
		log.Println(err)
	}
	cb := func(entry *bcgo.BlockEntry, key, data []byte) error {
		// Unmarshal as Registration
		registration := &Registration{}
		err := proto.Unmarshal(data, registration)
		if err != nil {
			return err
		} else if (merchantAlias == "" || registration.MerchantAlias == merchantAlias) && (customerAlias == "" || registration.CustomerAlias == customerAlias) {
			return callback(entry, registration)
		}
		return nil
	}
	// Read as merchant
	if merchantAlias != "" && merchantKey != nil {
		return bcgo.Read(registrations.Name, registrations.Head, nil, cache, network, merchantAlias, merchantKey, nil, cb)
	}
	// Read as customer
	return bcgo.Read(registrations.Name, registrations.Head, nil, cache, network, customerAlias, customerKey, nil, cb)
}

func GetRegistrationSync(registrations *bcgo.Channel, cache bcgo.Cache, network bcgo.Network, merchantAlias string, merchantKey *rsa.PrivateKey, customerAlias string, customerKey *rsa.PrivateKey) (*Registration, error) {
	var registration *Registration
	if err := GetRegistrationAsync(registrations, cache, network, merchantAlias, merchantKey, customerAlias, customerKey, func(e *bcgo.BlockEntry, r *Registration) error {
		registration = r
		return bcgo.StopIterationError{}
	}); err != nil {
		switch err.(type) {
		case bcgo.StopIterationError:
			// Do nothing
			break
		default:
			return nil, err
		}
	}
	return registration, nil
}

func GetSubscriptionAsync(subscriptions *bcgo.Channel, cache bcgo.Cache, network bcgo.Network, merchantAlias string, merchantKey *rsa.PrivateKey, customerAlias string, customerKey *rsa.PrivateKey, productId, planId string, callback SubscriptionCallback) error {
	if err := subscriptions.LoadCachedHead(cache); err != nil {
		log.Println(err)
	}
	cb := func(entry *bcgo.BlockEntry, key, data []byte) error {
		// Unmarshal as Subscription
		subscription := &Subscription{}
		err := proto.Unmarshal(data, subscription)
		if err != nil {
			return err
		} else if (merchantAlias == "" || subscription.MerchantAlias == merchantAlias) && (customerAlias == "" || subscription.CustomerAlias == customerAlias) && (productId == "" || subscription.ProductId == productId) && (planId == "" || subscription.PlanId == planId) {
			return callback(entry, subscription)
		}
		return nil
	}
	// Read as merchant
	if merchantAlias != "" && merchantKey != nil {
		return bcgo.Read(subscriptions.Name, subscriptions.Head, nil, cache, network, merchantAlias, merchantKey, nil, cb)
	}
	// Read as customer
	return bcgo.Read(subscriptions.Name, subscriptions.Head, nil, cache, network, customerAlias, customerKey, nil, cb)
}

func GetSubscriptionSync(subscriptions *bcgo.Channel, cache bcgo.Cache, network bcgo.Network, merchantAlias string, merchantKey *rsa.PrivateKey, customerAlias string, customerKey *rsa.PrivateKey, productId, planId string) (*Subscription, error) {
	var subscription *Subscription
	if err := GetSubscriptionAsync(subscriptions, cache, network, merchantAlias, merchantKey, customerAlias, customerKey, productId, planId, func(e *bcgo.BlockEntry, s *Subscription) error {
		subscription = s
		return bcgo.StopIterationError{}
	}); err != nil {
		switch err.(type) {
		case bcgo.StopIterationError:
			// Do nothing
			break
		default:
			return nil, err
		}
	}
	return subscription, nil
}

func GetUsageRecordAsync(usages *bcgo.Channel, cache bcgo.Cache, network bcgo.Network, merchantAlias string, merchantKey *rsa.PrivateKey, customerAlias string, customerKey *rsa.PrivateKey, callback UsageRecordCallback) error {
	if err := usages.LoadCachedHead(cache); err != nil {
		log.Println(err)
	}
	cb := func(entry *bcgo.BlockEntry, key, data []byte) error {
		// Unmarshal as UsageRecord
		usage := &UsageRecord{}
		err := proto.Unmarshal(data, usage)
		if err != nil {
			return err
		} else if (merchantAlias == "" || usage.MerchantAlias == merchantAlias) && (customerAlias == "" || usage.CustomerAlias == customerAlias) {
			return callback(entry, usage)
		}
		return nil
	}
	// Read as merchant
	if merchantAlias != "" && merchantKey != nil {
		return bcgo.Read(usages.Name, usages.Head, nil, cache, network, merchantAlias, merchantKey, nil, cb)
	}
	// Read as customer
	return bcgo.Read(usages.Name, usages.Head, nil, cache, network, customerAlias, customerKey, nil, cb)
}

func GetUsageRecordSync(usages *bcgo.Channel, cache bcgo.Cache, network bcgo.Network, merchantAlias string, merchantKey *rsa.PrivateKey, customerAlias string, customerKey *rsa.PrivateKey) (*UsageRecord, error) {
	var usage *UsageRecord
	if err := GetUsageRecordAsync(usages, cache, network, merchantAlias, merchantKey, customerAlias, customerKey, func(e *bcgo.BlockEntry, u *UsageRecord) error {
		usage = u
		return bcgo.StopIterationError{}
	}); err != nil {
		switch err.(type) {
		case bcgo.StopIterationError:
			// Do nothing
			break
		default:
			return nil, err
		}
	}
	return usage, nil
}

func IntervalToString(interval Service_Interval) string {
	switch interval {
	case Service_DAILY:
		return "Day"
	case Service_WEEKLY:
		return "Week"
	case Service_MONTHLY:
		return "Month"
	case Service_QUARTERLY: // Every 3 months
		return "Quarter"
	case Service_YEARLY:
		return "Year"
	}
	return "?"
}
