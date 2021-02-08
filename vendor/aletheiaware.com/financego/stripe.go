package financego

import (
	"github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/charge"
	"github.com/stripe/stripe-go/customer"
	"github.com/stripe/stripe-go/sub"
	"github.com/stripe/stripe-go/usagerecord"
	"github.com/stripe/stripe-go/webhook"
	"log"
	"os"
)

var _ Processor = (*Stripe)(nil)

type Stripe struct {
	//
}

func (p *Stripe) NewCharge(merchantAlias, customerAlias, paymentId, productId, planId, country, currency string, amount int64, description string) (*Charge, error) {
	// Stripe's minimum charge amount is 50 cents
	if amount < 50 {
		amount = 50
	}

	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")

	chargeParams := &stripe.ChargeParams{
		Amount:      stripe.Int64(amount),
		Currency:    stripe.String(string(stripe.CurrencyUSD)),
		Description: stripe.String(description),
	}
	chargeParams.SetSource(paymentId)
	ch, err := charge.New(chargeParams)
	if err != nil {
		return nil, err
	}

	charge := &Charge{
		MerchantAlias: merchantAlias,
		CustomerAlias: customerAlias,
		Processor:     PaymentProcessor_STRIPE,
		PaymentId:     paymentId,
		ChargeId:      ch.ID,
		Amount:        amount,
		ProductId:     productId,
		PlanId:        planId,
		Country:       country,
		Currency:      currency,
		Description:   description,
	}
	log.Println("Charge", ch, charge)
	return charge, nil
}

func (p *Stripe) NewRegistration(merchantAlias, customerAlias, email, paymentId, description string) (*Registration, error) {
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")
	// Create new Stripe customer
	customerParams := &stripe.CustomerParams{
		Name:        stripe.String(customerAlias),
		Description: stripe.String(description),
		Email:       stripe.String(email),
	}
	if err := customerParams.SetSource(paymentId); err != nil {
		return nil, err
	}
	c, err := customer.New(customerParams)
	if err != nil {
		return nil, err
	}

	registration := &Registration{
		MerchantAlias: merchantAlias,
		CustomerAlias: customerAlias,
		Processor:     PaymentProcessor_STRIPE,
		CustomerId:    c.ID,
		PaymentId:     paymentId,
	}
	log.Println("Registration", c, registration)
	return registration, nil
}

func (p *Stripe) NewCustomerCharge(registration *Registration, productId, planId, country, currency string, amount int64, description string) (*Charge, error) {
	// Stripe's minimum charge amount is 50 cents
	if amount < 50 {
		amount = 50
	}

	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")

	chargeParams := &stripe.ChargeParams{
		Amount:      stripe.Int64(amount),
		Currency:    stripe.String(string(stripe.CurrencyUSD)),
		Customer:    stripe.String(registration.CustomerId),
		Description: stripe.String(description),
	}
	ch, err := charge.New(chargeParams)
	if err != nil {
		return nil, err
	}

	charge := &Charge{
		MerchantAlias: registration.MerchantAlias,
		CustomerAlias: registration.CustomerAlias,
		Processor:     PaymentProcessor_STRIPE,
		CustomerId:    registration.CustomerId,
		ChargeId:      ch.ID,
		Amount:        amount,
		ProductId:     productId,
		PlanId:        planId,
		Country:       country,
		Currency:      currency,
		Description:   description,
	}
	log.Println("Charge", ch, charge)
	return charge, nil
}

func (p *Stripe) NewSubscription(merchantAlias, customerAlias, customerId, paymentId, productId, planId string) (*Subscription, error) {
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")
	// Create new Stripe subscription
	subscriptionParams := &stripe.SubscriptionParams{
		Customer: stripe.String(customerId),
		Items: []*stripe.SubscriptionItemsParams{
			{
				Plan: stripe.String(planId),
			},
		},
	}
	s, err := sub.New(subscriptionParams)
	if err != nil {
		return nil, err
	}

	// Create subscription
	subscription := &Subscription{
		MerchantAlias:      merchantAlias,
		CustomerAlias:      customerAlias,
		Processor:          PaymentProcessor_STRIPE,
		CustomerId:         customerId,
		ProductId:          productId,
		PlanId:             planId,
		SubscriptionId:     s.ID,
		SubscriptionItemId: s.Items.Data[0].ID,
	}
	if paymentId != "" {
		subscription.PaymentId = paymentId
	}
	log.Println("Subscription", s, subscription)
	return subscription, nil
}

func (p *Stripe) NewUsageRecord(merchantAlias, customerAlias, subscriptionId, subscriptionItemId, productId, planId string, timestamp int64, size int64) (*UsageRecord, error) {
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")

	params := &stripe.UsageRecordParams{
		SubscriptionItem: stripe.String(subscriptionItemId),
		Timestamp:        stripe.Int64(timestamp),
		Quantity:         stripe.Int64(size),
	}
	ur, err := usagerecord.New(params)
	if err != nil {
		return nil, err
	}

	// Create usage record
	usage := &UsageRecord{
		MerchantAlias:      merchantAlias,
		CustomerAlias:      customerAlias,
		Processor:          PaymentProcessor_STRIPE,
		SubscriptionId:     subscriptionId,
		SubscriptionItemId: subscriptionItemId,
		UsageRecordId:      ur.ID,
		Quantity:           size,
		ProductId:          productId,
		PlanId:             planId,
	}
	log.Println("UsageRecord", ur, usage)
	return usage, nil
}

func ConstructEvent(data []byte, signature string) (stripe.Event, error) {
	secretKey := os.Getenv("STRIPE_WEB_HOOK_SECRET_KEY")
	return webhook.ConstructEvent(data, signature, secretKey)
}
