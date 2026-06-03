package stripe

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"testing"
	"time"
)

func sign(payload []byte, secret string, t time.Time) string {
	ts := fmt.Sprintf("%d", t.Unix())
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(ts + "." + string(payload)))
	return fmt.Sprintf("t=%s,v1=%s", ts, hex.EncodeToString(mac.Sum(nil)))
}

func TestVerifySignature_Valid(t *testing.T) {
	payload := []byte(`{"id":"evt_1","type":"checkout.session.completed"}`)
	now := time.Now()
	header := sign(payload, "whsec_test", now)
	if err := VerifySignature(payload, header, "whsec_test", now); err != nil {
		t.Fatalf("expected valid signature, got %v", err)
	}
}

func TestVerifySignature_TamperedPayload(t *testing.T) {
	now := time.Now()
	header := sign([]byte(`{"id":"evt_1"}`), "whsec_test", now)
	if err := VerifySignature([]byte(`{"id":"evt_HACKED"}`), header, "whsec_test", now); err == nil {
		t.Fatal("expected error for tampered payload")
	}
}

func TestVerifySignature_WrongSecret(t *testing.T) {
	payload := []byte(`{"id":"evt_1"}`)
	now := time.Now()
	header := sign(payload, "whsec_test", now)
	if err := VerifySignature(payload, header, "whsec_OTHER", now); err == nil {
		t.Fatal("expected error for wrong secret")
	}
}

func TestVerifySignature_OutsideTolerance(t *testing.T) {
	payload := []byte(`{"id":"evt_1"}`)
	signed := time.Now().Add(-10 * time.Minute)
	header := sign(payload, "whsec_test", signed)
	if err := VerifySignature(payload, header, "whsec_test", time.Now()); err == nil {
		t.Fatal("expected error for stale timestamp")
	}
}

func TestVerifySignature_Malformed(t *testing.T) {
	if err := VerifySignature([]byte(`{}`), "garbage", "whsec_test", time.Now()); err == nil {
		t.Fatal("expected error for malformed header")
	}
	if err := VerifySignature([]byte(`{}`), "t=1,v1=ab", "", time.Now()); err == nil {
		t.Fatal("expected error for empty secret")
	}
}

func TestFirstPriceID(t *testing.T) {
	var o EventObject
	if o.FirstPriceID() != "" {
		t.Fatal("empty items should yield empty price id")
	}
	o.Items.Data = append(o.Items.Data, struct {
		Price struct {
			ID string `json:"id"`
		} `json:"price"`
	}{})
	o.Items.Data[0].Price.ID = "price_123"
	if o.FirstPriceID() != "price_123" {
		t.Fatalf("got %q", o.FirstPriceID())
	}
}
