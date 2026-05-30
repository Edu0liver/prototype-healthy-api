package http

import (
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/middleware"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes mounts /channels (admin only) within a tenant transaction.
func RegisterRoutes(e *gin.Engine, h *Handler, mw *middleware.Middleware) {
	g := e.Group("/channels", mw.Auth(), mw.Tenant(), mw.RBAC(middleware.RoleAdmin))
	g.POST("", h.Create)
	g.GET("", h.List)
	g.GET("/:id", h.Get)
	g.POST("/:id/connect", h.Connect)
	g.GET("/:id/connection-state", h.ConnectionState)
	g.DELETE("/:id", h.Disconnect)
}
