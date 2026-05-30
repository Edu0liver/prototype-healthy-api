package http

import (
	"github.com/Edu0liver/prototype-healthy-api/internal/platform/middleware"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes mounts knowledge endpoints. Admins and knowledge managers may
// manage RAG (RF-USR-02); all run within a tenant transaction.
func RegisterRoutes(e *gin.Engine, h *Handler, mw *middleware.Middleware) {
	rag := []gin.HandlerFunc{mw.Auth(), mw.Tenant(), mw.RBAC(middleware.RoleAdmin, middleware.RoleKnowledgeManager)}

	kb := e.Group("/knowledge-bases", rag...)
	kb.POST("", h.CreateKB)
	kb.GET("", h.ListKB)
	kb.GET("/:id", h.GetKB)
	kb.DELETE("/:id", h.DeleteKB)
	kb.POST("/:id/documents", h.UploadFile)
	kb.POST("/:id/documents/text", h.UploadText)
	kb.GET("/:id/documents", h.ListDocuments)

	docs := e.Group("/documents", rag...)
	docs.DELETE("/:docId", h.DeleteDocument)

	// NOTE: param must be ":id" to match the agent module's /agents/:id routes
	// (Gin panics on differing wildcard names at the same path position).
	link := e.Group("/agents/:id/knowledge-bases", rag...)
	link.POST("/:kbId", h.LinkAgent)
	link.DELETE("/:kbId", h.UnlinkAgent)
}
