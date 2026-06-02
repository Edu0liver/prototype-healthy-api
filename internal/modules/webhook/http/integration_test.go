package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	conversationrepo "github.com/Edu0liver/prototype-healthy-api/internal/modules/conversation/infra/repository"
	conversationservice "github.com/Edu0liver/prototype-healthy-api/internal/modules/conversation/service"
	webhookhttp "github.com/Edu0liver/prototype-healthy-api/internal/modules/webhook/http"
	webhookrepo "github.com/Edu0liver/prototype-healthy-api/internal/modules/webhook/infra/repository"
	webhookservice "github.com/Edu0liver/prototype-healthy-api/internal/modules/webhook/service"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/config"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/database"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/events"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/redisx"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/testsupport"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

const testWebhookToken = "super-secret-webhook-token"

// newRouter builds the webhook router. rdb is nil; the handler returns 200
// even when Process fails, so Redis absence only causes a logged error.
func newRouter(t *testing.T, db *database.DB, token string) *gin.Engine {
	t.Helper()
	cfg := &config.Config{}
	cfg.Evolution.WebhookToken = token

	var rdb *redisx.Client
	pub := events.New(nil, zap.NewNop())
	convSvc := conversationservice.New(conversationrepo.New(), pub)
	svc := webhookservice.New(db, rdb, convSvc, webhookrepo.New(), pub, cfg, zap.NewNop())
	h := webhookhttp.NewHandler(svc, cfg, zap.NewNop())

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(gin.Recovery()) // catches panics from nil rdb on unexpected code paths
	webhookhttp.RegisterRoutes(r.Group(""), h)
	return r
}

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

func do(t *testing.T, r *gin.Engine, method, path string, body any, authHeader string) *httptest.ResponseRecorder {
	t.Helper()
	var rdr io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		require.NoError(t, err)
		rdr = bytes.NewReader(b)
	}
	req := httptest.NewRequest(method, path, rdr)
	req.Header.Set("Content-Type", "application/json")
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

// TestWebhookTokenValidation covers POST /webhooks/evolution token enforcement.
func TestWebhookTokenValidation(t *testing.T) {
	db := testsupport.NewPostgres(t)

	t.Run("correct token returns 200", func(t *testing.T) {
		r := newRouter(t, db, testWebhookToken)
		w := do(t, r, http.MethodPost, "/webhooks/evolution",
			gin.H{"event": "MESSAGES_UPSERT", "instance": "test", "data": gin.H{}},
			"Bearer "+testWebhookToken)
		require.Equal(t, http.StatusOK, w.Code, "correct-token: %s", w.Body.String())
		require.Contains(t, w.Body.String(), "received")
	})

	t.Run("wrong token returns 401", func(t *testing.T) {
		r := newRouter(t, db, testWebhookToken)
		w := do(t, r, http.MethodPost, "/webhooks/evolution",
			gin.H{"event": "MESSAGES_UPSERT", "instance": "test", "data": gin.H{}},
			"Bearer wrong-token")
		require.Equal(t, http.StatusUnauthorized, w.Code, "wrong-token: %s", w.Body.String())
	})

	t.Run("missing auth header returns 401", func(t *testing.T) {
		r := newRouter(t, db, testWebhookToken)
		w := do(t, r, http.MethodPost, "/webhooks/evolution",
			gin.H{"event": "MESSAGES_UPSERT", "instance": "test", "data": gin.H{}},
			"")
		require.Equal(t, http.StatusUnauthorized, w.Code, "no-token: %s", w.Body.String())
	})

	t.Run("no token configured accepts any request", func(t *testing.T) {
		r := newRouter(t, db, "") // empty token = no validation
		w := do(t, r, http.MethodPost, "/webhooks/evolution",
			gin.H{"event": "MESSAGES_UPSERT", "instance": "test", "data": gin.H{}},
			"")
		require.Equal(t, http.StatusOK, w.Code, "no-token-cfg: %s", w.Body.String())
	})
}

// TestWebhookUnknownInstance verifies that an unknown Evolution instance
// is handled gracefully (process fails → logged error → still 200).
func TestWebhookUnknownInstance(t *testing.T) {
	db := testsupport.NewPostgres(t)
	seedCompany(t, db, "wh-"+uuid.New().String()[:8])
	r := newRouter(t, db, "")

	w := do(t, r, http.MethodPost, "/webhooks/evolution",
		gin.H{
			"event":    "MESSAGES_UPSERT",
			"instance": "unknown-instance",
			"data":     gin.H{"key": gin.H{}},
		}, "")
	// Handler always returns 200 to prevent provider retries.
	require.Equal(t, http.StatusOK, w.Code, "unknown-instance: %s", w.Body.String())
	require.Contains(t, w.Body.String(), "received")
}
