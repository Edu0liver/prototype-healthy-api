// Package stripe is a thin client over the Stripe API covering the slice the
// platform needs for subscription billing: creating Checkout Sessions and
// verifying webhook signatures. It has no dependency on internal/ — callers
// pass a stripe.Config (mirrors the openai/evolution clients).
package stripe

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const apiBase = "https://api.stripe.com"

// Config configures the Stripe client.
type Config struct {
	SecretKey       string // sk_live_... / sk_test_...
	WebhookSecret   string // whsec_... used to verify webhook signatures
	SuccessURL      string // checkout success redirect
	CancelURL       string // checkout cancel redirect
	PortalReturnURL string // billing portal return redirect
	Timeout         time.Duration
}

// CheckoutParams describes a subscription Checkout Session to create.
type CheckoutParams struct {
	PriceID       string            // Stripe price to subscribe to
	CompanyID     string            // our tenant id (client_reference_id + metadata)
	CustomerEmail string            // prefilled email (optional)
	CustomerID    string            // existing Stripe customer (optional)
	Metadata      map[string]string // extra metadata (e.g. plan_id)
}

// CheckoutSession is the relevant slice of a created Checkout Session.
type CheckoutSession struct {
	ID         string `json:"id"`
	URL        string `json:"url"`
	CustomerID string `json:"customer"`
}

// PortalSession is the relevant slice of a created Billing Portal Session.
type PortalSession struct {
	ID  string `json:"id"`
	URL string `json:"url"`
}

// Client is the Stripe API contract (interface for mocking).
type Client interface {
	CreateCheckoutSession(ctx context.Context, p CheckoutParams) (*CheckoutSession, error)
	CreatePortalSession(ctx context.Context, customerID string) (*PortalSession, error)
	VerifyWebhook(payload []byte, sigHeader string) (*Event, error)
}

// HTTPClient is the live implementation.
type HTTPClient struct {
	cfg  Config
	http *http.Client
}

// New builds the live Stripe client.
func New(cfg Config) *HTTPClient {
	if cfg.Timeout == 0 {
		cfg.Timeout = 20 * time.Second
	}
	return &HTTPClient{cfg: cfg, http: &http.Client{Timeout: cfg.Timeout}}
}

// CreateCheckoutSession creates a subscription-mode Checkout Session and returns
// its hosted URL for the tenant to complete payment.
func (c *HTTPClient) CreateCheckoutSession(ctx context.Context, p CheckoutParams) (*CheckoutSession, error) {
	form := url.Values{}
	form.Set("mode", "subscription")
	form.Set("line_items[0][price]", p.PriceID)
	form.Set("line_items[0][quantity]", "1")
	form.Set("success_url", c.cfg.SuccessURL)
	form.Set("cancel_url", c.cfg.CancelURL)
	form.Set("client_reference_id", p.CompanyID)
	form.Set("subscription_data[metadata][company_id]", p.CompanyID)
	if p.CustomerID != "" {
		form.Set("customer", p.CustomerID)
	} else if p.CustomerEmail != "" {
		form.Set("customer_email", p.CustomerEmail)
	}
	for k, v := range p.Metadata {
		form.Set("metadata["+k+"]", v)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiBase+"/v1/checkout/sessions", strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.cfg.SecretKey)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("stripe: checkout session %s: %s", resp.Status, string(body))
	}
	var out CheckoutSession
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// CreatePortalSession creates a Billing Portal session for an existing Stripe
// customer and returns the hosted portal URL.
func (c *HTTPClient) CreatePortalSession(ctx context.Context, customerID string) (*PortalSession, error) {
	if c.cfg.PortalReturnURL == "" {
		return nil, fmt.Errorf("stripe: STRIPE_PORTAL_RETURN_URL not configured")
	}
	form := url.Values{}
	form.Set("customer", customerID)
	form.Set("return_url", c.cfg.PortalReturnURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiBase+"/v1/billing_portal/sessions", strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.cfg.SecretKey)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("stripe: portal session %s: %s", resp.Status, string(body))
	}
	var out PortalSession
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// signatureTolerance bounds how old a webhook timestamp may be (replay defense).
const signatureTolerance = 5 * time.Minute

// VerifyWebhook validates the Stripe-Signature header against the configured
// webhook secret and returns the parsed event. It implements Stripe's scheme:
// the signed payload is "<t>.<rawBody>", HMAC-SHA256 with the secret, compared
// in constant time to the v1 signature, within the timestamp tolerance.
func (c *HTTPClient) VerifyWebhook(payload []byte, sigHeader string) (*Event, error) {
	if err := VerifySignature(payload, sigHeader, c.cfg.WebhookSecret, time.Now()); err != nil {
		return nil, err
	}
	var e Event
	if err := json.Unmarshal(payload, &e); err != nil {
		return nil, fmt.Errorf("stripe: parse event: %w", err)
	}
	return &e, nil
}

// VerifySignature is the pure signature check (exported for testing). now is
// injected so tests are deterministic.
func VerifySignature(payload []byte, sigHeader, secret string, now time.Time) error {
	if secret == "" {
		return fmt.Errorf("stripe: webhook secret not configured")
	}
	ts, sigs := parseSigHeader(sigHeader)
	if ts == "" || len(sigs) == 0 {
		return fmt.Errorf("stripe: malformed signature header")
	}
	tsInt, err := strconv.ParseInt(ts, 10, 64)
	if err != nil {
		return fmt.Errorf("stripe: bad signature timestamp")
	}
	if d := now.Sub(time.Unix(tsInt, 0)); d > signatureTolerance || d < -signatureTolerance {
		return fmt.Errorf("stripe: signature timestamp outside tolerance")
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(ts))
	mac.Write([]byte("."))
	mac.Write(payload)
	expected := mac.Sum(nil)
	for _, s := range sigs {
		got, err := hex.DecodeString(s)
		if err != nil {
			continue
		}
		if hmac.Equal(got, expected) {
			return nil
		}
	}
	return fmt.Errorf("stripe: no matching signature")
}

// parseSigHeader splits "t=123,v1=abc,v1=def" into its timestamp and v1 sigs.
func parseSigHeader(h string) (ts string, v1s []string) {
	for _, part := range strings.Split(h, ",") {
		kv := strings.SplitN(strings.TrimSpace(part), "=", 2)
		if len(kv) != 2 {
			continue
		}
		switch kv[0] {
		case "t":
			ts = kv[1]
		case "v1":
			v1s = append(v1s, kv[1])
		}
	}
	return ts, v1s
}

// Event is the relevant slice of a Stripe webhook event.
type Event struct {
	ID   string `json:"id"`
	Type string `json:"type"`
	Data struct {
		Object EventObject `json:"object"`
	} `json:"data"`
}

// EventObject covers the fields used across the subscription lifecycle events
// (checkout.session.completed, customer.subscription.updated/deleted,
// invoice.payment_failed).
type EventObject struct {
	ID                string            `json:"id"`
	Customer          string            `json:"customer"`
	Subscription      string            `json:"subscription"` // on checkout.session / invoice
	ClientReferenceID string            `json:"client_reference_id"`
	Status            string            `json:"status"`
	CancelAtPeriodEnd bool              `json:"cancel_at_period_end"`
	CurrentPeriodEnd  int64             `json:"current_period_end"`
	Metadata          map[string]string `json:"metadata"`
	Items             struct {
		Data []struct {
			Price struct {
				ID string `json:"id"`
			} `json:"price"`
		} `json:"data"`
	} `json:"items"`
}

// FirstPriceID returns the subscription's first line-item price id, if present.
func (o EventObject) FirstPriceID() string {
	if len(o.Items.Data) > 0 {
		return o.Items.Data[0].Price.ID
	}
	return ""
}
