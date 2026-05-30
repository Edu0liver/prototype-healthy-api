// Package token issues and verifies JWT access/refresh tokens.
package token

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Token types.
const (
	TypeAccess  = "access"
	TypeRefresh = "refresh"
	TypeInvite  = "invite"
)

// Claims is the JWT payload.
type Claims struct {
	jwt.RegisteredClaims
	CompanyID string `json:"company_id"`
	UserID    string `json:"user_id"`
	Role      string `json:"role,omitempty"`
	Type      string `json:"typ"`
}

// Manager signs and parses tokens.
type Manager struct {
	secret     []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
}

// New builds a Manager from a signing secret and token TTLs.
func New(secret string, accessTTL, refreshTTL time.Duration) *Manager {
	return &Manager{
		secret:     []byte(secret),
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
	}
}

// GenerateAccess issues a short-lived access token.
func (m *Manager) GenerateAccess(companyID, userID uuid.UUID, role string) (string, error) {
	return m.sign(companyID, userID, role, TypeAccess, m.accessTTL)
}

// GenerateRefresh issues a long-lived refresh token.
func (m *Manager) GenerateRefresh(companyID, userID uuid.UUID) (string, error) {
	return m.sign(companyID, userID, "", TypeRefresh, m.refreshTTL)
}

// GenerateInvite issues a single-use invite token carrying the new user's id.
func (m *Manager) GenerateInvite(companyID, userID uuid.UUID, ttl time.Duration) (string, error) {
	return m.sign(companyID, userID, "", TypeInvite, ttl)
}

func (m *Manager) sign(companyID, userID uuid.UUID, role, typ string, ttl time.Duration) (string, error) {
	now := time.Now()
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
		CompanyID: companyID.String(),
		UserID:    userID.String(),
		Role:      role,
		Type:      typ,
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(m.secret)
}

// Parse validates the signature/expiry and returns the claims.
func (m *Manager) Parse(tokenStr string) (*Claims, error) {
	claims := &Claims{}
	_, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("token: unexpected signing method: %v", t.Header["alg"])
		}
		return m.secret, nil
	})
	if err != nil {
		return nil, err
	}
	return claims, nil
}

// ErrWrongTokenType indicates a refresh token used where access was expected or vice versa.
var ErrWrongTokenType = errors.New("token: wrong token type")
