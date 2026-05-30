// Package appctx carries the authenticated identity (company, user, role)
// through context.Context so services and repositories can enforce tenancy.
package appctx

import (
	"context"

	"github.com/google/uuid"
)

// Identity is the resolved caller identity for a request or job.
type Identity struct {
	CompanyID uuid.UUID
	UserID    uuid.UUID
	Role      string
}

type identityKey struct{}

// With stores the identity in the context.
func With(ctx context.Context, id Identity) context.Context {
	return context.WithValue(ctx, identityKey{}, id)
}

// From returns the identity if present.
func From(ctx context.Context) (Identity, bool) {
	id, ok := ctx.Value(identityKey{}).(Identity)
	return id, ok
}

// CompanyID returns the tenant id or uuid.Nil if unauthenticated.
func CompanyID(ctx context.Context) uuid.UUID {
	if id, ok := From(ctx); ok {
		return id.CompanyID
	}
	return uuid.Nil
}

// UserID returns the user id or uuid.Nil.
func UserID(ctx context.Context) uuid.UUID {
	if id, ok := From(ctx); ok {
		return id.UserID
	}
	return uuid.Nil
}

// Role returns the caller role.
func Role(ctx context.Context) string {
	if id, ok := From(ctx); ok {
		return id.Role
	}
	return ""
}
