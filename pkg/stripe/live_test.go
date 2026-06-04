package stripe_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Edu0liver/prototype-healthy-api/pkg/stripe"
)

// stripePostForm is a minimal helper to set up test fixtures (product/price)
// directly against the Stripe API. The client under test only creates checkout
// sessions, so the test provisions the prerequisites itself.
func stripePostForm(t *testing.T, key, path string, form url.Values) map[string]any {
	t.Helper()
	req, err := http.NewRequest(http.MethodPost, "https://api.stripe.com"+path, strings.NewReader(form.Encode()))
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+key)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := (&http.Client{Timeout: 20 * time.Second}).Do(req)
	if err != nil {
		t.Fatalf("stripe %s: %v", path, err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		t.Fatalf("stripe %s -> %s: %s", path, resp.Status, string(body))
	}
	var out map[string]any
	if err := json.Unmarshal(body, &out); err != nil {
		t.Fatalf("decode %s: %v", path, err)
	}
	return out
}

// TestLiveCheckoutSession exercises CreateCheckoutSession against the real Stripe
// test API. It is skipped unless STRIPE_SECRET_KEY (a sk_test_/rk_test_ key) is
// set, so CI without credentials stays green. Run with:
//
//	STRIPE_SECRET_KEY=sk_test_... go test ./pkg/stripe/ -run TestLiveCheckout -v
func TestLiveCheckoutSession(t *testing.T) {
	key := os.Getenv("STRIPE_SECRET_KEY")
	if key == "" {
		t.Skip("STRIPE_SECRET_KEY not set; skipping live Stripe test")
	}
	if !strings.HasPrefix(key, "sk_test_") && !strings.HasPrefix(key, "rk_test_") {
		t.Fatalf("refusing to run live test with a non-test key (got prefix %.7s)", key)
	}

	// 1. Create a throwaway product.
	prod := stripePostForm(t, key, "/v1/products", url.Values{
		"name": {"Lumia live-test " + time.Now().Format("150405")},
	})
	prodID, _ := prod["id"].(string)
	if prodID == "" {
		t.Fatal("no product id returned")
	}

	// 2. Create a recurring monthly price for it.
	price := stripePostForm(t, key, "/v1/prices", url.Values{
		"product":             {prodID},
		"unit_amount":         {"29900"},
		"currency":            {"brl"},
		"recurring[interval]": {"month"},
	})
	priceID, _ := price["id"].(string)
	if priceID == "" {
		t.Fatal("no price id returned")
	}
	t.Logf("provisioned price %s", priceID)

	// 3. Create a Checkout Session through the client under test.
	client := stripe.New(stripe.Config{
		SecretKey:  key,
		SuccessURL: "https://example.test/ok",
		CancelURL:  "https://example.test/cancel",
	})
	sess, err := client.CreateCheckoutSession(context.Background(), stripe.CheckoutParams{
		PriceID:       priceID,
		CompanyID:     "11111111-1111-1111-1111-111111111111",
		CustomerEmail: "live-test@example.test",
		Metadata:      map[string]string{"plan_code": "pro"},
	})
	if err != nil {
		t.Fatalf("CreateCheckoutSession: %v", err)
	}
	if !strings.HasPrefix(sess.ID, "cs_") {
		t.Errorf("unexpected session id %q", sess.ID)
	}
	if !strings.Contains(sess.URL, "checkout.stripe.com") {
		t.Errorf("unexpected checkout url %q", sess.URL)
	}
	t.Logf("checkout session %s -> %s", sess.ID, sess.URL)
}
