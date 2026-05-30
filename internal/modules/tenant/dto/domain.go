package dto

import "time"

// AddDomainRequest registers a custom domain for the tenant.
type AddDomainRequest struct {
	Domain    string `json:"domain" binding:"required,fqdn"`
	IsPrimary bool   `json:"is_primary"`
}

// DomainResponse describes a registered domain.
type DomainResponse struct {
	ID         string     `json:"id"`
	Domain     string     `json:"domain"`
	IsPrimary  bool       `json:"is_primary"`
	VerifiedAt *time.Time `json:"verified_at"`
}
