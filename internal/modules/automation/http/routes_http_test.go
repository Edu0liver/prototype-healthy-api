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

	agentrepo "github.com/Edu0liver/prototype-healthy-api/internal/modules/agent/infra/repository"
	agentservice "github.com/Edu0liver/prototype-healthy-api/internal/modules/agent/service"
	agentdto "github.com/Edu0liver/prototype-healthy-api/internal/modules/agent/dto"
	automationhttp "github.com/Edu0liver/prototype-healthy-api/internal/modules/automation/http"
	automationrepo "github.com/Edu0liver/prototype-healthy-api/internal/modules/automation/infra/repository"
	automationservice "github.com/Edu0liver/prototype-healthy-api/internal/modules/automation/service"
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

	chanmodels "github.com/Edu0liver/prototype-healthy-api/internal/modules/channel/infra/models"
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

// newRouter returns (router, adminToken, companyID).
func newRouter(t *testing.T, db *database.DB) (*gin.Engine, string, uuid.UUID) {
	t.Helper()
	cfg := &config.Config{}
	cfg.App.PublicBaseURL = "http://test"
	cfg.JWT.Secret = "test-secret-please-change"
	cfg.JWT.AccessTTL = 15 * time.Minute
	cfg.JWT.RefreshTTL = time.Hour

	tok := token.New(cfg.JWT.Secret, cfg.JWT.AccessTTL, cfg.JWT.RefreshTTL)
	slug := "at-" + uuid.New().String()[:8]
	companyID := seedCompany(t, db, slug)

	ctx := context.Background()
	iamSvc := iamservice.New(iamrepo.New(), db, tok, noopMailer{}, cfg)
	_, err := iamSvc.RegisterFirstAdmin(ctx, slug, "admin@auto.test", "secret123", "Admin")
	require.NoError(t, err)
	tokens, _, err := iamSvc.Login(ctx, slug, "admin@auto.test", "secret123")
	require.NoError(t, err)

	var rdb *redisx.Client
	mw := middleware.New(cfg, db, rdb, tok, zap.NewNop())

	svc := automationservice.New(automationrepo.New(), db)
	h := automationhttp.NewHandler(svc)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	automationhttp.RegisterRoutes(r.Group(""), h, mw)
	return r, tokens.Access, companyID
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

// seedChannelAndAgent inserts a channel and an agent for the given company.
func seedChannelAndAgent(t *testing.T, db *database.DB, companyID uuid.UUID) (channelID, agentID uuid.UUID) {
	t.Helper()
	ctx := context.Background()

	// Insert channel directly (instagram — no Evolution API call needed).
	chID, _ := uuid.NewV7()
	err := db.System(ctx, func(ctx context.Context) error {
		ch := &chanmodels.Channel{
			ID:        chID,
			CompanyID: companyID,
			Type:      "instagram",
			Name:      "Test IG",
			Status:    "disconnected",
			Metadata:  map[string]any{},
		}
		return database.MustTx(ctx).Create(ch).Error
	})
	require.NoError(t, err)

	// Insert agent via agent service.
	agSvc := agentservice.New(agentrepo.New())
	var agID uuid.UUID
	err = db.Tenant(ctx, companyID, func(tctx context.Context) error {
		ag, err := agSvc.Create(tctx, agentdto.CreateAgentRequest{
			Name:         "Auto Bot",
			SystemPrompt: "hello",
			Status:       "active",
		})
		if err != nil {
			return err
		}
		agID = ag.ID
		return nil
	})
	require.NoError(t, err)
	return chID, agID
}

// TestAutomationRoutes covers all /automations endpoints.
func TestAutomationRoutes(t *testing.T) {
	db := testsupport.NewPostgres(t)
	r, adminTok, companyID := newRouter(t, db)
	channelID, agentID := seedChannelAndAgent(t, db, companyID)

	// 1. POST /automations — unauthenticated.
	w := do(t, r, http.MethodPost, "/automations", gin.H{
		"channel_id": channelID.String(), "agent_id": agentID.String(),
	}, "")
	require.Equal(t, http.StatusUnauthorized, w.Code, "create-noauth: %s", w.Body.String())

	// 2. POST /automations — missing required channel_id.
	w = do(t, r, http.MethodPost, "/automations", gin.H{
		"agent_id": agentID.String(),
	}, adminTok)
	require.Equal(t, http.StatusBadRequest, w.Code, "create-no-channel: %s", w.Body.String())

	// 3. POST /automations — missing required agent_id.
	w = do(t, r, http.MethodPost, "/automations", gin.H{
		"channel_id": channelID.String(),
	}, adminTok)
	require.Equal(t, http.StatusBadRequest, w.Code, "create-no-agent: %s", w.Body.String())

	// 4. POST /automations — happy path.
	w = do(t, r, http.MethodPost, "/automations", gin.H{
		"channel_id": channelID.String(),
		"agent_id":   agentID.String(),
		"is_active":  true,
	}, adminTok)
	require.Equal(t, http.StatusCreated, w.Code, "create: %s", w.Body.String())
	var auto struct {
		ID        string `json:"id"`
		ChannelID string `json:"channel_id"`
		AgentID   string `json:"agent_id"`
		IsActive  bool   `json:"is_active"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &auto))
	require.NotEmpty(t, auto.ID)
	require.Equal(t, channelID.String(), auto.ChannelID)
	require.Equal(t, agentID.String(), auto.AgentID)
	require.True(t, auto.IsActive)

	// 5. POST /automations — duplicate (1 active automation per channel).
	w = do(t, r, http.MethodPost, "/automations", gin.H{
		"channel_id": channelID.String(),
		"agent_id":   agentID.String(),
		"is_active":  true,
	}, adminTok)
	require.Equal(t, http.StatusConflict, w.Code, "create-dup: %s", w.Body.String())

	// 6. GET /automations — list includes created automation.
	w = do(t, r, http.MethodGet, "/automations", nil, adminTok)
	require.Equal(t, http.StatusOK, w.Code, "list: %s", w.Body.String())
	var list struct {
		Automations []struct {
			ID string `json:"id"`
		} `json:"automations"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &list))
	require.Len(t, list.Automations, 1)
	require.Equal(t, auto.ID, list.Automations[0].ID)

	// 7. GET /automations — unauthenticated.
	w = do(t, r, http.MethodGet, "/automations", nil, "")
	require.Equal(t, http.StatusUnauthorized, w.Code, "list-noauth: %s", w.Body.String())

	// 8. GET /automations/:id — happy path.
	w = do(t, r, http.MethodGet, "/automations/"+auto.ID, nil, adminTok)
	require.Equal(t, http.StatusOK, w.Code, "get: %s", w.Body.String())
	require.Contains(t, w.Body.String(), auto.ID)

	// 9. GET /automations/:id — not found.
	w = do(t, r, http.MethodGet, "/automations/"+uuid.New().String(), nil, adminTok)
	require.Equal(t, http.StatusNotFound, w.Code, "get-notfound: %s", w.Body.String())

	// 10. PUT /automations/:id — update.
	w = do(t, r, http.MethodPut, "/automations/"+auto.ID, gin.H{
		"is_active":        false,
		"fallback_message": "Off hours",
	}, adminTok)
	require.Equal(t, http.StatusOK, w.Code, "update: %s", w.Body.String())
	require.Contains(t, w.Body.String(), "Off hours")

	// 11. PUT /automations/:id — debounce out of range (>60) must be 400.
	w = do(t, r, http.MethodPut, "/automations/"+auto.ID, gin.H{"debounce_seconds": 99}, adminTok)
	require.Equal(t, http.StatusBadRequest, w.Code, "update-bad-debounce: %s", w.Body.String())

	// 12. PUT /automations/:id — not found.
	w = do(t, r, http.MethodPut, "/automations/"+uuid.New().String(), gin.H{"is_active": false}, adminTok)
	require.Equal(t, http.StatusNotFound, w.Code, "update-notfound: %s", w.Body.String())

	// 13. DELETE /automations/:id — delete.
	w = do(t, r, http.MethodDelete, "/automations/"+auto.ID, nil, adminTok)
	require.Equal(t, http.StatusNoContent, w.Code, "delete: %s", w.Body.String())

	// 14. GET /automations — empty after delete.
	w = do(t, r, http.MethodGet, "/automations", nil, adminTok)
	require.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &list))
	require.Empty(t, list.Automations, "list-after-delete: %s", w.Body.String())

	// 15. DELETE /automations/:id — not found.
	w = do(t, r, http.MethodDelete, "/automations/"+auto.ID, nil, adminTok)
	require.Equal(t, http.StatusNotFound, w.Code, "delete-notfound: %s", w.Body.String())
}
