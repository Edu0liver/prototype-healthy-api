package http

import (
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/iam/dto"
	"github.com/Edu0liver/prototype-healthy-api/pkg/httputil"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Register bootstraps the first admin for a company (public, first-user only).
// @Summary Register first admin
// @Tags    auth
// @Accept  json
// @Produce json
// @Param   body body dto.RegisterAdminRequest true "Admin"
// @Success 201 {object} dto.UserResponse
// @Router  /auth/register [post]
func (h *Handler) Register(c *gin.Context) {
	var in dto.RegisterAdminRequest
	if !httputil.BindJSON(c, &in) {
		return
	}
	companyID, _ := uuid.Parse(in.CompanyID) // validated as uuid by binding tag
	user, err := h.svc.RegisterFirstAdmin(c.Request.Context(), companyID, in.Email, in.Password, in.Name)
	if err != nil {
		httputil.Fail(c, err)
		return
	}
	httputil.Created(c, userResponse(user))
}
