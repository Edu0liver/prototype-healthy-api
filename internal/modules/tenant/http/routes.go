package http

import (
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/middleware"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes mounts the tenant routes. Public routes (signup, branding) need
// no auth; the rest require an authenticated admin within a tenant transaction.
func RegisterRoutes(e *gin.RouterGroup, h *Handler, mw *middleware.Middleware) {
	// Public.
	e.POST("/companies", h.CreateCompany)
	e.GET("/branding", h.GetBrandingByHost)

	// Authenticated tenant admin.
	admin := e.Group("/", mw.Auth(), mw.Tenant(), mw.RBAC(middleware.RoleAdmin))
	admin.GET("company", h.GetCompany)
	admin.PUT("branding", h.UpdateBranding)
	admin.POST("domains", h.AddDomain)
	admin.GET("domains", h.ListDomains)
}
