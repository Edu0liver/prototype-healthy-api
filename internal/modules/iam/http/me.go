package http

import (
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/appctx"
	"github.com/Edu0liver/prototype-healthy-api/pkg/httputil"
	"github.com/gin-gonic/gin"
)

// Me returns the authenticated user.
// @Summary  Current user
// @Tags     users
// @Security BearerAuth
// @Produce  json
// @Success  200
// @Failure  401 {object} httputil.ErrorResponse "missing or invalid token"
// @Failure  429 {object} httputil.ErrorResponse "rate limit exceeded"
// @Failure  500 {object} httputil.ErrorResponse "internal error"
// @Router   /auth/me [get]
func (h *Handler) Me(c *gin.Context) {
	user, err := h.svc.Me(c.Request.Context(), appctx.UserID(c.Request.Context()))
	if err != nil {
		httputil.Fail(c, err)
		return
	}
	httputil.OK(c, userResponse(user))
}
