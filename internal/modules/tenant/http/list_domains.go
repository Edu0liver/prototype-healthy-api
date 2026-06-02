package http

import (
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/tenant/dto"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/appctx"
	"github.com/Edu0liver/prototype-healthy-api/pkg/httputil"
	"github.com/gin-gonic/gin"
)

// ListDomains lists the tenant's domains.
// @Summary  List domains
// @Tags     tenant
// @Security BearerAuth
// @Produce  json
// @Success  200 {object} map[string][]dto.DomainResponse
// @Failure  401 {object} httputil.ErrorResponse "missing or invalid token"
// @Failure  403 {object} httputil.ErrorResponse "insufficient role"
// @Failure  429 {object} httputil.ErrorResponse "rate limit exceeded"
// @Failure  500 {object} httputil.ErrorResponse "internal error"
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
