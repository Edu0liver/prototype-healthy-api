package http

import (
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/iam/dto"
	"github.com/Edu0liver/prototype-healthy-api/pkg/httputil"
	"github.com/gin-gonic/gin"
)

// Login authenticates and returns tokens (public).
// @Summary Login
// @Tags    auth
// @Accept  json
// @Produce json
// @Param   body body dto.LoginRequest true "Credentials"
// @Success 200 {object} dto.TokenResponse
// @Router  /auth/login [post]
func (h *Handler) Login(c *gin.Context) {
	var in dto.LoginRequest
	if !httputil.BindJSON(c, &in) {
		return
	}
	tokens, user, err := h.svc.Login(c.Request.Context(), in.CompanySlug, in.Email, in.Password)
	if err != nil {
		httputil.Fail(c, err)
		return
	}
	httputil.OK(c, tokenResponse(tokens, user))
}

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
