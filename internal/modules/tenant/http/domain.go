package http

import (
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/tenant/dto"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/appctx"
	"github.com/Edu0liver/prototype-healthy-api/pkg/httputil"
	"github.com/gin-gonic/gin"
)

// AddDomain registers a custom domain for the tenant.
// @Summary  Add domain
// @Tags     tenant
// @Security BearerAuth
// @Accept   json
// @Produce  json
// @Param    body body dto.AddDomainRequest true "Domain"
// @Success  201 {object} dto.DomainResponse
// @Router   /domains [post]
func (h *Handler) AddDomain(c *gin.Context) {
	var in dto.AddDomainRequest
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
// @Summary  List domains
// @Tags     tenant
// @Security BearerAuth
// @Produce  json
// @Success  200 {object} map[string][]dto.DomainResponse
// @Router   /domains [get]
func (h *Handler) ListDomains(c *gin.Context) {
	id := appctx.CompanyID(c.Request.Context())
	domains, err := h.svc.ListDomains(c.Request.Context(), id)
	if err != nil {
		httputil.Fail(c, err)
		return
	}
	out := make([]dto.DomainResponse, len(domains))
	for i := range domains {
		out[i] = domainResponse(&domains[i])
	}
	httputil.OK(c, gin.H{"domains": out})
}
