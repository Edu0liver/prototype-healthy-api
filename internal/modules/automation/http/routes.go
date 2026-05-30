package http

import (
	"github.com/Edu0liver/prototype-healthy-api/internal/platform/middleware"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes mounts /automations (admin only) within a tenant transaction.
func RegisterRoutes(e *gin.Engine, h *Handler, mw *middleware.Middleware) {
	g := e.Group("/automations", mw.Auth(), mw.Tenant(), mw.RBAC(middleware.RoleAdmin))
	g.POST("", h.Create)
	g.GET("", h.List)
	g.GET("/:id", h.Get)
	g.PUT("/:id", h.Update)
	g.DELETE("/:id", h.Delete)
}
