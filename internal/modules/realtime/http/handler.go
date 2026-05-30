// Package http exposes the realtime WebSocket endpoint. It authenticates via a
// token query param (browsers cannot set headers on WS), resolves the tenant,
// and bridges the company's Redis Pub/Sub channel to the socket (RF-LOG-01).
package http

import (
	"context"
	"net/http"

	"github.com/Edu0liver/prototype-healthy-api/internal/platform/events"
	"github.com/Edu0liver/prototype-healthy-api/internal/platform/redisx"
	"github.com/Edu0liver/prototype-healthy-api/internal/platform/token"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(*http.Request) bool { return true }, // tighten in production
}

// Handler serves the realtime WebSocket.
type Handler struct {
	tokens *token.Manager
	rdb    *redisx.Client
	log    *zap.Logger
}

// NewHandler builds the handler.
func NewHandler(tokens *token.Manager, rdb *redisx.Client, log *zap.Logger) *Handler {
	return &Handler{tokens: tokens, rdb: rdb, log: log}
}

// RegisterRoutes mounts GET /ws.
func RegisterRoutes(e *gin.Engine, h *Handler) {
	e.GET("/ws", h.WS)
}

// WS upgrades the connection and streams the tenant's realtime events.
func (h *Handler) WS(c *gin.Context) {
	claims, err := h.tokens.Parse(c.Query("token"))
	if err != nil || claims.Type != token.TypeAccess {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		return
	}
	companyID, err := uuid.Parse(claims.CompanyID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "bad claims"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return // upgrade writes its own error
	}
	defer conn.Close()

	ctx, cancel := context.WithCancel(c.Request.Context())
	defer cancel()

	pubsub := h.rdb.Subscribe(ctx, events.Channel(companyID))
	defer pubsub.Close()
	ch := pubsub.Channel()

	// Detect client disconnect via the read pump.
	go func() {
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				cancel()
				return
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-ch:
			if !ok {
				return
			}
			if err := conn.WriteMessage(websocket.TextMessage, []byte(msg.Payload)); err != nil {
				return
			}
		}
	}
}
