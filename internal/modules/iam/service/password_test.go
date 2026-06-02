package service

import (
	"strings"
	"testing"
)

func TestHashPasswordProducesPHCFormat(t *testing.T) {
	hash, err := hashPassword("correct horse battery staple")
	if err != nil {
		t.Fatalf("hashPassword: %v", err)
	}
	if !strings.HasPrefix(hash, "$argon2id$") {
		t.Errorf("hash %q does not start with $argon2id$", hash)
	}
	if got := len(strings.Split(hash, "$")); got != 6 {
		t.Errorf("hash has %d $-separated parts, want 6", got)
	}
}

func TestHashPasswordIsSalted(t *testing.T) {
	a, _ := hashPassword("same-password")
	b, _ := hashPassword("same-password")
	if a == b {
		t.Fatal("two hashes of the same password are identical; salt is not random")
	}
}

func TestVerifyPasswordCorrect(t *testing.T) {
	hash, _ := hashPassword("s3cret-pass")
	ok, err := verifyPassword("s3cret-pass", hash)
	if err != nil {
		t.Fatalf("verifyPassword: %v", err)
	}
	if !ok {
		t.Fatal("verifyPassword returned false for the correct password")
	}
}

func TestVerifyPasswordWrong(t *testing.T) {
	hash, _ := hashPassword("s3cret-pass")
	ok, err := verifyPassword("wrong-pass", hash)
	if err != nil {
		t.Fatalf("verifyPassword: %v", err)
	}
	if ok {
		t.Fatal("verifyPassword returned true for the wrong password")
	}
}

func TestVerifyPasswordRejectsMalformedHash(t *testing.T) {
	cases := map[string]string{
		"empty":          "",
		"not argon":      "$bcrypt$v=19$m=65536,t=1,p=4$abc$def",
		"too few parts":  "$argon2id$v=19$m=65536",
		"bad params":     "$argon2id$v=19$not-params$c2FsdA$aGFzaA",
		"bad salt b64":   "$argon2id$v=19$m=65536,t=1,p=4$!!!$aGFzaA",
		"bad hash b64":   "$argon2id$v=19$m=65536,t=1,p=4$c2FsdA$!!!",
	}
	for name, encoded := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := verifyPassword("pw", encoded); err == nil {
				t.Errorf("verifyPassword(%q) returned nil error, want error", encoded)
			}
		})
	}
}
