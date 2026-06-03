package http_test

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	agenthttp "github.com/Edu0liver/prototype-healthy-api/internal/modules/agent/http"
	agentrepo "github.com/Edu0liver/prototype-healthy-api/internal/modules/agent/infra/repository"
	agentservice "github.com/Edu0liver/prototype-healthy-api/internal/modules/agent/service"
	billinghttp "github.com/Edu0liver/prototype-healthy-api/internal/modules/billing/http"
	billingrepo "github.com/Edu0liver/prototype-healthy-api/internal/modules/billing/infra/repository"
	billingservice "github.com/Edu0liver/prototype-healthy-api/internal/modules/billing/service"
	iamrepo "github.com/Edu0liver/prototype-healthy-api/internal/modules/iam/infra/repository"
	iamservice "github.com/Edu0liver/prototype-healthy-api/internal/modules/iam/service"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/config"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/database"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/middleware"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/testsupport"
	"github.com/Edu0liver/prototype-healthy-api/pkg/stripe"
	"github.com/Edu0liver/prototype-healthy-api/pkg/token"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type noopMailer struct{}

func (noopMailer) Send(string, string, string) error { return nil }

func seedCompany(t *testing.T, db *database.DB, slug string) uuid.UUID {
	t.Helper()
	id := uuid.New()
	err := db.System(context.Background(), func(ctx context.Context) error {
		return database.MustTx(ctx).Exec(
			"INSERT INTO companies (id, name, slug, status, plan) VALUES (?, ?, ?, 'active', 'free')",
			id, slug, slug,
		).Error
	})
	require.NoError(t, err)
	return id
}

// seedSubscription links a company to the seeded `free` plan.
func seedSubscription(t *testing.T, db *database.DB, companyID uuid.UUID) {
	t.Helper()
	err := db.System(context.Background(), func(ctx context.Context) error {
		return database.MustTx(ctx).Exec(
			"INSERT INTO subscriptions (id, company_id, plan_id, status) "+
				"SELECT gen_random_uuid(), ?, id, 'active' FROM plans WHERE code = 'free'",
			companyID,
		).Error
	})
	require.NoError(t, err)
}

// newRouter wires the billing + agent routers (agent with real billing quota
// enforcement) on one engine, scoped to a freshly seeded admin tenant.
func newRouter(t *testing.T, db *database.DB, withSubscription bool) (*gin.Engine, string) {
	t.Helper()
	cfg := &config.Config{}
	cfg.App.PublicBaseURL = "http://test"
	cfg.JWT.Secret = "test-secret-please-change"
	cfg.JWT.AccessTTL = 15 * time.Minute
	cfg.JWT.RefreshTTL = time.Hour
	tok := token.New(cfg.JWT.Secret, cfg.JWT.AccessTTL, cfg.JWT.RefreshTTL)

	slug := "bl-" + uuid.New().String()[:8]
	companyID := seedCompany(t, db, slug)
	if withSubscription {
		seedSubscription(t, db, companyID)
	}
	email := "admin+" + slug + "@billing.test"

	ctx := context.Background()
	iamSvc := iamservice.New(iamrepo.New(), db, tok, noopMailer{}, cfg, nil)
	_, err := iamSvc.RegisterFirstAdmin(ctx, companyID, email, "secret123", "Admin")
	require.NoError(t, err)
	tokens, _, err := iamSvc.Login(ctx, email, "secret123")
	require.NoError(t, err)

	mw := middleware.New(cfg, db, nil, tok, zap.NewNop())

	// Real billing service (no Redis: degraded counters, DB-backed limits; no
	// Stripe gateway in this integration scope).
	billSvc := billingservice.New(billingrepo.New(), db, nil, nil, zap.NewNop())

	gin.SetMode(gin.TestMode)
	r := gin.New()
	billinghttp.RegisterRoutes(r.Group(""), billinghttp.NewHandler(billSvc), mw)
	agentSvc := agentservice.New(agentrepo.New()).WithBilling(billSvc)
	agenthttp.RegisterRoutes(r.Group(""), agenthttp.NewHandler(agentSvc), mw)
	return r, tokens.Access
}

func do(t *testing.T, r *gin.Engine, method, path string, body any, bearer string) *httptest.ResponseRecorder {
	t.Helper()
	var rdr io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		require.NoError(t, err)
		rdr = bytes.NewReader(b)
	}
	req := httptest.NewRequest(method, path, rdr)
	req.Header.Set("Content-Type", "application/json")
	if bearer != "" {
		req.Header.Set("Authorization", "Bearer "+bearer)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestBillingRoutes(t *testing.T) {
	db := testsupport.NewPostgres(t)
	r, adminTok := newRouter(t, db, true)

	// GET /billing/subscription — unauthenticated → 401.
	w := do(t, r, http.MethodGet, "/billing/subscription", nil, "")
	require.Equal(t, http.StatusUnauthorized, w.Code, "sub-noauth: %s", w.Body.String())

	// GET /billing/subscription — happy path → free plan.
	w = do(t, r, http.MethodGet, "/billing/subscription", nil, adminTok)
	require.Equal(t, http.StatusOK, w.Code, "sub: %s", w.Body.String())
	var sub struct {
		PlanCode string `json:"plan_code"`
		Status   string `json:"status"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &sub))
	require.Equal(t, "free", sub.PlanCode)
	require.Equal(t, "active", sub.Status)

	// GET /billing/usage — returns the gated dimensions with quotas.
	w = do(t, r, http.MethodGet, "/billing/usage", nil, adminTok)
	require.Equal(t, http.StatusOK, w.Code, "usage: %s", w.Body.String())
	var usage struct {
		Items []struct {
			Kind  string `json:"kind"`
			Used  int64  `json:"used"`
			Quota int64  `json:"quota"`
		} `json:"items"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &usage))
	require.NotEmpty(t, usage.Items)
	kinds := map[string]int64{}
	for _, it := range usage.Items {
		kinds[it.Kind] = it.Quota
	}
	require.Equal(t, int64(100), kinds["ai_message"], "free plan ai_message quota")
}

// TestBillingQuotaHardStop verifies the free plan's max_agents=1 cap yields a
// 402 on the second agent create (EnsureResource end-to-end).
func TestBillingQuotaHardStop(t *testing.T) {
	db := testsupport.NewPostgres(t)
	r, adminTok := newRouter(t, db, true)

	body := gin.H{"name": "Bot", "system_prompt": "hi", "status": "active"}

	// First agent — within the cap.
	w := do(t, r, http.MethodPost, "/agents", body, adminTok)
	require.Equal(t, http.StatusCreated, w.Code, "agent1: %s", w.Body.String())

	// Second agent — over the free plan cap (max_agents=1) → 402.
	w = do(t, r, http.MethodPost, "/agents", body, adminTok)
	require.Equal(t, http.StatusPaymentRequired, w.Code, "agent2 should hit quota: %s", w.Body.String())
}

// TestBillingFailOpenNoSubscription ensures a company without a subscription is
// not blocked (fail-open) and its subscription read returns 404.
func TestBillingFailOpenNoSubscription(t *testing.T) {
	db := testsupport.NewPostgres(t)
	r, adminTok := newRouter(t, db, false)

	// No subscription → 404 on read.
	w := do(t, r, http.MethodGet, "/billing/subscription", nil, adminTok)
	require.Equal(t, http.StatusNotFound, w.Code, "sub-missing: %s", w.Body.String())

	// Create still works (quota fail-open).
	w = do(t, r, http.MethodPost, "/agents", gin.H{"name": "Bot", "system_prompt": "hi", "status": "active"}, adminTok)
	require.Equal(t, http.StatusCreated, w.Code, "create-failopen: %s", w.Body.String())
}

const testWebhookSecret = "whsec_integration_test"

func freePlanID(t *testing.T, db *database.DB) string {
	t.Helper()
	var id string
	err := db.System(context.Background(), func(ctx context.Context) error {
		return database.MustTx(ctx).Raw("SELECT id::text FROM plans WHERE code = 'free'").Scan(&id).Error
	})
	require.NoError(t, err)
	require.NotEmpty(t, id)
	return id
}

func signStripe(payload []byte) string {
	ts := fmt.Sprintf("%d", time.Now().Unix())
	mac := hmac.New(sha256.New, []byte(testWebhookSecret))
	mac.Write([]byte(ts + "." + string(payload)))
	return fmt.Sprintf("t=%s,v1=%s", ts, hex.EncodeToString(mac.Sum(nil)))
}

// TestStripeWebhookActivatesSubscription drives a signed checkout.session.completed
// event through the public webhook and asserts the subscription is provisioned.
func TestStripeWebhookActivatesSubscription(t *testing.T) {
	db := testsupport.NewPostgres(t)

	cfg := &config.Config{}
	cfg.JWT.Secret = "test-secret-please-change"
	cfg.JWT.AccessTTL = 15 * time.Minute
	cfg.JWT.RefreshTTL = time.Hour
	tok := token.New(cfg.JWT.Secret, cfg.JWT.AccessTTL, cfg.JWT.RefreshTTL)

	slug := "sw-" + uuid.New().String()[:8]
	companyID := seedCompany(t, db, slug)
	email := "admin+" + slug + "@billing.test"
	ctx := context.Background()
	iamSvc := iamservice.New(iamrepo.New(), db, tok, noopMailer{}, cfg, nil)
	_, err := iamSvc.RegisterFirstAdmin(ctx, companyID, email, "secret123", "Admin")
	require.NoError(t, err)
	tokens, _, err := iamSvc.Login(ctx, email, "secret123")
	require.NoError(t, err)

	mw := middleware.New(cfg, db, nil, tok, zap.NewNop())
	stripeClient := stripe.New(stripe.Config{WebhookSecret: testWebhookSecret})
	billSvc := billingservice.New(billingrepo.New(), db, nil, stripeClient, zap.NewNop())

	gin.SetMode(gin.TestMode)
	r := gin.New()
	billinghttp.RegisterRoutes(r.Group(""), billinghttp.NewHandler(billSvc), mw)

	// No subscription yet → 404.
	w := do(t, r, http.MethodGet, "/billing/subscription", nil, tokens.Access)
	require.Equal(t, http.StatusNotFound, w.Code, "pre: %s", w.Body.String())

	// Signed checkout.session.completed → activates the free plan.
	event := map[string]any{
		"id":   "evt_" + uuid.New().String(),
		"type": "checkout.session.completed",
		"data": map[string]any{"object": map[string]any{
			"client_reference_id": companyID.String(),
			"customer":            "cus_test",
			"subscription":        "sub_test",
			"metadata": map[string]string{
				"company_id": companyID.String(),
				"plan_id":    freePlanID(t, db),
			},
		}},
	}
	payload, _ := json.Marshal(event)
	req := httptest.NewRequest(http.MethodPost, "/webhooks/stripe", bytes.NewReader(payload))
	req.Header.Set("Stripe-Signature", signStripe(payload))
	wh := httptest.NewRecorder()
	r.ServeHTTP(wh, req)
	require.Equal(t, http.StatusOK, wh.Code, "webhook: %s", wh.Body.String())

	// Subscription now readable and active.
	w = do(t, r, http.MethodGet, "/billing/subscription", nil, tokens.Access)
	require.Equal(t, http.StatusOK, w.Code, "post: %s", w.Body.String())
	var sub struct {
		PlanCode string `json:"plan_code"`
		Status   string `json:"status"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &sub))
	require.Equal(t, "free", sub.PlanCode)
	require.Equal(t, "active", sub.Status)

	// Bad signature → 400.
	req = httptest.NewRequest(http.MethodPost, "/webhooks/stripe", bytes.NewReader(payload))
	req.Header.Set("Stripe-Signature", "t=1,v1=deadbeef")
	wh = httptest.NewRecorder()
	r.ServeHTTP(wh, req)
	require.Equal(t, http.StatusBadRequest, wh.Code, "bad-sig: %s", wh.Body.String())
}
