package http

import (
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/middleware"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes mounts /conversations/:id/handover/* (admin + operator).
func RegisterRoutes(e *gin.RouterGroup, h *Handler, mw *middleware.Middleware) {
	g := e.Group("/conversations/:id/handover",
		mw.Auth(), mw.Tenant(), mw.RequireActiveSubscription(), mw.RBAC(middleware.RoleAdmin, middleware.RoleOperator))
	g.POST("/take", h.Take)
	g.POST("/reply", h.Reply)
	g.POST("/return", h.Return)
	g.POST("/close", h.Close)
}
