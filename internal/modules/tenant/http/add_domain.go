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
