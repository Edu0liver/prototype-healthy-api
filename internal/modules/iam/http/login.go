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
	tokens, user, err := h.svc.Login(c.Request.Context(), in.Email, in.Password)
	if err != nil {
		httputil.Fail(c, err)
		return
	}
	httputil.OK(c, tokenResponse(tokens, user))
}
