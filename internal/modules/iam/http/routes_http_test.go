package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"

	iamhttp "github.com/Edu0liver/prototype-healthy-api/internal/modules/iam/http"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/iam/infra/repository"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/iam/service"
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

// capMailer records the last email body so the accept-invite token can be
// recovered (it is only delivered by email in production).
type capMailer struct{ lastBody string }

func (m *capMailer) Send(_, _, html string) error { m.lastBody = html; return nil }

func newRouter(t *testing.T, db *database.DB) (*gin.Engine, *capMailer) {
	t.Helper()
	cfg := &config.Config{}
	cfg.App.PublicBaseURL = "http://test"
	cfg.JWT.Secret = "test-secret-please-change"
	cfg.JWT.AccessTTL = 15 * time.Minute
	cfg.JWT.RefreshTTL = time.Hour

	tok := token.New(cfg.JWT.Secret, cfg.JWT.AccessTTL, cfg.JWT.RefreshTTL)
	mail := &capMailer{}
	svc := service.New(repository.New(), db, tok, mail, cfg)
	h := iamhttp.NewHandler(svc)
	// rdb is nil: the per-tenant rate limit is disabled because
	// cfg.Security.RateLimitPerMinute is 0, so Redis is never touched.
	var rdb *redisx.Client
	mw := middleware.New(cfg, db, rdb, tok, zap.NewNop())

	gin.SetMode(gin.TestMode)
	r := gin.New()
	iamhttp.RegisterRoutes(r.Group(""), h, mw)
	return r, mail
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

// TestAuthAndUserRoutes exercises every route under /auth and /users end-to-end
// through the real Gin router + middleware (auth, tenant-RLS, RBAC).
func TestAuthAndUserRoutes(t *testing.T) {
	db := testsupport.NewPostgres(t)
	r, mail := newRouter(t, db)

	slug := "rt-" + uuid.New().String()[:8]
	seedCompany(t, db, slug)

	const adminEmail = "admin@rt.com"
	const adminPass = "supersecret"

	// 1. POST /auth/register — bootstrap first admin.
	w := do(t, r, http.MethodPost, "/auth/register", gin.H{
		"company_slug": slug, "email": adminEmail, "password": adminPass, "name": "Admin",
	}, "")
	require.Equal(t, http.StatusCreated, w.Code, "register: %s", w.Body.String())

	// 2. POST /auth/login — obtain tokens.
	w = do(t, r, http.MethodPost, "/auth/login", gin.H{
		"company_slug": slug, "email": adminEmail, "password": adminPass,
	}, "")
	require.Equal(t, http.StatusOK, w.Code, "login: %s", w.Body.String())
	var tok struct {
		Access  string `json:"access_token"`
		Refresh string `json:"refresh_token"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &tok))
	require.NotEmpty(t, tok.Access)
	require.NotEmpty(t, tok.Refresh)

	// 3. GET /auth/me — authenticated identity.
	w = do(t, r, http.MethodGet, "/auth/me", nil, tok.Access)
	require.Equal(t, http.StatusOK, w.Code, "me: %s", w.Body.String())
	require.Contains(t, w.Body.String(), adminEmail)

	// 3b. GET /auth/me without a token must be rejected.
	w = do(t, r, http.MethodGet, "/auth/me", nil, "")
	require.Equal(t, http.StatusUnauthorized, w.Code, "me-noauth: %s", w.Body.String())

	// 4. POST /users — admin invites an operator (valid role per the dto enum).
	w = do(t, r, http.MethodPost, "/users", gin.H{
		"email": "agent@rt.com", "name": "Agent", "role": "operator",
	}, tok.Access)
	require.Equal(t, http.StatusCreated, w.Code, "invite: %s", w.Body.String())

	// 5. GET /users — admin lists tenant users (admin + operator).
	w = do(t, r, http.MethodGet, "/users", nil, tok.Access)
	require.Equal(t, http.StatusOK, w.Code, "list: %s", w.Body.String())
	var list struct {
		Users []struct {
			Email string `json:"email"`
		} `json:"users"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &list))
	require.Len(t, list.Users, 2)

	// 6. POST /auth/refresh — rotate tokens.
	w = do(t, r, http.MethodPost, "/auth/refresh", gin.H{"refresh_token": tok.Refresh}, "")
	require.Equal(t, http.StatusOK, w.Code, "refresh: %s", w.Body.String())
	require.Contains(t, w.Body.String(), "access_token")

	// 7. POST /auth/accept-invite — agent sets a password from the emailed link.
	m := regexp.MustCompile(`token=([^"]+)`).FindStringSubmatch(mail.lastBody)
	require.Len(t, m, 2, "invite email body: %s", mail.lastBody)
	w = do(t, r, http.MethodPost, "/auth/accept-invite", gin.H{
		"token": m[1], "password": "newpassword",
	}, "")
	require.Equal(t, http.StatusNoContent, w.Code, "accept-invite: %s", w.Body.String())

	// The agent can now log in with the password it just set.
	w = do(t, r, http.MethodPost, "/auth/login", gin.H{
		"company_slug": slug, "email": "agent@rt.com", "password": "newpassword",
	}, "")
	require.Equal(t, http.StatusOK, w.Code, "agent-login: %s", w.Body.String())

	_ = context.Background()
}
