package http

import (
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/middleware"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes mounts /billing. Reads (plans/subscription/usage) are open to
// any authenticated tenant user so the app can show plan status to operators
// too; only checkout is admin-only. NOT subscription-gated — a tenant with an
// inactive plan must still reach billing to pay. Plus the public Stripe webhook.
func RegisterRoutes(e *gin.RouterGroup, h *Handler, mw *middleware.Middleware) {
	g := e.Group("/billing", mw.Auth(), mw.Tenant())
	g.GET("/plans", h.GetPlans)
	g.GET("/subscription", h.GetSubscription)
	g.GET("/usage", h.GetUsage)
	g.POST("/checkout", mw.RBAC(middleware.RoleAdmin), h.CreateCheckout)

	// Public plan catalogue (no auth) — powers the marketing/landing pricing.
	e.GET("/plans", h.GetPlans)

	// Public gateway webhook (system-scoped; authenticated by signature).
	e.POST("/webhooks/stripe", h.StripeWebhook)
}
