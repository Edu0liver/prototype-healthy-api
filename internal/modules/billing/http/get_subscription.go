package http

import (
	"time"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/billing/dto"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/appctx"
	"github.com/Edu0liver/prototype-healthy-api/pkg/httputil"
	"github.com/gin-gonic/gin"
)

// GetSubscription handles GET /billing/subscription.
// @Summary  Get current subscription
// @Tags     billing
// @Security BearerAuth
// @Produce  json
// @Success  200 {object} dto.SubscriptionResponse
// @Failure  401 {object} httputil.ErrorResponse "missing or invalid token"
// @Failure  403 {object} httputil.ErrorResponse "insufficient role"
// @Failure  404 {object} httputil.ErrorResponse "subscription not found"
// @Router   /billing/subscription [get]
func (h *Handler) GetSubscription(c *gin.Context) {
	companyID := appctx.CompanyID(c.Request.Context())
	sub, plan, err := h.svc.GetSubscription(c.Request.Context(), companyID)
	if err != nil {
		httputil.Fail(c, err)
		return
	}
	httputil.OK(c, dto.SubscriptionResponse{
		PlanCode:           plan.Code,
		PlanName:           plan.Name,
		Status:             sub.Status,
		BillingCycle:       sub.BillingCycle,
		CurrentPeriodStart: sub.CurrentPeriodStart.Format(time.RFC3339),
		CurrentPeriodEnd:   sub.CurrentPeriodEnd.Format(time.RFC3339),
		CancelAtPeriodEnd:  sub.CancelAtPeriodEnd,
		PriceCents:         plan.PriceCents,
		Currency:           plan.Currency,
	})
}
