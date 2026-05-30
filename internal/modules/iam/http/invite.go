package http

import (
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/iam/dto"
	"github.com/Edu0liver/prototype-healthy-api/pkg/httputil"
	"github.com/gin-gonic/gin"
)

// Invite creates an invited user (admin).
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
