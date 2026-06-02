package crypto

import (
	"encoding/hex"
	"strings"
	"testing"
)

// 32-byte key, hex-encoded (64 chars).
const testKey = "000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f"

func TestRoundTrip(t *testing.T) {
	c, err := New(testKey)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	plaintext := "evolution-instance-api-key-secret"

	enc, err := c.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}
	if enc == plaintext {
		t.Fatal("ciphertext equals plaintext; encryption did not run")
	}

	dec, err := c.Decrypt(enc)
	if err != nil {
		t.Fatalf("Decrypt: %v", err)
	}
	if dec != plaintext {
		t.Errorf("Decrypt = %q, want %q", dec, plaintext)
	}
}

func TestEncryptUsesFreshNonce(t *testing.T) {
	c, _ := New(testKey)
	a, _ := c.Encrypt("same-input")
	b, _ := c.Encrypt("same-input")
	if a == b {
		t.Fatal("two encryptions of the same plaintext are identical; nonce is reused")
	}
}

func TestDecryptRejectsTamperedCiphertext(t *testing.T) {
	c, _ := New(testKey)
	enc, _ := c.Encrypt("secret")
	raw, _ := hex.DecodeString(enc)
	raw[len(raw)-1] ^= 0xFF // flip a bit in the GCM tag region
	if _, err := c.Decrypt(hex.EncodeToString(raw)); err == nil {
		t.Fatal("Decrypt accepted tampered ciphertext, want error")
	}
}

func TestDecryptRejectsShortCiphertext(t *testing.T) {
	c, _ := New(testKey)
	if _, err := c.Decrypt("00"); err == nil {
		t.Fatal("Decrypt accepted ciphertext shorter than nonce, want error")
	}
}

func TestDecryptRejectsNonHex(t *testing.T) {
	c, _ := New(testKey)
	if _, err := c.Decrypt("zzzz"); err == nil {
		t.Fatal("Decrypt accepted non-hex input, want error")
	}
}

func TestNewRejectsNonHexKey(t *testing.T) {
	if _, err := New("not-hex-key"); err == nil {
		t.Fatal("New accepted a non-hex key, want error")
	}
}

func TestNewRejectsWrongKeyLength(t *testing.T) {
	if _, err := New("0011223344"); err == nil {
		t.Fatal("New accepted a key that is not 32 bytes, want error")
	}
}

func TestPassthroughWhenNoKey(t *testing.T) {
	c, err := New("")
	if err != nil {
		t.Fatalf("New(\"\"): %v", err)
	}
	enc, _ := c.Encrypt("plaintext")
	if enc != "plaintext" {
		t.Errorf("passthrough Encrypt = %q, want plaintext", enc)
	}
	dec, _ := c.Decrypt("plaintext")
	if dec != "plaintext" {
		t.Errorf("passthrough Decrypt = %q, want plaintext", dec)
	}
}

func TestDecryptWrongKeyFails(t *testing.T) {
	c1, _ := New(testKey)
	enc, _ := c1.Encrypt("secret")
	otherKey := strings.Repeat("ab", 32)
	c2, _ := New(otherKey)
	if _, err := c2.Decrypt(enc); err == nil {
		t.Fatal("Decrypt with the wrong key succeeded, want error")
	}
}
