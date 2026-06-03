package http

import (
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/middleware"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes mounts /billing (admin only, tenant tx) and the public Stripe
// webhook (no auth — verified by the Stripe-Signature header).
func RegisterRoutes(e *gin.RouterGroup, h *Handler, mw *middleware.Middleware) {
	g := e.Group("/billing", mw.Auth(), mw.Tenant(), mw.RBAC(middleware.RoleAdmin))
	g.GET("/subscription", h.GetSubscription)
	g.GET("/usage", h.GetUsage)
	g.POST("/checkout", h.CreateCheckout)

	// Public gateway webhook (system-scoped; authenticated by signature).
	e.POST("/webhooks/stripe", h.StripeWebhook)
}
