package http

import (
	"github.com/Edu0liver/prototype-healthy-api/pkg/httputil"
	"github.com/gin-gonic/gin"
)

// GetBrandingByHost serves the white-label theme for a Host (public).
// @Summary  Get branding by host
// @Tags     tenant
// @Produce  json
// @Param    host query string false "Hostname (defaults to request Host header)"
// @Success  200
// @Failure  404 {object} httputil.ErrorResponse "branding not found"
// @Failure  500 {object} httputil.ErrorResponse "internal error"
// @Router   /branding/host [get]
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
