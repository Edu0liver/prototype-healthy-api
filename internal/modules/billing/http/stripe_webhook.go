package http

import (
	"github.com/Edu0liver/prototype-healthy-api/pkg/httputil"
	"github.com/gin-gonic/gin"
)

// StripeWebhook handles POST /webhooks/stripe. It is public (no auth/tenant
// middleware) and authenticated instead by the Stripe-Signature header, which
// the service verifies against the webhook secret.
// @Summary  Stripe billing webhook
// @Tags     billing
// @Accept   json
// @Produce  json
// @Param    Stripe-Signature header string true "Stripe signature"
// @Success  200 {object} map[string]bool
// @Failure  400 {object} httputil.ErrorResponse "invalid signature"
// @Failure  503 {object} httputil.ErrorResponse "billing gateway not configured"
// @Router   /webhooks/stripe [post]
func (h *Handler) StripeWebhook(c *gin.Context) {
	payload, err := c.GetRawData()
	if err != nil {
		httputil.Fail(c, httputil.BadRequest("cannot read body"))
		return
	}
	sig := c.GetHeader("Stripe-Signature")
	if err := h.svc.HandleWebhook(c.Request.Context(), payload, sig); err != nil {
		httputil.Fail(c, err)
		return
	}
	httputil.OK(c, gin.H{"received": true})
}
