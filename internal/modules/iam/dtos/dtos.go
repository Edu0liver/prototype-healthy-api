// Package dtos holds request/response payloads for the iam module.
package dtos

// LoginRequest authenticates a user inside a company.
type LoginRequest struct {
	CompanySlug string `json:"company_slug" binding:"required"`
	Email       string `json:"email" binding:"required,email"`
	Password    string `json:"password" binding:"required"`
}

// RegisterAdminRequest bootstraps the first admin for a company.
type RegisterAdminRequest struct {
	CompanySlug string `json:"company_slug" binding:"required"`
	Email       string `json:"email" binding:"required,email"`
	Password    string `json:"password" binding:"required,min=8"`
	Name        string `json:"name" binding:"omitempty"`
}

// RefreshRequest exchanges a refresh token for a new access token.
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// InviteRequest invites a new user with a role.
type InviteRequest struct {
	Email string `json:"email" binding:"required,email"`
	Name  string `json:"name" binding:"omitempty"`
	Role  string `json:"role" binding:"required,oneof=admin operator knowledge_manager"`
}

// AcceptInviteRequest sets the password for an invited user.
type AcceptInviteRequest struct {
	Token    string `json:"token" binding:"required"`
	Password string `json:"password" binding:"required,min=8"`
}

// UserResponse describes a user.
type UserResponse struct {
	ID     string `json:"id"`
	Email  string `json:"email"`
	Name   string `json:"name"`
	Role   string `json:"role"`
	Status string `json:"status"`
}

// TokenResponse is the auth result.
type TokenResponse struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	TokenType    string       `json:"token_type"`
	User         UserResponse `json:"user"`
}
