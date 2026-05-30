package http

import (
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/tenant/dto"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/appctx"
	"github.com/Edu0liver/prototype-healthy-api/pkg/httputil"
	"github.com/gin-gonic/gin"
)

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
