// Package storage abstracts object storage for RAG uploads and branding assets.
// Objects are namespaced per tenant: tenant/<company_id>/... . A MinIO
// (S3-compatible) driver backs all environments behind this interface.
package storage

import (
	"context"
	"path/filepath"

	"github.com/google/uuid"
)

// Storage is the object storage contract.
type Storage interface {
	Put(ctx context.Context, companyID uuid.UUID, key string, data []byte) (path string, err error)
	Get(ctx context.Context, path string) ([]byte, error)
	Delete(ctx context.Context, path string) error
}

// TenantPrefix builds the per-tenant object key prefix.
func TenantPrefix(companyID uuid.UUID, key string) string {
	return filepath.ToSlash(filepath.Join("tenant", companyID.String(), key))
}
