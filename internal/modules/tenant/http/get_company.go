package http

import (
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/appctx"
	"github.com/Edu0liver/prototype-healthy-api/pkg/httputil"
	"github.com/gin-gonic/gin"
)

// GetCompany returns the authenticated tenant's company.
// @Summary  Get company
// @Tags     tenant
// @Security BearerAuth
// @Produce  json
// @Success  200
// @Router   /company [get]
func (h *Handler) GetCompany(c *gin.Context) {
	id := appctx.CompanyID(c.Request.Context())
	company, err := h.svc.GetCompany(c.Request.Context(), id)
	if err != nil {
		httputil.Fail(c, err)
		return
	}
	httputil.OK(c, companyResponse(company))
}
