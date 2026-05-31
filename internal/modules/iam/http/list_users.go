package http

import (
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/iam/dto"
	"github.com/Edu0liver/prototype-healthy-api/pkg/httputil"
	"github.com/gin-gonic/gin"
)

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
