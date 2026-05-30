package http

import (
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/tenant/dto"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/appctx"
	"github.com/Edu0liver/prototype-healthy-api/pkg/httputil"
	"github.com/gin-gonic/gin"
)

// CreateCompany handles tenant signup (public).
func (h *Handler) CreateCompany(c *gin.Context) {
	var in dto.CreateCompanyRequest
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
