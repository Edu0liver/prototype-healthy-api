package http

import (
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/iam/dto"
	"github.com/Edu0liver/prototype-healthy-api/pkg/httputil"
	"github.com/gin-gonic/gin"
)

// AcceptInvite sets the password for an invited user (public).
// @Summary  Accept invite
// @Tags     auth
// @Accept   json
// @Produce  json
// @Param    body body dto.AcceptInviteRequest true "Token and password"
// @Success  204
// @Failure  400 {object} httputil.ErrorResponse "invalid or expired invite"
// @Failure  500 {object} httputil.ErrorResponse "internal error"
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
