package token

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func newManager() *Manager {
	return New("test-secret-value", 15*time.Minute, 720*time.Hour)
}

func TestGenerateAndParseAccess(t *testing.T) {
	m := newManager()
	companyID, userID := uuid.New(), uuid.New()

	tok, err := m.GenerateAccess(companyID, userID, "admin")
	if err != nil {
		t.Fatalf("GenerateAccess: %v", err)
	}

	claims, err := m.Parse(tok)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if claims.Type != TypeAccess {
		t.Errorf("Type = %q, want %q", claims.Type, TypeAccess)
	}
	if claims.CompanyID != companyID.String() {
		t.Errorf("CompanyID = %q, want %q", claims.CompanyID, companyID)
	}
	if claims.UserID != userID.String() {
		t.Errorf("UserID = %q, want %q", claims.UserID, userID)
	}
	if claims.Role != "admin" {
		t.Errorf("Role = %q, want admin", claims.Role)
	}
}

func TestGenerateRefreshHasNoRole(t *testing.T) {
	m := newManager()
	tok, err := m.GenerateRefresh(uuid.New(), uuid.New())
	if err != nil {
		t.Fatalf("GenerateRefresh: %v", err)
	}
	claims, err := m.Parse(tok)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if claims.Type != TypeRefresh {
		t.Errorf("Type = %q, want %q", claims.Type, TypeRefresh)
	}
	if claims.Role != "" {
		t.Errorf("Role = %q, want empty", claims.Role)
	}
}

func TestParseRejectsExpiredToken(t *testing.T) {
	// Negative TTL => token already expired at issue time.
	m := New("test-secret-value", -time.Minute, -time.Minute)
	tok, err := m.GenerateAccess(uuid.New(), uuid.New(), "admin")
	if err != nil {
		t.Fatalf("GenerateAccess: %v", err)
	}
	if _, err := m.Parse(tok); err == nil {
		t.Fatal("Parse accepted an expired token, want error")
	}
}

func TestParseRejectsWrongSecret(t *testing.T) {
	signer := newManager()
	tok, err := signer.GenerateAccess(uuid.New(), uuid.New(), "admin")
	if err != nil {
		t.Fatalf("GenerateAccess: %v", err)
	}
	verifier := New("a-different-secret", 15*time.Minute, 720*time.Hour)
	if _, err := verifier.Parse(tok); err == nil {
		t.Fatal("Parse accepted token signed with a different secret, want error")
	}
}

func TestParseRejectsNoneAlgorithm(t *testing.T) {
	// Forge a token with alg=none — a classic JWT bypass. Parse must reject it.
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
		CompanyID: uuid.New().String(),
		UserID:    uuid.New().String(),
		Type:      TypeAccess,
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
	signed, err := tok.SignedString(jwt.UnsafeAllowNoneSignatureType)
	if err != nil {
		t.Fatalf("sign none: %v", err)
	}
	if _, err := newManager().Parse(signed); err == nil {
		t.Fatal("Parse accepted an alg=none token, want error")
	}
}

func TestParseRejectsTamperedToken(t *testing.T) {
	m := newManager()
	tok, err := m.GenerateAccess(uuid.New(), uuid.New(), "operator")
	if err != nil {
		t.Fatalf("GenerateAccess: %v", err)
	}
	// Flip the last character of the signature.
	tampered := tok[:len(tok)-1]
	if tok[len(tok)-1] == 'a' {
		tampered += "b"
	} else {
		tampered += "a"
	}
	if _, err := m.Parse(tampered); err == nil {
		t.Fatal("Parse accepted a tampered token, want error")
	}
}

func TestParseRejectsGarbage(t *testing.T) {
	if _, err := newManager().Parse("not-a-jwt"); err == nil {
		t.Fatal("Parse accepted garbage, want error")
	}
}

func TestTokensCarryUniqueJTI(t *testing.T) {
	m := newManager()
	a, _ := m.GenerateRefresh(uuid.New(), uuid.New())
	b, _ := m.GenerateRefresh(uuid.New(), uuid.New())
	ca, err := m.Parse(a)
	if err != nil {
		t.Fatalf("parse a: %v", err)
	}
	cb, err := m.Parse(b)
	if err != nil {
		t.Fatalf("parse b: %v", err)
	}
	if ca.ID == "" {
		t.Error("refresh token has empty jti; revocation cannot work")
	}
	if ca.ID == cb.ID {
		t.Error("two tokens share the same jti; jti is not unique")
	}
}

func TestGenerateInvite(t *testing.T) {
	m := newManager()
	tok, err := m.GenerateInvite(uuid.New(), uuid.New(), time.Hour)
	if err != nil {
		t.Fatalf("GenerateInvite: %v", err)
	}
	claims, err := m.Parse(tok)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if claims.Type != TypeInvite {
		t.Errorf("Type = %q, want %q", claims.Type, TypeInvite)
	}
}
