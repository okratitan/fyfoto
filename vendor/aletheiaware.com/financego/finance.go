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

func ChargeAsync(charges bcgo.Channel, cache bcgo.Cache, network bcgo.Network, reader bcgo.Account, merchant, customer string, callback ChargeCallback) error {
	if err := charges.Load(cache, nil); err != nil {
		log.Println(err)
	}
	name := charges.Name()
	head := charges.Head()
	cb := func(entry *bcgo.BlockEntry, key, data []byte) error {
		// Unmarshal as Charge
		charge := &Charge{}
		err := proto.Unmarshal(data, charge)
		if err != nil {
			return err
		} else if (merchant == "" || charge.MerchantAlias == merchant) && (customer == "" || charge.CustomerAlias == customer) {
			return callback(entry, charge)
		}
		return nil
	}
	return bcgo.Read(name, head, nil, cache, network, reader, nil, cb)
}

func ChargeSync(charges bcgo.Channel, cache bcgo.Cache, network bcgo.Network, reader bcgo.Account, merchant, customer string) (*Charge, error) {
	var charge *Charge
	if err := ChargeAsync(charges, cache, network, reader, merchant, customer, func(e *bcgo.BlockEntry, c *Charge) error {
		charge = c
		return bcgo.ErrStopIteration{}
	}); err != nil {
		switch err.(type) {
		case bcgo.ErrStopIteration:
			// Do nothing
			break
		default:
			return nil, err
		}
	}
	return charge, nil
}

func RegistrationAsync(registrations bcgo.Channel, cache bcgo.Cache, network bcgo.Network, reader bcgo.Account, merchant, customer string, callback RegistrationCallback) error {
	if err := registrations.Load(cache, nil); err != nil {
		log.Println(err)
	}
	name := registrations.Name()
	head := registrations.Head()
	cb := func(entry *bcgo.BlockEntry, key, data []byte) error {
		// Unmarshal as Registration
		registration := &Registration{}
		err := proto.Unmarshal(data, registration)
		if err != nil {
			return err
		} else if (merchant == "" || registration.MerchantAlias == merchant) && (customer == "" || registration.CustomerAlias == customer) {
			return callback(entry, registration)
		}
		return nil
	}
	return bcgo.Read(name, head, nil, cache, network, reader, nil, cb)
}

func RegistrationSync(registrations bcgo.Channel, cache bcgo.Cache, network bcgo.Network, reader bcgo.Account, merchant, customer string) (*Registration, error) {
	var registration *Registration
	if err := RegistrationAsync(registrations, cache, network, reader, merchant, customer, func(e *bcgo.BlockEntry, r *Registration) error {
		registration = r
		return bcgo.ErrStopIteration{}
	}); err != nil {
		switch err.(type) {
		case bcgo.ErrStopIteration:
			// Do nothing
			break
		default:
			return nil, err
		}
	}
	return registration, nil
}

func SubscriptionAsync(subscriptions bcgo.Channel, cache bcgo.Cache, network bcgo.Network, reader bcgo.Account, merchant, customer string, productId, planId string, callback SubscriptionCallback) error {
	if err := subscriptions.Load(cache, nil); err != nil {
		log.Println(err)
	}
	name := subscriptions.Name()
	head := subscriptions.Head()
	cb := func(entry *bcgo.BlockEntry, key, data []byte) error {
		// Unmarshal as Subscription
		subscription := &Subscription{}
		err := proto.Unmarshal(data, subscription)
		if err != nil {
			return err
		} else if (merchant == "" || subscription.MerchantAlias == merchant) && (customer == "" || subscription.CustomerAlias == customer) && (productId == "" || subscription.ProductId == productId) && (planId == "" || subscription.PlanId == planId) {
			return callback(entry, subscription)
		}
		return nil
	}
	return bcgo.Read(name, head, nil, cache, network, reader, nil, cb)
}

func SubscriptionSync(subscriptions bcgo.Channel, cache bcgo.Cache, network bcgo.Network, reader bcgo.Account, merchant, customer string, productId, planId string) (*Subscription, error) {
	var subscription *Subscription
	if err := SubscriptionAsync(subscriptions, cache, network, reader, merchant, customer, productId, planId, func(e *bcgo.BlockEntry, s *Subscription) error {
		subscription = s
		return bcgo.ErrStopIteration{}
	}); err != nil {
		switch err.(type) {
		case bcgo.ErrStopIteration:
			// Do nothing
			break
		default:
			return nil, err
		}
	}
	return subscription, nil
}

func UsageRecordAsync(usages bcgo.Channel, cache bcgo.Cache, network bcgo.Network, reader bcgo.Account, merchant, customer string, callback UsageRecordCallback) error {
	if err := usages.Load(cache, nil); err != nil {
		log.Println(err)
	}
	name := usages.Name()
	head := usages.Head()
	cb := func(entry *bcgo.BlockEntry, key, data []byte) error {
		// Unmarshal as UsageRecord
		usage := &UsageRecord{}
		err := proto.Unmarshal(data, usage)
		if err != nil {
			return err
		} else if (merchant == "" || usage.MerchantAlias == merchant) && (customer == "" || usage.CustomerAlias == customer) {
			return callback(entry, usage)
		}
		return nil
	}
	return bcgo.Read(name, head, nil, cache, network, reader, nil, cb)
}

func UsageRecordSync(usages bcgo.Channel, cache bcgo.Cache, network bcgo.Network, reader bcgo.Account, merchant, customer string) (*UsageRecord, error) {
	var usage *UsageRecord
	if err := UsageRecordAsync(usages, cache, network, reader, merchant, customer, func(e *bcgo.BlockEntry, u *UsageRecord) error {
		usage = u
		return bcgo.ErrStopIteration{}
	}); err != nil {
		switch err.(type) {
		case bcgo.ErrStopIteration:
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
