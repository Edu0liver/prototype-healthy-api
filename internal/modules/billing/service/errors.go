package service

import (
	"net/http"

	"github.com/Edu0liver/prototype-healthy-api/pkg/httputil"
)

// Errors. ErrQuotaExceeded maps to HTTP 402 Payment Required and is returned by
// the hard-limit checks used in the create_* flows.
var (
	ErrQuotaExceeded     = httputil.PaymentRequired("plan limit reached for this resource")
	ErrNoSubscription    = httputil.PaymentRequired("no active subscription for this company")
	ErrSubscriptionNotFn = httputil.NotFound("subscription not found")

	// Stripe gateway errors.
	ErrStripeDisabled     = httputil.NewDomainError(http.StatusServiceUnavailable, "stripe_disabled", "billing gateway is not configured")
	ErrPlanNotFound       = httputil.NotFound("plan not found")
	ErrPlanNotPurchasable = httputil.NewDomainError(http.StatusConflict, "plan_not_purchasable", "plan has no Stripe price configured")
	ErrInvalidSignature   = httputil.BadRequest("invalid stripe signature")
)
