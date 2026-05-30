// Package dto holds request/response payloads for the tenant module.
package dto

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
