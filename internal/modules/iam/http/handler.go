// Package http exposes the iam module's Gin handlers.
package http

import (
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/iam/dtos"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/iam/infra/models"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/iam/services"
	"github.com/Edu0liver/prototype-healthy-api/internal/platform/appctx"
	"github.com/Edu0liver/prototype-healthy-api/internal/platform/httputil"
	"github.com/gin-gonic/gin"
)

// Handler serves auth + user-management endpoints.
type Handler struct {
	svc *services.Service
}

// NewHandler builds the handler.
func NewHandler(svc *services.Service) *Handler { return &Handler{svc: svc} }

// Login authenticates and returns tokens (public).
func (h *Handler) Login(c *gin.Context) {
	var in dtos.LoginRequest
	if !httputil.BindJSON(c, &in) {
		return
	}
	tokens, user, err := h.svc.Login(c.Request.Context(), in.CompanySlug, in.Email, in.Password)
	if err != nil {
		httputil.Fail(c, err)
		return
	}
	httputil.OK(c, tokenResponse(tokens, user))
}

// Register bootstraps the first admin for a company (public, first-user only).
func (h *Handler) Register(c *gin.Context) {
	var in dtos.RegisterAdminRequest
	if !httputil.BindJSON(c, &in) {
		return
	}
	user, err := h.svc.RegisterFirstAdmin(c.Request.Context(), in.CompanySlug, in.Email, in.Password, in.Name)
	if err != nil {
		httputil.Fail(c, err)
		return
	}
	httputil.Created(c, userResponse(user))
}

// Refresh issues a new token pair from a refresh token (public).
func (h *Handler) Refresh(c *gin.Context) {
	var in dtos.RefreshRequest
	if !httputil.BindJSON(c, &in) {
		return
	}
	tokens, err := h.svc.Refresh(c.Request.Context(), in.RefreshToken)
	if err != nil {
		httputil.Fail(c, err)
		return
	}
	httputil.OK(c, dtos.TokenResponse{
		AccessToken: tokens.Access, RefreshToken: tokens.Refresh, TokenType: "Bearer",
	})
}

// AcceptInvite sets the password for an invited user (public).
func (h *Handler) AcceptInvite(c *gin.Context) {
	var in dtos.AcceptInviteRequest
	if !httputil.BindJSON(c, &in) {
		return
	}
	if err := h.svc.AcceptInvite(c.Request.Context(), in.Token, in.Password); err != nil {
		httputil.Fail(c, err)
		return
	}
	httputil.NoContent(c)
}

// Invite creates an invited user (admin).
func (h *Handler) Invite(c *gin.Context) {
	var in dtos.InviteRequest
	if !httputil.BindJSON(c, &in) {
		return
	}
	user, err := h.svc.Invite(c.Request.Context(), in.Email, in.Name, in.Role)
	if err != nil {
		httputil.Fail(c, err)
		return
	}
	httputil.Created(c, userResponse(user))
}

// Me returns the authenticated user.
func (h *Handler) Me(c *gin.Context) {
	user, err := h.svc.Me(c.Request.Context(), appctx.UserID(c.Request.Context()))
	if err != nil {
		httputil.Fail(c, err)
		return
	}
	httputil.OK(c, userResponse(user))
}

// ListUsers lists the tenant's users (admin).
func (h *Handler) ListUsers(c *gin.Context) {
	users, err := h.svc.ListUsers(c.Request.Context())
	if err != nil {
		httputil.Fail(c, err)
		return
	}
	out := make([]dtos.UserResponse, len(users))
	for i := range users {
		out[i] = userResponse(&users[i])
	}
	httputil.OK(c, gin.H{"users": out})
}

// ---- mappers --------------------------------------------------------------

func userResponse(u *models.User) dtos.UserResponse {
	return dtos.UserResponse{
		ID: u.ID.String(), Email: u.Email, Name: u.Name, Role: u.Role, Status: u.Status,
	}
}

func tokenResponse(t *services.Tokens, u *models.User) dtos.TokenResponse {
	return dtos.TokenResponse{
		AccessToken: t.Access, RefreshToken: t.Refresh, TokenType: "Bearer", User: userResponse(u),
	}
}
