// Package storage abstracts object storage for RAG uploads and branding assets.
// Objects are namespaced per tenant: tenant/<company_id>/... . A local-fs
// implementation backs development; an S3 driver can be added behind the same
// interface later.
package storage

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Edu0liver/prototype-healthy-api/internal/platform/config"
	"github.com/google/uuid"
)

// Storage is the object storage contract.
type Storage interface {
	Put(ctx context.Context, companyID uuid.UUID, key string, data []byte) (path string, err error)
	Get(ctx context.Context, path string) ([]byte, error)
	Delete(ctx context.Context, path string) error
}

// TenantPrefix builds the per-tenant key prefix.
func TenantPrefix(companyID uuid.UUID, key string) string {
	return filepath.ToSlash(filepath.Join("tenant", companyID.String(), key))
}

// LocalStorage stores objects on the local filesystem (dev).
type LocalStorage struct {
	root string
}

// NewLocal builds the local-fs storage rooted at the configured path.
func NewLocal(cfg *config.Config) (*LocalStorage, error) {
	if err := os.MkdirAll(cfg.Storage.LocalPath, 0o755); err != nil {
		return nil, fmt.Errorf("storage: mkdir root: %w", err)
	}
	return &LocalStorage{root: cfg.Storage.LocalPath}, nil
}

func (s *LocalStorage) Put(_ context.Context, companyID uuid.UUID, key string, data []byte) (string, error) {
	rel := TenantPrefix(companyID, key)
	abs := filepath.Join(s.root, rel)
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		return "", err
	}
	if err := os.WriteFile(abs, data, 0o644); err != nil {
		return "", err
	}
	return rel, nil
}

func (s *LocalStorage) Get(_ context.Context, path string) ([]byte, error) {
	return os.ReadFile(filepath.Join(s.root, path))
}

func (s *LocalStorage) Delete(_ context.Context, path string) error {
	err := os.Remove(filepath.Join(s.root, path))
	if os.IsNotExist(err) {
		return nil
	}
	return err
}
