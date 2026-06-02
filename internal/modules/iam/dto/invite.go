package dto

// InviteRequest invites a new user with a role.
type InviteRequest struct {
	Email string `json:"email" binding:"required,email"`
	Name  string `json:"name" binding:"omitempty"`
	Role  string `json:"role" binding:"required,oneof=admin operator knowledge_manager"`
}

// AcceptInviteRequest sets the password for an invited user.
type AcceptInviteRequest struct {
	Token    string `json:"token" binding:"required"`
	Password string `json:"password" binding:"required,min=8,max=128"`
}
