package dto

// RegisterAdminRequest bootstraps the first admin for a company.
type RegisterAdminRequest struct {
	CompanyID string `json:"company_id" binding:"required,uuid"`
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=8"`
	Name      string `json:"name" binding:"omitempty"`
}
