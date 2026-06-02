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
// @Failure  400 {object} httputil.ErrorResponse "invalid request body"
// @Failure  401 {object} httputil.ErrorResponse "missing or invalid token"
// @Failure  403 {object} httputil.ErrorResponse "insufficient role"
// @Failure  409 {object} httputil.ErrorResponse "domain already registered"
// @Failure  429 {object} httputil.ErrorResponse "rate limit exceeded"
// @Failure  500 {object} httputil.ErrorResponse "internal error"
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
