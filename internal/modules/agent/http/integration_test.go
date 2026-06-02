package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	agenthttp "github.com/Edu0liver/prototype-healthy-api/internal/modules/agent/http"
	agentrepo "github.com/Edu0liver/prototype-healthy-api/internal/modules/agent/infra/repository"
	agentservice "github.com/Edu0liver/prototype-healthy-api/internal/modules/agent/service"
	iamrepo "github.com/Edu0liver/prototype-healthy-api/internal/modules/iam/infra/repository"
	iamservice "github.com/Edu0liver/prototype-healthy-api/internal/modules/iam/service"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/config"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/database"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/middleware"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/redisx"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/testsupport"
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

// newRouter spins up the agent Gin router with real middleware and returns an
// admin access token scoped to a freshly seeded tenant.
func newRouter(t *testing.T, db *database.DB) (*gin.Engine, string) {
	t.Helper()
	cfg := &config.Config{}
	cfg.App.PublicBaseURL = "http://test"
	cfg.JWT.Secret = "test-secret-please-change"
	cfg.JWT.AccessTTL = 15 * time.Minute
	cfg.JWT.RefreshTTL = time.Hour

	tok := token.New(cfg.JWT.Secret, cfg.JWT.AccessTTL, cfg.JWT.RefreshTTL)

	slug := "ag-" + uuid.New().String()[:8]
	companyID := seedCompany(t, db, slug)
	email := "admin+" + slug + "@agent.test"

	ctx := context.Background()
	iamSvc := iamservice.New(iamrepo.New(), db, tok, noopMailer{}, cfg, nil)
	_, err := iamSvc.RegisterFirstAdmin(ctx, companyID, email, "secret123", "Admin")
	require.NoError(t, err)
	tokens, _, err := iamSvc.Login(ctx, email, "secret123")
	require.NoError(t, err)

	var rdb *redisx.Client
	mw := middleware.New(cfg, db, rdb, tok, zap.NewNop())

	svc := agentservice.New(agentrepo.New())
	h := agenthttp.NewHandler(svc)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	agenthttp.RegisterRoutes(r.Group(""), h, mw)
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

// TestAgentRoutes exercises every /agents endpoint through the real Gin router,
// auth middleware, tenant middleware, and RBAC middleware.
func TestAgentRoutes(t *testing.T) {
	db := testsupport.NewPostgres(t)
	r, adminTok := newRouter(t, db)

	// 1. POST /agents — unauthenticated must be rejected.
	w := do(t, r, http.MethodPost, "/agents", gin.H{
		"name": "Bot", "system_prompt": "hi",
	}, "")
	require.Equal(t, http.StatusUnauthorized, w.Code, "create-noauth: %s", w.Body.String())

	// 2. POST /agents — missing required field system_prompt must be 400.
	w = do(t, r, http.MethodPost, "/agents", gin.H{"name": "Bot"}, adminTok)
	require.Equal(t, http.StatusBadRequest, w.Code, "create-missing-prompt: %s", w.Body.String())

	// 3. POST /agents — name too short (min=2) must be 400.
	w = do(t, r, http.MethodPost, "/agents", gin.H{
		"name": "X", "system_prompt": "hi",
	}, adminTok)
	require.Equal(t, http.StatusBadRequest, w.Code, "create-short-name: %s", w.Body.String())

	// 4. POST /agents — happy path.
	w = do(t, r, http.MethodPost, "/agents", gin.H{
		"name":          "Support Bot",
		"system_prompt": "You are a customer support agent.",
		"model":         "gpt-4o",
		"status":        "active",
	}, adminTok)
	require.Equal(t, http.StatusCreated, w.Code, "create: %s", w.Body.String())
	var created struct {
		ID           string `json:"id"`
		Name         string `json:"name"`
		Status       string `json:"status"`
		Model        string `json:"model"`
		SystemPrompt string `json:"system_prompt"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &created))
	require.NotEmpty(t, created.ID)
	require.Equal(t, "Support Bot", created.Name)
	require.Equal(t, "active", created.Status)
	require.Equal(t, "gpt-4o", created.Model)

	// 5. GET /agents — list contains created agent.
	w = do(t, r, http.MethodGet, "/agents", nil, adminTok)
	require.Equal(t, http.StatusOK, w.Code, "list: %s", w.Body.String())
	var list struct {
		Agents []struct {
			ID string `json:"id"`
		} `json:"agents"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &list))
	require.Len(t, list.Agents, 1)
	require.Equal(t, created.ID, list.Agents[0].ID)

	// 6. GET /agents — unauthenticated.
	w = do(t, r, http.MethodGet, "/agents", nil, "")
	require.Equal(t, http.StatusUnauthorized, w.Code, "list-noauth: %s", w.Body.String())

	// 7. GET /agents/:id — happy path.
	w = do(t, r, http.MethodGet, "/agents/"+created.ID, nil, adminTok)
	require.Equal(t, http.StatusOK, w.Code, "get: %s", w.Body.String())
	require.Contains(t, w.Body.String(), created.ID)
	require.Contains(t, w.Body.String(), "Support Bot")

	// 8. GET /agents/:id — invalid UUID.
	w = do(t, r, http.MethodGet, "/agents/not-a-uuid", nil, adminTok)
	require.Equal(t, http.StatusBadRequest, w.Code, "get-bad-id: %s", w.Body.String())

	// 9. GET /agents/:id — not found.
	w = do(t, r, http.MethodGet, "/agents/"+uuid.New().String(), nil, adminTok)
	require.Equal(t, http.StatusNotFound, w.Code, "get-notfound: %s", w.Body.String())

	// 10. PUT /agents/:id — update name and status.
	w = do(t, r, http.MethodPut, "/agents/"+created.ID, gin.H{
		"name":   "Updated Bot",
		"status": "draft",
	}, adminTok)
	require.Equal(t, http.StatusOK, w.Code, "update: %s", w.Body.String())
	require.Contains(t, w.Body.String(), "Updated Bot")
	require.Contains(t, w.Body.String(), "draft")

	// 11. PUT /agents/:id — invalid status value must be 400.
	w = do(t, r, http.MethodPut, "/agents/"+created.ID, gin.H{
		"status": "invalid-status",
	}, adminTok)
	require.Equal(t, http.StatusBadRequest, w.Code, "update-bad-status: %s", w.Body.String())

	// 12. PUT /agents/:id — not found.
	w = do(t, r, http.MethodPut, "/agents/"+uuid.New().String(), gin.H{"name": "Ghost"}, adminTok)
	require.Equal(t, http.StatusNotFound, w.Code, "update-notfound: %s", w.Body.String())

	// 13. DELETE /agents/:id — happy path.
	w = do(t, r, http.MethodDelete, "/agents/"+created.ID, nil, adminTok)
	require.Equal(t, http.StatusNoContent, w.Code, "delete: %s", w.Body.String())

	// 14. GET /agents — list is empty after delete.
	w = do(t, r, http.MethodGet, "/agents", nil, adminTok)
	require.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &list))
	require.Empty(t, list.Agents, "list-after-delete: %s", w.Body.String())

	// 15. DELETE /agents/:id — already deleted returns 404.
	w = do(t, r, http.MethodDelete, "/agents/"+created.ID, nil, adminTok)
	require.Equal(t, http.StatusNotFound, w.Code, "delete-notfound: %s", w.Body.String())
}

// TestAgentTenantIsolation ensures agents of company A are invisible to company B.
func TestAgentTenantIsolation(t *testing.T) {
	db := testsupport.NewPostgres(t)

	rA, tokA := newRouter(t, db)
	_, tokB := newRouter(t, db)

	// Create an agent for tenant A.
	w := do(t, rA, http.MethodPost, "/agents", gin.H{
		"name": "A-Bot", "system_prompt": "hello", "status": "active",
	}, tokA)
	require.Equal(t, http.StatusCreated, w.Code)
	var a struct {
		ID string `json:"id"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &a))

	// Tenant B lists agents — must not include A's agent.
	w = do(t, rA, http.MethodGet, "/agents", nil, tokB)
	require.Equal(t, http.StatusOK, w.Code)
	var listB struct {
		Agents []struct {
			ID string `json:"id"`
		} `json:"agents"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &listB))
	for _, ag := range listB.Agents {
		require.NotEqual(t, a.ID, ag.ID, "tenant isolation breach: B sees A's agent")
	}

	// Tenant B cannot fetch A's agent by ID directly.
	w = do(t, rA, http.MethodGet, "/agents/"+a.ID, nil, tokB)
	require.Equal(t, http.StatusNotFound, w.Code, "B must not fetch A's agent: %s", w.Body.String())
}
