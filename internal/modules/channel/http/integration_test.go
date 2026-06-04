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

	channelhttp "github.com/Edu0liver/prototype-healthy-api/internal/modules/channel/http"
	channelrepo "github.com/Edu0liver/prototype-healthy-api/internal/modules/channel/infra/repository"
	channelservice "github.com/Edu0liver/prototype-healthy-api/internal/modules/channel/service"
	iamrepo "github.com/Edu0liver/prototype-healthy-api/internal/modules/iam/infra/repository"
	iamservice "github.com/Edu0liver/prototype-healthy-api/internal/modules/iam/service"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/config"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/database"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/middleware"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/redisx"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/testsupport"
	"github.com/Edu0liver/prototype-healthy-api/pkg/crypto"
	"github.com/Edu0liver/prototype-healthy-api/pkg/evolution"
	"github.com/Edu0liver/prototype-healthy-api/pkg/token"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// ---- stubs ------------------------------------------------------------------

type noopMailer struct{}

func (noopMailer) Send(string, string, string) error { return nil }

// noopEvo implements evolution.Client without hitting any external API.
type noopEvo struct{}

func (noopEvo) CreateInstance(_ context.Context, _ evolution.CreateInstanceRequest) (*evolution.CreateInstanceResult, error) {
	return &evolution.CreateInstanceResult{InstanceID: "test-id", APIKey: "test-key"}, nil
}
func (noopEvo) Connect(_ context.Context, _, _ string) (*evolution.ConnectResult, error) {
	return &evolution.ConnectResult{}, nil
}
func (noopEvo) ConnectionState(_ context.Context, _ string) (string, error) { return "open", nil }
func (noopEvo) Restart(_ context.Context, _ string) error                   { return nil }
func (noopEvo) Logout(_ context.Context, _ string) error                    { return nil }
func (noopEvo) DeleteInstance(_ context.Context, _ string) error            { return nil }
func (noopEvo) SendText(_ context.Context, _, _ string, _ evolution.SendTextRequest) (*evolution.SendResult, error) {
	return &evolution.SendResult{}, nil
}
func (noopEvo) SendPresence(_ context.Context, _, _, _, _ string) error { return nil }
func (noopEvo) MarkAsRead(_ context.Context, _, _, _, _ string) error   { return nil }
func (noopEvo) GetMediaBase64(_ context.Context, _, _, _ string) (string, string, error) {
	return "", "", nil
}

// ---- setup ------------------------------------------------------------------

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

// newRouter returns (router, adminToken).
func newRouter(t *testing.T, db *database.DB) (*gin.Engine, string) {
	t.Helper()
	cfg := &config.Config{}
	cfg.App.PublicBaseURL = "http://test"
	cfg.JWT.Secret = "test-secret-please-change"
	cfg.JWT.AccessTTL = 15 * time.Minute
	cfg.JWT.RefreshTTL = time.Hour
	cfg.Evolution.WebhookURL = "http://test/api/v1/webhooks/evolution"
	cfg.Evolution.WebhookToken = "test-wh-token"

	tok := token.New(cfg.JWT.Secret, cfg.JWT.AccessTTL, cfg.JWT.RefreshTTL)
	slug := "ch-" + uuid.New().String()[:8]
	companyID := seedCompany(t, db, slug)
	testsupport.SeedActiveSubscription(t, db, companyID)

	ctx := context.Background()
	iamSvc := iamservice.New(iamrepo.New(), db, tok, noopMailer{}, cfg, nil)
	_, err := iamSvc.RegisterFirstAdmin(ctx, companyID, "admin@channel.test", "secret123", "Admin")
	require.NoError(t, err)
	tokens, _, err := iamSvc.Login(ctx, "admin@channel.test", "secret123")
	require.NoError(t, err)

	// AES key: 32 zero bytes encoded as 64 hex chars.
	cipher, err := crypto.New("0000000000000000000000000000000000000000000000000000000000000000")
	require.NoError(t, err)

	var rdb *redisx.Client
	mw := middleware.New(cfg, db, rdb, tok, zap.NewNop())

	svc := channelservice.New(channelrepo.New(), noopEvo{}, cipher, cfg, zap.NewNop())
	h := channelhttp.NewHandler(svc)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	channelhttp.RegisterRoutes(r.Group(""), h, mw)
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

// ---- tests ------------------------------------------------------------------

// TestChannelRoutes covers all /channels endpoints using the instagram type
// (which skips the external Evolution API call on creation).
func TestChannelRoutes(t *testing.T) {
	db := testsupport.NewPostgres(t)
	r, adminTok := newRouter(t, db)

	// 1. POST /channels — unauthenticated.
	w := do(t, r, http.MethodPost, "/channels", gin.H{"type": "instagram", "name": "IG"}, "")
	require.Equal(t, http.StatusUnauthorized, w.Code, "create-noauth: %s", w.Body.String())

	// 2. POST /channels — missing required type.
	w = do(t, r, http.MethodPost, "/channels", gin.H{"name": "No Type"}, adminTok)
	require.Equal(t, http.StatusBadRequest, w.Code, "create-no-type: %s", w.Body.String())

	// 3. POST /channels — invalid type.
	w = do(t, r, http.MethodPost, "/channels", gin.H{"type": "telegram"}, adminTok)
	require.Equal(t, http.StatusBadRequest, w.Code, "create-bad-type: %s", w.Body.String())

	// 4. POST /channels — instagram happy path (no Evolution API call).
	w = do(t, r, http.MethodPost, "/channels", gin.H{
		"type": "instagram", "name": "Main IG",
	}, adminTok)
	require.Equal(t, http.StatusCreated, w.Code, "create-instagram: %s", w.Body.String())
	var ch struct {
		ID     string `json:"id"`
		Type   string `json:"type"`
		Name   string `json:"name"`
		Status string `json:"status"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &ch))
	require.NotEmpty(t, ch.ID)
	require.Equal(t, "instagram", ch.Type)
	require.Equal(t, "disconnected", ch.Status)

	// 5. GET /channels — list includes created channel.
	w = do(t, r, http.MethodGet, "/channels", nil, adminTok)
	require.Equal(t, http.StatusOK, w.Code, "list: %s", w.Body.String())
	var list struct {
		Channels []struct {
			ID string `json:"id"`
		} `json:"channels"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &list))
	require.Len(t, list.Channels, 1)
	require.Equal(t, ch.ID, list.Channels[0].ID)

	// 6. GET /channels — unauthenticated.
	w = do(t, r, http.MethodGet, "/channels", nil, "")
	require.Equal(t, http.StatusUnauthorized, w.Code, "list-noauth: %s", w.Body.String())

	// 7. GET /channels/:id — happy path.
	w = do(t, r, http.MethodGet, "/channels/"+ch.ID, nil, adminTok)
	require.Equal(t, http.StatusOK, w.Code, "get: %s", w.Body.String())
	require.Contains(t, w.Body.String(), ch.ID)

	// 8. GET /channels/:id — not found.
	w = do(t, r, http.MethodGet, "/channels/"+uuid.New().String(), nil, adminTok)
	require.Equal(t, http.StatusNotFound, w.Code, "get-notfound: %s", w.Body.String())

	// 9. POST /channels/:id/connect — instagram channels cannot connect (whatsapp only).
	w = do(t, r, http.MethodPost, "/channels/"+ch.ID+"/connect", gin.H{"method": "qr"}, adminTok)
	require.Equal(t, http.StatusBadRequest, w.Code, "connect-instagram: %s", w.Body.String())

	// 10. GET /channels/:id/connection-state — returns state even for instagram.
	w = do(t, r, http.MethodGet, "/channels/"+ch.ID+"/connection-state", nil, adminTok)
	require.Equal(t, http.StatusOK, w.Code, "state: %s", w.Body.String())
	require.Contains(t, w.Body.String(), "state")

	// 11. POST /channels with whatsapp — noopEvo allows creation.
	w = do(t, r, http.MethodPost, "/channels", gin.H{
		"type": "whatsapp", "name": "WA Channel", "number": "+5511999990000",
	}, adminTok)
	require.Equal(t, http.StatusCreated, w.Code, "create-whatsapp: %s", w.Body.String())

	// 12. DELETE /channels/:id — Disconnect marks channel as disconnected but
	// does NOT remove it from the database (no Delete in the channel repo).
	w = do(t, r, http.MethodDelete, "/channels/"+ch.ID, nil, adminTok)
	require.Equal(t, http.StatusNoContent, w.Code, "disconnect: %s", w.Body.String())

	// 13. GET /channels/:id — channel still exists, status is disconnected.
	w = do(t, r, http.MethodGet, "/channels/"+ch.ID, nil, adminTok)
	require.Equal(t, http.StatusOK, w.Code, "get-after-disconnect: %s", w.Body.String())
	require.Contains(t, w.Body.String(), "disconnected")

	// 14. GET /channels — both channels still listed.
	w = do(t, r, http.MethodGet, "/channels", nil, adminTok)
	require.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &list))
	require.Len(t, list.Channels, 2, "list-after-disconnect: %s", w.Body.String())

	// 15. DELETE /channels/:id — disconnect channel not in DB → 404.
	w = do(t, r, http.MethodDelete, "/channels/"+uuid.New().String(), nil, adminTok)
	require.Equal(t, http.StatusNotFound, w.Code, "disconnect-notfound: %s", w.Body.String())
}
