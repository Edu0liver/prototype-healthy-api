// Package http exposes the handover operator endpoints.
package http

import (
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/handover/services"
	"github.com/Edu0liver/prototype-healthy-api/internal/platform/appctx"
	"github.com/Edu0liver/prototype-healthy-api/internal/platform/httputil"
	"github.com/Edu0liver/prototype-healthy-api/internal/platform/middleware"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ReplyRequest carries the operator's message.
type ReplyRequest struct {
	Content string `json:"content" binding:"required"`
}

// Handler serves handover endpoints.
type Handler struct {
	svc *services.Service
}

// NewHandler builds the handler.
func NewHandler(svc *services.Service) *Handler { return &Handler{svc: svc} }

// RegisterRoutes mounts /conversations/:id/handover/* (admin + operator).
func RegisterRoutes(e *gin.Engine, h *Handler, mw *middleware.Middleware) {
	g := e.Group("/conversations/:id/handover",
		mw.Auth(), mw.Tenant(), mw.RBAC(middleware.RoleAdmin, middleware.RoleOperator))
	g.POST("/take", h.Take)
	g.POST("/reply", h.Reply)
	g.POST("/return", h.Return)
	g.POST("/close", h.Close)
}

// Take handles POST /conversations/:id/handover/take.
func (h *Handler) Take(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	if err := h.svc.Take(c.Request.Context(), id, appctx.UserID(c.Request.Context())); err != nil {
		httputil.Fail(c, err)
		return
	}
	httputil.OK(c, gin.H{"state": "human"})
}

// Reply handles POST /conversations/:id/handover/reply.
func (h *Handler) Reply(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var in ReplyRequest
	if !httputil.BindJSON(c, &in) {
		return
	}
	if err := h.svc.Reply(c.Request.Context(), id, in.Content); err != nil {
		httputil.Fail(c, err)
		return
	}
	httputil.NoContent(c)
}

// Return handles POST /conversations/:id/handover/return.
func (h *Handler) Return(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	if err := h.svc.Return(c.Request.Context(), id); err != nil {
		httputil.Fail(c, err)
		return
	}
	httputil.OK(c, gin.H{"state": "ai"})
}

// Close handles POST /conversations/:id/handover/close.
func (h *Handler) Close(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	if err := h.svc.Close(c.Request.Context(), id); err != nil {
		httputil.Fail(c, err)
		return
	}
	httputil.OK(c, gin.H{"state": "closed"})
}

func parseID(c *gin.Context) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.Fail(c, httputil.BadRequest("invalid id"))
		return uuid.Nil, false
	}
	return id, true
}
