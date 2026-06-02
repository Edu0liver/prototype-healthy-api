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

	conversationhttp "github.com/Edu0liver/prototype-healthy-api/internal/modules/conversation/http"
	conversationrepo "github.com/Edu0liver/prototype-healthy-api/internal/modules/conversation/infra/repository"
	conversationservice "github.com/Edu0liver/prototype-healthy-api/internal/modules/conversation/service"
	iamrepo "github.com/Edu0liver/prototype-healthy-api/internal/modules/iam/infra/repository"
	iamservice "github.com/Edu0liver/prototype-healthy-api/internal/modules/iam/service"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/config"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/database"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/events"
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

// newRouter returns (router, adminToken).
func newRouter(t *testing.T, db *database.DB) (*gin.Engine, string) {
	t.Helper()
	cfg := &config.Config{}
	cfg.App.PublicBaseURL = "http://test"
	cfg.JWT.Secret = "test-secret-please-change"
	cfg.JWT.AccessTTL = 15 * time.Minute
	cfg.JWT.RefreshTTL = time.Hour

	tok := token.New(cfg.JWT.Secret, cfg.JWT.AccessTTL, cfg.JWT.RefreshTTL)
	slug := "cv-" + uuid.New().String()[:8]
	companyID := seedCompany(t, db, slug)

	ctx := context.Background()
	iamSvc := iamservice.New(iamrepo.New(), db, tok, noopMailer{}, cfg, nil)
	_, err := iamSvc.RegisterFirstAdmin(ctx, companyID, "admin@conv.test", "secret123", "Admin")
	require.NoError(t, err)
	tokens, _, err := iamSvc.Login(ctx, "admin@conv.test", "secret123")
	require.NoError(t, err)

	var rdb *redisx.Client
	mw := middleware.New(cfg, db, rdb, tok, zap.NewNop())

	// events.Publisher with nil rdb: Publish is never called by panel read endpoints.
	pub := events.New(nil, zap.NewNop())
	svc := conversationservice.New(conversationrepo.New(), pub)
	h := conversationhttp.NewHandler(svc)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	conversationhttp.RegisterRoutes(r.Group(""), h, mw)
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

// TestConversationRoutes covers the read-only /conversations panel endpoints.
func TestConversationRoutes(t *testing.T) {
	db := testsupport.NewPostgres(t)
	r, adminTok := newRouter(t, db)

	// 1. GET /conversations — unauthenticated.
	w := do(t, r, http.MethodGet, "/conversations", nil, "")
	require.Equal(t, http.StatusUnauthorized, w.Code, "list-noauth: %s", w.Body.String())

	// 2. GET /conversations — authenticated, empty tenant (no conversations yet).
	w = do(t, r, http.MethodGet, "/conversations", nil, adminTok)
	require.Equal(t, http.StatusOK, w.Code, "list: %s", w.Body.String())
	var list struct {
		Conversations []any `json:"conversations"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &list))
	require.Empty(t, list.Conversations)

	// 3. GET /conversations/:id — invalid UUID.
	w = do(t, r, http.MethodGet, "/conversations/not-a-uuid", nil, adminTok)
	require.Equal(t, http.StatusBadRequest, w.Code, "get-bad-id: %s", w.Body.String())

	// 4. GET /conversations/:id — not found.
	w = do(t, r, http.MethodGet, "/conversations/"+uuid.New().String(), nil, adminTok)
	require.Equal(t, http.StatusNotFound, w.Code, "get-notfound: %s", w.Body.String())

	// 5. GET /conversations/:id/messages — invalid UUID.
	w = do(t, r, http.MethodGet, "/conversations/not-a-uuid/messages", nil, adminTok)
	require.Equal(t, http.StatusBadRequest, w.Code, "messages-bad-id: %s", w.Body.String())

	// 6. GET /conversations/:id/messages — conversation not found.
	w = do(t, r, http.MethodGet, "/conversations/"+uuid.New().String()+"/messages", nil, adminTok)
	require.Equal(t, http.StatusNotFound, w.Code, "messages-notfound: %s", w.Body.String())

	// 7. GET /conversations/:id — unauthenticated.
	w = do(t, r, http.MethodGet, "/conversations/"+uuid.New().String(), nil, "")
	require.Equal(t, http.StatusUnauthorized, w.Code, "get-noauth: %s", w.Body.String())
}
