package http

import (
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/billing/dto"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/appctx"
	"github.com/Edu0liver/prototype-healthy-api/pkg/httputil"
	"github.com/gin-gonic/gin"
)

// CreatePortal handles POST /billing/portal.
// @Summary  Open Stripe Billing Portal
// @Tags     billing
// @Security BearerAuth
// @Produce  json
// @Success  200 {object} dto.PortalResponse
// @Failure  401 {object} httputil.ErrorResponse "missing or invalid token"
// @Failure  403 {object} httputil.ErrorResponse "insufficient role"
// @Failure  404 {object} httputil.ErrorResponse "no subscription found"
// @Failure  503 {object} httputil.ErrorResponse "billing gateway not configured"
// @Router   /billing/portal [post]
func (h *Handler) CreatePortal(c *gin.Context) {
	companyID := appctx.CompanyID(c.Request.Context())
	portalURL, err := h.svc.CreatePortalSession(c.Request.Context(), companyID)
	if err != nil {
		httputil.Fail(c, err)
		return
	}
	httputil.OK(c, dto.PortalResponse{PortalURL: portalURL})
}
