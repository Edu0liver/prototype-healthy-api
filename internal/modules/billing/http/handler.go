// Package http exposes the billing module's Gin handlers (read-only tenant views
// of subscription and usage).
package http

import (
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/billing/service"
)

// Handler serves billing endpoints.
type Handler struct {
	svc *service.Service
}

// NewHandler builds the handler.
func NewHandler(svc *service.Service) *Handler { return &Handler{svc: svc} }
