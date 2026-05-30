// Package dtos holds request/response payloads for the tenant module.
package dtos

import "time"

// CreateCompanyRequest is the tenant signup payload.
type CreateCompanyRequest struct {
	Name  string `json:"name" binding:"required,min=2"`
	Slug  string `json:"slug" binding:"required,min=2,alphanumunicode|contains=-"`
	Plan  string `json:"plan" binding:"omitempty"`
	Email string `json:"email" binding:"omitempty,email"`
}

// CompanyResponse describes a company.
type CompanyResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	Status    string    `json:"status"`
	Plan      string    `json:"plan"`
	CreatedAt time.Time `json:"created_at"`
}

// UpdateBrandingRequest updates white-label settings.
type UpdateBrandingRequest struct {
	LogoURL         string `json:"logo_url" binding:"omitempty,url"`
	FaviconURL      string `json:"favicon_url" binding:"omitempty,url"`
	PrimaryColor    string `json:"primary_color" binding:"omitempty,hexcolor"`
	SecondaryColor  string `json:"secondary_color" binding:"omitempty,hexcolor"`
	EmailSenderName string `json:"email_sender_name" binding:"omitempty"`
}

// BrandingResponse is the white-label theme served to the frontend.
type BrandingResponse struct {
	CompanyID       string `json:"company_id"`
	LogoURL         string `json:"logo_url"`
	FaviconURL      string `json:"favicon_url"`
	PrimaryColor    string `json:"primary_color"`
	SecondaryColor  string `json:"secondary_color"`
	EmailSenderName string `json:"email_sender_name"`
}

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
