package http

import (
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/iam/dto"
	"github.com/Edu0liver/prototype-healthy-api/pkg/httputil"
	"github.com/gin-gonic/gin"
)

// Invite creates an invited user (admin).
// @Summary  Invite user
// @Tags     users
// @Security BearerAuth
// @Accept   json
// @Produce  json
// @Param    body body dto.InviteRequest true "Invite"
// @Success  201 {object} dto.UserResponse
// @Failure  400 {object} httputil.ErrorResponse "invalid role or body"
// @Failure  401 {object} httputil.ErrorResponse "missing or invalid token"
// @Failure  403 {object} httputil.ErrorResponse "insufficient role"
// @Failure  409 {object} httputil.ErrorResponse "email already in use"
// @Failure  429 {object} httputil.ErrorResponse "rate limit exceeded"
// @Failure  500 {object} httputil.ErrorResponse "internal error"
// @Router   /users [post]
func (h *Handler) Invite(c *gin.Context) {
	var in dto.InviteRequest
	if !httputil.BindJSON(c, &in) {
		return
	}
	user, err := h.svc.Invite(c.Request.Context(), in.Email, in.Name, in.Role)
	if err != nil {
		httputil.Fail(c, err)
		return
	}
	httputil.Created(c, userResponse(user))
}
