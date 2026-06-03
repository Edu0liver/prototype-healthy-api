package http

import (
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/billing/dto"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/appctx"
	"github.com/Edu0liver/prototype-healthy-api/pkg/httputil"
	"github.com/gin-gonic/gin"
)

// CreateCheckout handles POST /billing/checkout.
// @Summary  Start a subscription checkout
// @Tags     billing
// @Security BearerAuth
// @Accept   json
// @Produce  json
// @Param    body body dto.CreateCheckoutRequest true "Plan to subscribe to"
// @Success  200 {object} dto.CheckoutResponse
// @Failure  400 {object} httputil.ErrorResponse "invalid request"
// @Failure  401 {object} httputil.ErrorResponse "missing or invalid token"
// @Failure  403 {object} httputil.ErrorResponse "insufficient role"
// @Failure  404 {object} httputil.ErrorResponse "plan not found"
// @Failure  409 {object} httputil.ErrorResponse "plan not purchasable"
// @Failure  503 {object} httputil.ErrorResponse "billing gateway not configured"
// @Router   /billing/checkout [post]
func (h *Handler) CreateCheckout(c *gin.Context) {
	var in dto.CreateCheckoutRequest
	if !httputil.BindJSON(c, &in) {
		return
	}
	companyID := appctx.CompanyID(c.Request.Context())
	url, err := h.svc.CreateCheckout(c.Request.Context(), companyID, in.PlanCode, "")
	if err != nil {
		httputil.Fail(c, err)
		return
	}
	httputil.OK(c, dto.CheckoutResponse{CheckoutURL: url})
}
