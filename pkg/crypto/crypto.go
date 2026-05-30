// Package crypto provides AES-256-GCM encryption for secrets at rest
// (e.g. Evolution instance API keys stored in channels.evolution_apikey_enc).
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
)

// Cipher encrypts and decrypts short secrets using AES-256-GCM.
type Cipher struct {
	aead cipher.AEAD
}

// New builds a Cipher from a 32-byte hex key. When the key is empty (local dev),
// encryption is disabled and a passthrough cipher is returned.
func New(keyHex string) (*Cipher, error) {
	key := keyHex
	if key == "" {
		return &Cipher{}, nil // passthrough (dev only)
	}
	raw, err := hex.DecodeString(key)
	if err != nil {
		return nil, fmt.Errorf("crypto: ENCRYPTION_KEY must be hex: %w", err)
	}
	if len(raw) != 32 {
		return nil, errors.New("crypto: ENCRYPTION_KEY must decode to 32 bytes (AES-256)")
	}
	block, err := aes.NewCipher(raw)
	if err != nil {
		return nil, fmt.Errorf("crypto: new cipher: %w", err)
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("crypto: new gcm: %w", err)
	}
	return &Cipher{aead: aead}, nil
}

// Encrypt returns hex(nonce||ciphertext). With no key configured it returns the
// plaintext unchanged (dev passthrough).
func (c *Cipher) Encrypt(plaintext string) (string, error) {
	if c.aead == nil {
		return plaintext, nil
	}
	nonce := make([]byte, c.aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	ct := c.aead.Seal(nonce, nonce, []byte(plaintext), nil)
	return hex.EncodeToString(ct), nil
}

// Decrypt reverses Encrypt.
func (c *Cipher) Decrypt(enc string) (string, error) {
	if c.aead == nil {
		return enc, nil
	}
	raw, err := hex.DecodeString(enc)
	if err != nil {
		return "", fmt.Errorf("crypto: decode: %w", err)
	}
	ns := c.aead.NonceSize()
	if len(raw) < ns {
		return "", errors.New("crypto: ciphertext too short")
	}
	nonce, ct := raw[:ns], raw[ns:]
	pt, err := c.aead.Open(nil, nonce, ct, nil)
	if err != nil {
		return "", fmt.Errorf("crypto: open: %w", err)
	}
	return string(pt), nil
}
