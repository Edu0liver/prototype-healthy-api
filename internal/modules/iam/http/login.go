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
// @Failure  400 {object} httputil.ErrorResponse "invalid request body"
// @Failure  401 {object} httputil.ErrorResponse "invalid credentials"
// @Failure  403 {object} httputil.ErrorResponse "user disabled"
// @Failure  429 {object} httputil.ErrorResponse "too many attempts (IP or account lockout)"
// @Failure  500 {object} httputil.ErrorResponse "internal error"
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
