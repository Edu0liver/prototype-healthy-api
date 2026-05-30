// Package http exposes the tenant module's Gin handlers (split per resource).
package http

import (
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/tenant/dto"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/tenant/infra/models"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/tenant/service"
)

// Handler serves tenant + white-label endpoints.
type Handler struct {
	svc *service.Service
}

// NewHandler builds the handler.
func NewHandler(svc *service.Service) *Handler { return &Handler{svc: svc} }

func companyResponse(c *models.Company) dto.CompanyResponse {
	return dto.CompanyResponse{
		ID: c.ID.String(), Name: c.Name, Slug: c.Slug,
		Status: c.Status, Plan: c.Plan, CreatedAt: c.CreatedAt,
	}
}

func brandingResponse(b *models.CompanyBranding) dto.BrandingResponse {
	return dto.BrandingResponse{
		CompanyID: b.CompanyID.String(), LogoURL: b.LogoURL, FaviconURL: b.FaviconURL,
		PrimaryColor: b.PrimaryColor, SecondaryColor: b.SecondaryColor, EmailSenderName: b.EmailSenderName,
	}
}

func domainResponse(d *models.CompanyDomain) dto.DomainResponse {
	return dto.DomainResponse{
		ID: d.ID.String(), Domain: d.Domain, IsPrimary: d.IsPrimary, VerifiedAt: d.VerifiedAt,
	}
}
