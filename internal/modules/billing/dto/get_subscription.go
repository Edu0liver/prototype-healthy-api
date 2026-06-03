// Package dto holds the billing module's request/response shapes.
package dto

// SubscriptionResponse is the tenant-facing view of its subscription + plan.
type SubscriptionResponse struct {
	PlanCode           string `json:"plan_code"`
	PlanName           string `json:"plan_name"`
	Status             string `json:"status"`
	BillingCycle       string `json:"billing_cycle"`
	CurrentPeriodStart string `json:"current_period_start"`
	CurrentPeriodEnd   string `json:"current_period_end"`
	CancelAtPeriodEnd  bool   `json:"cancel_at_period_end"`
	PriceCents         int    `json:"price_cents"`
	Currency           string `json:"currency"`
}
