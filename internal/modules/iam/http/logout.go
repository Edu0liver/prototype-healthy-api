package http

import (
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/iam/dto"
	"github.com/Edu0liver/prototype-healthy-api/pkg/httputil"
	"github.com/gin-gonic/gin"
)

// Logout revokes the supplied refresh token (public, idempotent).
// @Summary  Logout
// @Tags     auth
// @Accept   json
// @Produce  json
// @Param    body body dto.RefreshRequest true "Refresh token to revoke"
// @Success  204
// @Failure  400 {object} httputil.ErrorResponse "invalid request body"
// @Failure  500 {object} httputil.ErrorResponse "internal error"
// @Router   /auth/logout [post]
func (h *Handler) Logout(c *gin.Context) {
	var in dto.RefreshRequest
	if !httputil.BindJSON(c, &in) {
		return
	}
	if err := h.svc.Logout(c.Request.Context(), in.RefreshToken); err != nil {
		httputil.Fail(c, err)
		return
	}
	httputil.NoContent(c)
}
