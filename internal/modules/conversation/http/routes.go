package http

import (
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/middleware"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes mounts /conversations (admin + operator) within a tenant tx.
func RegisterRoutes(e *gin.Engine, h *Handler, mw *middleware.Middleware) {
	g := e.Group("/conversations", mw.Auth(), mw.Tenant(), mw.RBAC(middleware.RoleAdmin, middleware.RoleOperator))
	g.GET("", h.List)
	g.GET("/:id", h.Get)
	g.GET("/:id/messages", h.Messages)
}
