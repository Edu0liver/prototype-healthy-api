package http

import (
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/iam/dto"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/appctx"
	"github.com/Edu0liver/prototype-healthy-api/pkg/httputil"
	"github.com/gin-gonic/gin"
)

// Me returns the authenticated user.
// @Summary  Current user
// @Tags     users
// @Security BearerAuth
// @Produce  json
// @Success  200 {object} dto.UserResponse
// @Router   /auth/me [get]
func (h *Handler) Me(c *gin.Context) {
	user, err := h.svc.Me(c.Request.Context(), appctx.UserID(c.Request.Context()))
	if err != nil {
		httputil.Fail(c, err)
		return
	}
	httputil.OK(c, userResponse(user))
}

// ListUsers lists the tenant's users (admin).
// @Summary  List users
// @Tags     users
// @Security BearerAuth
// @Produce  json
// @Success  200 {object} map[string][]dto.UserResponse
// @Router   /users [get]
func (h *Handler) ListUsers(c *gin.Context) {
	users, err := h.svc.ListUsers(c.Request.Context())
	if err != nil {
		httputil.Fail(c, err)
		return
	}
	out := make([]dto.UserResponse, len(users))
	for i := range users {
		out[i] = userResponse(&users[i])
	}
	httputil.OK(c, gin.H{"users": out})
}
