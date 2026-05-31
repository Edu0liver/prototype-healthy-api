// Package models holds the GORM entities for the tenant module.
package models

import (
	"time"

	"github.com/google/uuid"
)

// Company is a tenant.
type Company struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	Name      string
	Slug      string
	Status    string
	Plan      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (Company) TableName() string { return "companies" }

// CompanyBranding is the white-label theme (1:1 with Company).
type CompanyBranding struct {
	ID              uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CompanyID       uuid.UUID `gorm:"type:uuid;uniqueIndex"`
	LogoURL         string
	FaviconURL      string
	PrimaryColor    string
	SecondaryColor  string
	EmailSenderName string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func (CompanyBranding) TableName() string { return "company_branding" }

// CompanyDomain maps a Host to a tenant.
type CompanyDomain struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey"`
	CompanyID  uuid.UUID `gorm:"type:uuid"`
	Domain     string
	IsPrimary  bool
	VerifiedAt *time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func (CompanyDomain) TableName() string { return "company_domains" }
