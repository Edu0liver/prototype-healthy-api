package http

import (
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/middleware"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes mounts the iam routes under /auth (public) and /users (admin).
func RegisterRoutes(e *gin.RouterGroup, h *Handler, mw *middleware.Middleware) {
	auth := e.Group("/auth")
	auth.POST("/register", h.Register)
	auth.POST("/login", h.Login)
	auth.POST("/refresh", h.Refresh)
	auth.POST("/accept-invite", h.AcceptInvite)
	auth.GET("/me", mw.Auth(), mw.Tenant(), h.Me)

	users := e.Group("/users", mw.Auth(), mw.Tenant(), mw.RBAC(middleware.RoleAdmin))
	users.POST("", h.Invite)
	users.GET("", h.ListUsers)
}
