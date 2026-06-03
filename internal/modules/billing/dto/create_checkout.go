package dto

// CreateCheckoutRequest starts a subscription checkout for a plan.
type CreateCheckoutRequest struct {
	PlanCode string `json:"plan_code" binding:"required"`
}

// CheckoutResponse carries the hosted Stripe Checkout URL to redirect the user.
type CheckoutResponse struct {
	CheckoutURL string `json:"checkout_url"`
}
