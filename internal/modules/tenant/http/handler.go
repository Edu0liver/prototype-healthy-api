// Package http exposes the tenant module's Gin handlers.
package http

import (
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/tenant/dtos"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/tenant/infra/models"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/tenant/services"
	"github.com/Edu0liver/prototype-healthy-api/internal/platform/appctx"
	"github.com/Edu0liver/prototype-healthy-api/internal/platform/httputil"
	"github.com/gin-gonic/gin"
)

// Handler serves tenant + white-label endpoints.
type Handler struct {
	svc *services.Service
}

// NewHandler builds the handler.
func NewHandler(svc *services.Service) *Handler { return &Handler{svc: svc} }

// CreateCompany handles tenant signup (public).
func (h *Handler) CreateCompany(c *gin.Context) {
	var in dtos.CreateCompanyRequest
	if !httputil.BindJSON(c, &in) {
		return
	}
	company, err := h.svc.CreateCompany(c.Request.Context(), in)
	if err != nil {
		httputil.Fail(c, err)
		return
	}
	httputil.Created(c, companyResponse(company))
}

// GetCompany returns the authenticated tenant's company.
func (h *Handler) GetCompany(c *gin.Context) {
	id := appctx.CompanyID(c.Request.Context())
	company, err := h.svc.GetCompany(c.Request.Context(), id)
	if err != nil {
		httputil.Fail(c, err)
		return
	}
	httputil.OK(c, companyResponse(company))
}

// GetBrandingByHost serves the white-label theme for a Host (public).
func (h *Handler) GetBrandingByHost(c *gin.Context) {
	host := c.Query("host")
	if host == "" {
		host = c.Request.Host
	}
	b, err := h.svc.GetBrandingByHost(c.Request.Context(), host)
	if err != nil {
		httputil.Fail(c, err)
		return
	}
	httputil.OK(c, brandingResponse(b))
}

// UpdateBranding upserts the authenticated tenant's branding.
func (h *Handler) UpdateBranding(c *gin.Context) {
	var in dtos.UpdateBrandingRequest
	if !httputil.BindJSON(c, &in) {
		return
	}
	id := appctx.CompanyID(c.Request.Context())
	b, err := h.svc.UpdateBranding(c.Request.Context(), id, in)
	if err != nil {
		httputil.Fail(c, err)
		return
	}
	httputil.OK(c, brandingResponse(b))
}

// AddDomain registers a custom domain for the tenant.
func (h *Handler) AddDomain(c *gin.Context) {
	var in dtos.AddDomainRequest
	if !httputil.BindJSON(c, &in) {
		return
	}
	id := appctx.CompanyID(c.Request.Context())
	d, err := h.svc.AddDomain(c.Request.Context(), id, in)
	if err != nil {
		httputil.Fail(c, err)
		return
	}
	httputil.Created(c, domainResponse(d))
}

// ListDomains lists the tenant's domains.
func (h *Handler) ListDomains(c *gin.Context) {
	id := appctx.CompanyID(c.Request.Context())
	domains, err := h.svc.ListDomains(c.Request.Context(), id)
	if err != nil {
		httputil.Fail(c, err)
		return
	}
	out := make([]dtos.DomainResponse, len(domains))
	for i := range domains {
		out[i] = domainResponse(&domains[i])
	}
	httputil.OK(c, gin.H{"domains": out})
}

// ---- mappers --------------------------------------------------------------

func companyResponse(c *models.Company) dtos.CompanyResponse {
	return dtos.CompanyResponse{
		ID: c.ID.String(), Name: c.Name, Slug: c.Slug,
		Status: c.Status, Plan: c.Plan, CreatedAt: c.CreatedAt,
	}
}

func brandingResponse(b *models.CompanyBranding) dtos.BrandingResponse {
	return dtos.BrandingResponse{
		CompanyID: b.CompanyID.String(), LogoURL: b.LogoURL, FaviconURL: b.FaviconURL,
		PrimaryColor: b.PrimaryColor, SecondaryColor: b.SecondaryColor, EmailSenderName: b.EmailSenderName,
	}
}

func domainResponse(d *models.CompanyDomain) dtos.DomainResponse {
	return dtos.DomainResponse{
		ID: d.ID.String(), Domain: d.Domain, IsPrimary: d.IsPrimary, VerifiedAt: d.VerifiedAt,
	}
}
