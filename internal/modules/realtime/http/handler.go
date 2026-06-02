// Package http exposes the realtime WebSocket endpoint. It authenticates via a
// bearer token (sent in the Sec-WebSocket-Protocol header, falling back to a
// query param for older clients), resolves the tenant, and bridges the
// company's Redis Pub/Sub channel to the socket (RF-LOG-01).
package http

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Edu0liver/prototype-healthy-api/internal/shared/config"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/events"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/redisx"
	"github.com/Edu0liver/prototype-healthy-api/pkg/httputil"
	"github.com/Edu0liver/prototype-healthy-api/pkg/token"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

const (
	pingInterval = 30 * time.Second
	pongWait     = 60 * time.Second
	// bearerSubprotocol is the WebSocket subprotocol used to carry the JWT, so the
	// token travels in a header rather than the URL (which leaks into access logs).
	bearerSubprotocol = "bearer"
)

// Handler serves the realtime WebSocket.
type Handler struct {
	tokens   *token.Manager
	rdb      *redisx.Client
	log      *zap.Logger
	upgrader websocket.Upgrader
}

// NewHandler builds the handler with an Origin-checked upgrader.
func NewHandler(tokens *token.Manager, rdb *redisx.Client, cfg *config.Config, log *zap.Logger) *Handler {
	allowed := cfg.Security.AllowedOrigins
	return &Handler{
		tokens: tokens,
		rdb:    rdb,
		log:    log,
		upgrader: websocket.Upgrader{
			Subprotocols: []string{bearerSubprotocol},
			CheckOrigin:  originChecker(allowed),
		},
	}
}

// originChecker permits requests with no Origin (non-browser clients, which carry
// no ambient credentials), same-origin requests (Origin host == request Host),
// and any Origin in the configured allowlist. Everything else is rejected to
// prevent cross-site WebSocket hijacking.
func originChecker(allowed []string) func(*http.Request) bool {
	return func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		if origin == "" {
			return true
		}
		u, err := url.Parse(origin)
		if err != nil {
			return false
		}
		if strings.EqualFold(u.Host, r.Host) {
			return true
		}
		for _, a := range allowed {
			if strings.EqualFold(a, origin) || strings.EqualFold(a, u.Host) {
				return true
			}
		}
		return false
	}
}

// RegisterRoutes mounts GET /ws.
func RegisterRoutes(e *gin.RouterGroup, h *Handler) {
	e.GET("/ws", h.WS)
}

// bearerToken extracts the JWT from the Sec-WebSocket-Protocol header
// ("bearer, <token>") when present, otherwise from the ?token= query param.
func bearerToken(c *gin.Context) string {
	if proto := c.GetHeader("Sec-WebSocket-Protocol"); proto != "" {
		parts := strings.Split(proto, ",")
		if len(parts) == 2 && strings.EqualFold(strings.TrimSpace(parts[0]), bearerSubprotocol) {
			return strings.TrimSpace(parts[1])
		}
	}
	return c.Query("token")
}

// WS upgrades the connection and streams the tenant's realtime events.
// @Summary  Realtime WebSocket
// @Tags     realtime
// @Param    token query string false "Access token (JWT). Prefer the Sec-WebSocket-Protocol header ('bearer, <token>')."
// @Success  101
// @Failure  401 {object} httputil.ErrorResponse "invalid token"
// @Failure  403 {object} httputil.ErrorResponse "origin not allowed"
// @Failure  500 {object} httputil.ErrorResponse "internal error"
// @Router   /ws [get]
func (h *Handler) WS(c *gin.Context) {
	claims, err := h.tokens.Parse(bearerToken(c))
	if err != nil || claims.Type != token.TypeAccess {
		c.JSON(http.StatusUnauthorized, httputil.ErrorResponse{Error: "unauthorized", Message: "invalid token"})
		return
	}
	companyID, err := uuid.Parse(claims.CompanyID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, httputil.ErrorResponse{Error: "unauthorized", Message: "bad claims"})
		return
	}

	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return // upgrade writes its own error (incl. 403 on Origin rejection)
	}
	defer conn.Close()

	conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(string) error {
		return conn.SetReadDeadline(time.Now().Add(pongWait))
	})

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

	ticker := time.NewTicker(pingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
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
