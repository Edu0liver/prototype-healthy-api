// Package http exposes the channel module's Gin handlers.
package http

import (
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/channel/dtos"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/channel/infra/models"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/channel/services"
	"github.com/Edu0liver/prototype-healthy-api/internal/platform/httputil"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Handler serves channel endpoints.
type Handler struct {
	svc *services.Service
}

// NewHandler builds the handler.
func NewHandler(svc *services.Service) *Handler { return &Handler{svc: svc} }

// Create handles POST /channels.
func (h *Handler) Create(c *gin.Context) {
	var in dtos.CreateChannelRequest
	if !httputil.BindJSON(c, &in) {
		return
	}
	ch, err := h.svc.Create(c.Request.Context(), in)
	if err != nil {
		httputil.Fail(c, err)
		return
	}
	httputil.Created(c, channelResponse(ch))
}

// Connect handles POST /channels/:id/connect.
func (h *Handler) Connect(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var in dtos.ConnectRequest
	_ = c.ShouldBindJSON(&in)
	if in.Method == "" {
		in.Method = "qr"
	}
	res, err := h.svc.Connect(c.Request.Context(), id, in.Method, in.Number)
	if err != nil {
		httputil.Fail(c, err)
		return
	}
	httputil.OK(c, res)
}

// ConnectionState handles GET /channels/:id/connection-state.
func (h *Handler) ConnectionState(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	ch, err := h.svc.RefreshState(c.Request.Context(), id)
	if err != nil {
		httputil.Fail(c, err)
		return
	}
	httputil.OK(c, gin.H{"state": ch.Status})
}

// Get handles GET /channels/:id.
func (h *Handler) Get(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	ch, err := h.svc.Get(c.Request.Context(), id)
	if err != nil {
		httputil.Fail(c, err)
		return
	}
	httputil.OK(c, channelResponse(ch))
}

// List handles GET /channels.
func (h *Handler) List(c *gin.Context) {
	channels, err := h.svc.List(c.Request.Context())
	if err != nil {
		httputil.Fail(c, err)
		return
	}
	out := make([]dtos.ChannelResponse, len(channels))
	for i := range channels {
		out[i] = channelResponse(&channels[i])
	}
	httputil.OK(c, gin.H{"channels": out})
}

// Disconnect handles DELETE /channels/:id.
func (h *Handler) Disconnect(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	if err := h.svc.Disconnect(c.Request.Context(), id); err != nil {
		httputil.Fail(c, err)
		return
	}
	httputil.NoContent(c)
}

func parseID(c *gin.Context) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.Fail(c, httputil.BadRequest("invalid id"))
		return uuid.Nil, false
	}
	return id, true
}

func channelResponse(ch *models.Channel) dtos.ChannelResponse {
	var agentID *string
	if ch.ActiveAgentID != nil {
		s := ch.ActiveAgentID.String()
		agentID = &s
	}
	return dtos.ChannelResponse{
		ID: ch.ID.String(), Type: ch.Type, Name: ch.Name, Status: ch.Status,
		ExternalAccountID: ch.ExternalAccountID, InstanceName: ch.EvolutionInstanceName, ActiveAgentID: agentID,
	}
}
