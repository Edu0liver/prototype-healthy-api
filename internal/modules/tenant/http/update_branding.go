package http

import (
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/tenant/dto"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/appctx"
	"github.com/Edu0liver/prototype-healthy-api/pkg/httputil"
	"github.com/gin-gonic/gin"
)

// UpdateBranding upserts the authenticated tenant's branding.
// @Summary  Update branding
// @Tags     tenant
// @Security BearerAuth
// @Accept   json
// @Produce  json
// @Param    body body dto.UpdateBrandingRequest true "Branding"
// @Success  200 {object} dto.BrandingResponse
// @Failure  400 {object} httputil.ErrorResponse "invalid request body"
// @Failure  401 {object} httputil.ErrorResponse "missing or invalid token"
// @Failure  403 {object} httputil.ErrorResponse "insufficient role"
// @Failure  429 {object} httputil.ErrorResponse "rate limit exceeded"
// @Failure  500 {object} httputil.ErrorResponse "internal error"
// @Router   /branding [put]
func (h *Handler) UpdateBranding(c *gin.Context) {
	var in dto.UpdateBrandingRequest
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
