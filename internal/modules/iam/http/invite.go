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

// AcceptInvite sets the password for an invited user (public).
// @Summary  Accept invite
// @Tags     auth
// @Accept   json
// @Produce  json
// @Param    body body dto.AcceptInviteRequest true "Token and password"
// @Success  204
// @Router   /auth/accept-invite [post]
func (h *Handler) AcceptInvite(c *gin.Context) {
	var in dto.AcceptInviteRequest
	if !httputil.BindJSON(c, &in) {
		return
	}
	if err := h.svc.AcceptInvite(c.Request.Context(), in.Token, in.Password); err != nil {
		httputil.Fail(c, err)
		return
	}
	httputil.NoContent(c)
}
