package dto

// PortalResponse carries the Stripe Billing Portal URL to redirect the user.
type PortalResponse struct {
	PortalURL string `json:"portal_url"`
}
