package dto

// UpdateBrandingRequest updates white-label settings.
type UpdateBrandingRequest struct {
	LogoURL         string `json:"logo_url" binding:"omitempty,url"`
	FaviconURL      string `json:"favicon_url" binding:"omitempty,url"`
	PrimaryColor    string `json:"primary_color" binding:"omitempty,hexcolor"`
	SecondaryColor  string `json:"secondary_color" binding:"omitempty,hexcolor"`
	EmailSenderName string `json:"email_sender_name" binding:"omitempty"`
}

// BrandingResponse is the white-label theme served to the frontend.
type BrandingResponse struct {
	CompanyID       string `json:"company_id"`
	LogoURL         string `json:"logo_url"`
	FaviconURL      string `json:"favicon_url"`
	PrimaryColor    string `json:"primary_color"`
	SecondaryColor  string `json:"secondary_color"`
	EmailSenderName string `json:"email_sender_name"`
}
