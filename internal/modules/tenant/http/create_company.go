package http

import (
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/tenant/dto"
	"github.com/Edu0liver/prototype-healthy-api/pkg/httputil"
	"github.com/gin-gonic/gin"
)

// CreateCompany handles tenant signup (public).
// @Summary  Create company
// @Tags     tenant
// @Accept   json
// @Produce  json
// @Param    body body dto.CreateCompanyRequest true "Company"
// @Success  201 {object} dto.CompanyResponse
// @Failure  400 {object} httputil.ErrorResponse "invalid request body"
// @Failure  409 {object} httputil.ErrorResponse "slug already in use"
// @Failure  500 {object} httputil.ErrorResponse "internal error"
// @Router   /companies [post]
func (h *Handler) CreateCompany(c *gin.Context) {
	var in dto.CreateCompanyRequest
	if !httputil.BindJSON(c, &in) {
		return
	}
	company, err := h.svc.CreateCompany(c.Request.Context(), in)
	if err != nil {
		httputil.Fail(c, err)
		return
	}
	httputil.Created(c, companyResponse(company))
}
