package http

import (
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/iam/dto"
	"github.com/Edu0liver/prototype-healthy-api/pkg/httputil"
	"github.com/gin-gonic/gin"
)

// Refresh issues a new token pair from a refresh token (public).
// @Summary Refresh token
// @Tags    auth
// @Accept  json
// @Produce json
// @Param   body body dto.RefreshRequest true "Refresh token"
// @Success 200 {object} dto.TokenResponse
// @Router  /auth/refresh [post]
func (h *Handler) Refresh(c *gin.Context) {
	var in dto.RefreshRequest
	if !httputil.BindJSON(c, &in) {
		return
	}
	tokens, err := h.svc.Refresh(c.Request.Context(), in.RefreshToken)
	if err != nil {
		httputil.Fail(c, err)
		return
	}
	httputil.OK(c, dto.TokenResponse{
		AccessToken: tokens.Access, RefreshToken: tokens.Refresh, TokenType: "Bearer",
	})
}
