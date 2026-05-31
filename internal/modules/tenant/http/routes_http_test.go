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

	iamrepo "github.com/Edu0liver/prototype-healthy-api/internal/modules/iam/infra/repository"
	iamservice "github.com/Edu0liver/prototype-healthy-api/internal/modules/iam/service"
	tenantdto "github.com/Edu0liver/prototype-healthy-api/internal/modules/tenant/dto"
	tenanthttp "github.com/Edu0liver/prototype-healthy-api/internal/modules/tenant/http"
	tenantrepo "github.com/Edu0liver/prototype-healthy-api/internal/modules/tenant/infra/repository"
	tenantservice "github.com/Edu0liver/prototype-healthy-api/internal/modules/tenant/service"
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

// newRouter creates the tenant Gin router with a freshly seeded company + admin.
// Returns (router, adminToken, companySlug).
func newRouter(t *testing.T, db *database.DB) (*gin.Engine, string, string) {
	t.Helper()
	cfg := &config.Config{}
	cfg.App.PublicBaseURL = "http://test"
	cfg.JWT.Secret = "test-secret-please-change"
	cfg.JWT.AccessTTL = 15 * time.Minute
	cfg.JWT.RefreshTTL = time.Hour

	tok := token.New(cfg.JWT.Secret, cfg.JWT.AccessTTL, cfg.JWT.RefreshTTL)

	slug := "tn-" + uuid.New().String()[:8]
	tenantSvc := tenantservice.New(tenantrepo.New(), db)
	ctx := context.Background()
	company, err := tenantSvc.CreateCompany(ctx, tenantdto.CreateCompanyRequest{Name: slug, Slug: slug})
	require.NoError(t, err)

	iamSvc := iamservice.New(iamrepo.New(), db, tok, noopMailer{}, cfg)
	_, err = iamSvc.RegisterFirstAdmin(ctx, company.ID, "admin@tenant.test", "secret123", "Admin")
	require.NoError(t, err)
	tokens, _, err := iamSvc.Login(ctx, "admin@tenant.test", "secret123")
	require.NoError(t, err)

	var rdb *redisx.Client
	mw := middleware.New(cfg, db, rdb, tok, zap.NewNop())

	h := tenanthttp.NewHandler(tenantSvc)
	gin.SetMode(gin.TestMode)
	r := gin.New()
	tenanthttp.RegisterRoutes(r.Group(""), h, mw)
	return r, tokens.Access, slug
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

// TestCompanyPublicRoutes covers the public POST /companies endpoint.
func TestCompanyPublicRoutes(t *testing.T) {
	db := testsupport.NewPostgres(t)
	r, _, _ := newRouter(t, db)

	// 1. POST /companies — happy path.
	w := do(t, r, http.MethodPost, "/companies", gin.H{
		"name": "AcmeCo", "slug": "acme-co", "plan": "free",
	}, "")
	require.Equal(t, http.StatusCreated, w.Code, "create: %s", w.Body.String())
	var co struct {
		ID     string `json:"id"`
		Name   string `json:"name"`
		Slug   string `json:"slug"`
		Status string `json:"status"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &co))
	require.NotEmpty(t, co.ID)
	require.Equal(t, "AcmeCo", co.Name)
	require.Equal(t, "acme-co", co.Slug)
	require.Equal(t, "active", co.Status)

	// 2. POST /companies — duplicate slug must be 409.
	w = do(t, r, http.MethodPost, "/companies", gin.H{
		"name": "Another", "slug": "acme-co",
	}, "")
	require.Equal(t, http.StatusConflict, w.Code, "dup-slug: %s", w.Body.String())

	// 3. POST /companies — missing required name must be 400.
	w = do(t, r, http.MethodPost, "/companies", gin.H{"slug": "no-name"}, "")
	require.Equal(t, http.StatusBadRequest, w.Code, "missing-name: %s", w.Body.String())

	// 4. POST /companies — name too short (min=2) must be 400.
	w = do(t, r, http.MethodPost, "/companies", gin.H{"name": "X", "slug": "x"}, "")
	require.Equal(t, http.StatusBadRequest, w.Code, "short-name: %s", w.Body.String())
}

// TestTenantAuthenticatedRoutes covers the admin-authenticated tenant endpoints.
func TestTenantAuthenticatedRoutes(t *testing.T) {
	db := testsupport.NewPostgres(t)
	r, adminTok, _ := newRouter(t, db)

	// 1. GET /company — unauthenticated.
	w := do(t, r, http.MethodGet, "/company", nil, "")
	require.Equal(t, http.StatusUnauthorized, w.Code, "company-noauth: %s", w.Body.String())

	// 2. GET /company — authenticated admin.
	w = do(t, r, http.MethodGet, "/company", nil, adminTok)
	require.Equal(t, http.StatusOK, w.Code, "company: %s", w.Body.String())
	var company struct {
		ID   string `json:"id"`
		Slug string `json:"slug"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &company))
	require.NotEmpty(t, company.ID)

	// 3. PUT /branding — unauthenticated.
	w = do(t, r, http.MethodPut, "/branding", gin.H{"email_sender_name": "Test"}, "")
	require.Equal(t, http.StatusUnauthorized, w.Code, "branding-noauth: %s", w.Body.String())

	// 4. PUT /branding — valid update.
	w = do(t, r, http.MethodPut, "/branding", gin.H{
		"primary_color":     "#123456",
		"email_sender_name": "Support Team",
	}, adminTok)
	require.Equal(t, http.StatusOK, w.Code, "branding: %s", w.Body.String())
	require.Contains(t, w.Body.String(), "#123456")
	require.Contains(t, w.Body.String(), "Support Team")

	// 5. PUT /branding — invalid hex color must be 400.
	w = do(t, r, http.MethodPut, "/branding", gin.H{"primary_color": "notacolor"}, adminTok)
	require.Equal(t, http.StatusBadRequest, w.Code, "branding-badcolor: %s", w.Body.String())

	// 6. POST /domains — add a domain.
	w = do(t, r, http.MethodPost, "/domains", gin.H{
		"domain": "app.acme.test", "is_primary": true,
	}, adminTok)
	require.Equal(t, http.StatusCreated, w.Code, "domain-add: %s", w.Body.String())
	var dom struct {
		ID        string `json:"id"`
		Domain    string `json:"domain"`
		IsPrimary bool   `json:"is_primary"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &dom))
	require.NotEmpty(t, dom.ID)
	require.Equal(t, "app.acme.test", dom.Domain)
	require.True(t, dom.IsPrimary)

	// 7. POST /domains — invalid FQDN must be 400.
	w = do(t, r, http.MethodPost, "/domains", gin.H{"domain": "not a domain"}, adminTok)
	require.Equal(t, http.StatusBadRequest, w.Code, "domain-bad-fqdn: %s", w.Body.String())

	// 8. POST /domains — unauthenticated.
	w = do(t, r, http.MethodPost, "/domains", gin.H{"domain": "other.test"}, "")
	require.Equal(t, http.StatusUnauthorized, w.Code, "domain-noauth: %s", w.Body.String())

	// 9. GET /domains — list includes the domain just added.
	w = do(t, r, http.MethodGet, "/domains", nil, adminTok)
	require.Equal(t, http.StatusOK, w.Code, "domains: %s", w.Body.String())
	var list struct {
		Domains []struct {
			Domain string `json:"domain"`
		} `json:"domains"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &list))
	require.Len(t, list.Domains, 1)
	require.Equal(t, "app.acme.test", list.Domains[0].Domain)

	// 10. GET /domains — unauthenticated.
	w = do(t, r, http.MethodGet, "/domains", nil, "")
	require.Equal(t, http.StatusUnauthorized, w.Code, "domains-noauth: %s", w.Body.String())
}

// TestBrandingByHost covers the public GET /branding?host= lookup.
func TestBrandingByHost(t *testing.T) {
	db := testsupport.NewPostgres(t)
	r, adminTok, _ := newRouter(t, db)

	// Unknown host must return 404.
	w := do(t, r, http.MethodGet, "/branding?host=unknown.no.such.host", nil, "")
	require.Equal(t, http.StatusNotFound, w.Code, "branding-unknown: %s", w.Body.String())

	// Register a domain for the tenant, then look it up.
	w = do(t, r, http.MethodPost, "/domains", gin.H{
		"domain": "brand.acme.test", "is_primary": true,
	}, adminTok)
	require.Equal(t, http.StatusCreated, w.Code, "add-domain: %s", w.Body.String())

	// Update branding so the row is non-empty.
	w = do(t, r, http.MethodPut, "/branding", gin.H{
		"primary_color":     "#ABCDEF",
		"email_sender_name": "Acme",
	}, adminTok)
	require.Equal(t, http.StatusOK, w.Code, "update-branding: %s", w.Body.String())

	// GET /branding?host=brand.acme.test → must return the tenant's branding.
	w = do(t, r, http.MethodGet, "/branding?host=brand.acme.test", nil, "")
	require.Equal(t, http.StatusOK, w.Code, "branding-by-host: %s", w.Body.String())
	require.Contains(t, w.Body.String(), "#ABCDEF")
	require.Contains(t, w.Body.String(), "Acme")
}
