package http

import (
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/appctx"
	"github.com/Edu0liver/prototype-healthy-api/pkg/httputil"
	"github.com/gin-gonic/gin"
)

// GetBranding returns the authenticated tenant's branding.
// @Summary  Get own branding
// @Tags     tenant
// @Security BearerAuth
// @Produce  json
// @Success  200
// @Failure  401 {object} httputil.ErrorResponse "missing or invalid token"
// @Failure  403 {object} httputil.ErrorResponse "insufficient role"
// @Failure  404 {object} httputil.ErrorResponse "branding not found"
// @Failure  429 {object} httputil.ErrorResponse "rate limit exceeded"
// @Failure  500 {object} httputil.ErrorResponse "internal error"
// @Router   /branding [get]
func (h *Handler) GetBranding(c *gin.Context) {
	id := appctx.CompanyID(c.Request.Context())
	b, err := h.svc.GetBranding(c.Request.Context(), id)
	if err != nil {
		httputil.Fail(c, err)
		return
	}
	httputil.OK(c, brandingResponse(b))
}
