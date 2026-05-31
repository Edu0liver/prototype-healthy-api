// Package http exposes the iam module's Gin handlers (one file per use case).
package http

import (
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/iam/dto"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/iam/infra/models"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/iam/service"
)

// Handler serves auth + user-management endpoints.
type Handler struct {
	svc *service.Service
}

// NewHandler builds the handler.
func NewHandler(svc *service.Service) *Handler { return &Handler{svc: svc} }

func userResponse(u *models.User) dto.UserResponse {
	return dto.UserResponse{
		ID: u.ID.String(), Email: u.Email, Name: u.Name, Role: u.Role.Name, Status: u.Status,
	}
}

func tokenResponse(t *service.Tokens, u *models.User) dto.TokenResponse {
	return dto.TokenResponse{
		AccessToken: t.Access, RefreshToken: t.Refresh, TokenType: "Bearer", User: userResponse(u),
	}
}
