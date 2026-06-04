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

	agentdto "github.com/Edu0liver/prototype-healthy-api/internal/modules/agent/dto"
	agentrepo "github.com/Edu0liver/prototype-healthy-api/internal/modules/agent/infra/repository"
	agentservice "github.com/Edu0liver/prototype-healthy-api/internal/modules/agent/service"
	iamrepo "github.com/Edu0liver/prototype-healthy-api/internal/modules/iam/infra/repository"
	iamservice "github.com/Edu0liver/prototype-healthy-api/internal/modules/iam/service"
	knowledgehttp "github.com/Edu0liver/prototype-healthy-api/internal/modules/knowledge/http"
	knowledgerepo "github.com/Edu0liver/prototype-healthy-api/internal/modules/knowledge/infra/repository"
	knowledgeservice "github.com/Edu0liver/prototype-healthy-api/internal/modules/knowledge/service"
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

// ---- stubs ------------------------------------------------------------------

type noopMailer struct{}

func (noopMailer) Send(string, string, string) error { return nil }

type noopStore struct{}

func (noopStore) Put(_ context.Context, _ uuid.UUID, key string, _ []byte) (string, error) {
	return key, nil
}
func (noopStore) Get(_ context.Context, _ string) ([]byte, error) { return []byte("ok"), nil }
func (noopStore) Delete(_ context.Context, _ string) error        { return nil }

type noopEmbedder struct{}

func (noopEmbedder) Embed(_ context.Context, inputs []string) ([][]float32, error) {
	out := make([][]float32, len(inputs))
	for i := range out {
		out[i] = make([]float32, 1536)
	}
	return out, nil
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

// newRouter spins up the knowledge Gin router and returns (router, adminToken, companyID).
func newRouter(t *testing.T, db *database.DB) (*gin.Engine, string, uuid.UUID) {
	t.Helper()
	cfg := &config.Config{}
	cfg.App.PublicBaseURL = "http://test"
	cfg.JWT.Secret = "test-secret-please-change"
	cfg.JWT.AccessTTL = 15 * time.Minute
	cfg.JWT.RefreshTTL = time.Hour

	tok := token.New(cfg.JWT.Secret, cfg.JWT.AccessTTL, cfg.JWT.RefreshTTL)

	slug := "kn-" + uuid.New().String()[:8]
	companyID := seedCompany(t, db, slug)
	testsupport.SeedActiveSubscription(t, db, companyID)

	ctx := context.Background()
	iamSvc := iamservice.New(iamrepo.New(), db, tok, noopMailer{}, cfg, nil)
	_, err := iamSvc.RegisterFirstAdmin(ctx, companyID, "admin@knowledge.test", "secret123", "Admin")
	require.NoError(t, err)
	tokens, _, err := iamSvc.Login(ctx, "admin@knowledge.test", "secret123")
	require.NoError(t, err)

	var rdb *redisx.Client
	mw := middleware.New(cfg, db, rdb, tok, zap.NewNop())

	svc := knowledgeservice.New(knowledgerepo.New(), db, noopStore{}, noopEmbedder{}, zap.NewNop())
	h := knowledgehttp.NewHandler(svc)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	knowledgehttp.RegisterRoutes(r.Group(""), h, mw)
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

// ---- tests ------------------------------------------------------------------

// TestKnowledgeBaseRoutes covers CRUD on /knowledge-bases.
func TestKnowledgeBaseRoutes(t *testing.T) {
	db := testsupport.NewPostgres(t)
	r, adminTok, _ := newRouter(t, db)

	// 1. POST /knowledge-bases — unauthenticated.
	w := do(t, r, http.MethodPost, "/knowledge-bases", gin.H{
		"name": "KB", "description": "test",
	}, "")
	require.Equal(t, http.StatusUnauthorized, w.Code, "create-noauth: %s", w.Body.String())

	// 2. POST /knowledge-bases — missing required name.
	w = do(t, r, http.MethodPost, "/knowledge-bases", gin.H{"description": "no name"}, adminTok)
	require.Equal(t, http.StatusBadRequest, w.Code, "create-noname: %s", w.Body.String())

	// 3. POST /knowledge-bases — happy path.
	w = do(t, r, http.MethodPost, "/knowledge-bases", gin.H{
		"name":        "Support KB",
		"description": "Customer support knowledge base",
	}, adminTok)
	require.Equal(t, http.StatusCreated, w.Code, "create: %s", w.Body.String())
	var kb struct {
		ID             string `json:"id"`
		Name           string `json:"name"`
		EmbeddingModel string `json:"embedding_model"`
		ChunkSize      int    `json:"chunk_size"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &kb))
	require.NotEmpty(t, kb.ID)
	require.Equal(t, "Support KB", kb.Name)
	require.NotEmpty(t, kb.EmbeddingModel)
	require.Greater(t, kb.ChunkSize, 0)

	// 4. GET /knowledge-bases — list includes created KB.
	w = do(t, r, http.MethodGet, "/knowledge-bases", nil, adminTok)
	require.Equal(t, http.StatusOK, w.Code, "list: %s", w.Body.String())
	var list struct {
		KBs []struct {
			ID string `json:"id"`
		} `json:"knowledge_bases"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &list))
	require.Len(t, list.KBs, 1)
	require.Equal(t, kb.ID, list.KBs[0].ID)

	// 5. GET /knowledge-bases — unauthenticated.
	w = do(t, r, http.MethodGet, "/knowledge-bases", nil, "")
	require.Equal(t, http.StatusUnauthorized, w.Code, "list-noauth: %s", w.Body.String())

	// 6. GET /knowledge-bases/:id — happy path.
	w = do(t, r, http.MethodGet, "/knowledge-bases/"+kb.ID, nil, adminTok)
	require.Equal(t, http.StatusOK, w.Code, "get: %s", w.Body.String())
	require.Contains(t, w.Body.String(), "Support KB")

	// 7. GET /knowledge-bases/:id — invalid UUID.
	w = do(t, r, http.MethodGet, "/knowledge-bases/not-a-uuid", nil, adminTok)
	require.Equal(t, http.StatusBadRequest, w.Code, "get-bad-id: %s", w.Body.String())

	// 8. GET /knowledge-bases/:id — not found.
	w = do(t, r, http.MethodGet, "/knowledge-bases/"+uuid.New().String(), nil, adminTok)
	require.Equal(t, http.StatusNotFound, w.Code, "get-notfound: %s", w.Body.String())

	// 9. DELETE /knowledge-bases/:id — delete.
	w = do(t, r, http.MethodDelete, "/knowledge-bases/"+kb.ID, nil, adminTok)
	require.Equal(t, http.StatusNoContent, w.Code, "delete: %s", w.Body.String())

	// 10. GET /knowledge-bases — list is empty after delete.
	w = do(t, r, http.MethodGet, "/knowledge-bases", nil, adminTok)
	require.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &list))
	require.Empty(t, list.KBs, "list-after-delete: %s", w.Body.String())

	// 11. DELETE /knowledge-bases/:id — already deleted returns 404.
	w = do(t, r, http.MethodDelete, "/knowledge-bases/"+kb.ID, nil, adminTok)
	require.Equal(t, http.StatusNotFound, w.Code, "delete-notfound: %s", w.Body.String())
}

// TestDocumentRoutes covers text document upload, listing and deletion.
func TestDocumentRoutes(t *testing.T) {
	db := testsupport.NewPostgres(t)
	r, adminTok, _ := newRouter(t, db)

	// Create a KB to work with.
	w := do(t, r, http.MethodPost, "/knowledge-bases", gin.H{
		"name": "Doc KB", "description": "docs test",
	}, adminTok)
	require.Equal(t, http.StatusCreated, w.Code)
	var kb struct {
		ID string `json:"id"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &kb))

	// 1. POST /knowledge-bases/:id/documents/text — missing content must be 400.
	w = do(t, r, http.MethodPost, "/knowledge-bases/"+kb.ID+"/documents/text", gin.H{
		"title": "No Content",
	}, adminTok)
	require.Equal(t, http.StatusBadRequest, w.Code, "upload-no-content: %s", w.Body.String())

	// 2. POST /knowledge-bases/:id/documents/text — happy path.
	w = do(t, r, http.MethodPost, "/knowledge-bases/"+kb.ID+"/documents/text", gin.H{
		"title":   "FAQ",
		"content": "Frequently asked questions content here.",
	}, adminTok)
	require.Equal(t, http.StatusCreated, w.Code, "upload-text: %s", w.Body.String())
	var doc struct {
		ID         string `json:"id"`
		Filename   string `json:"filename"`
		SourceType string `json:"source_type"`
		Status     string `json:"status"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &doc))
	require.NotEmpty(t, doc.ID)
	require.Equal(t, "text", doc.SourceType)
	require.Equal(t, "pending", doc.Status)

	// 3. POST /knowledge-bases/:id/documents/text — unauthenticated.
	w = do(t, r, http.MethodPost, "/knowledge-bases/"+kb.ID+"/documents/text", gin.H{
		"title": "T", "content": "C",
	}, "")
	require.Equal(t, http.StatusUnauthorized, w.Code, "upload-noauth: %s", w.Body.String())

	// 4. POST /knowledge-bases/:id/documents/text — KB not found.
	w = do(t, r, http.MethodPost, "/knowledge-bases/"+uuid.New().String()+"/documents/text", gin.H{
		"title": "T", "content": "C",
	}, adminTok)
	require.Equal(t, http.StatusNotFound, w.Code, "upload-kb-notfound: %s", w.Body.String())

	// 5. GET /knowledge-bases/:id/documents — list includes uploaded doc.
	w = do(t, r, http.MethodGet, "/knowledge-bases/"+kb.ID+"/documents", nil, adminTok)
	require.Equal(t, http.StatusOK, w.Code, "list-docs: %s", w.Body.String())
	var docs struct {
		Documents []struct {
			ID string `json:"id"`
		} `json:"documents"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &docs))
	require.Len(t, docs.Documents, 1)
	require.Equal(t, doc.ID, docs.Documents[0].ID)

	// 6. DELETE /documents/:docId — delete document.
	w = do(t, r, http.MethodDelete, "/documents/"+doc.ID, nil, adminTok)
	require.Equal(t, http.StatusNoContent, w.Code, "delete-doc: %s", w.Body.String())

	// 7. GET /knowledge-bases/:id/documents — empty after delete.
	w = do(t, r, http.MethodGet, "/knowledge-bases/"+kb.ID+"/documents", nil, adminTok)
	require.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &docs))
	require.Empty(t, docs.Documents, "list-after-delete: %s", w.Body.String())

	// 8. DELETE /documents/:docId — already deleted returns 404.
	w = do(t, r, http.MethodDelete, "/documents/"+doc.ID, nil, adminTok)
	require.Equal(t, http.StatusNotFound, w.Code, "delete-doc-notfound: %s", w.Body.String())
}

// TestAgentKBLinkRoutes covers the agent↔knowledge-base linking endpoints.
func TestAgentKBLinkRoutes(t *testing.T) {
	db := testsupport.NewPostgres(t)
	r, adminTok, companyID := newRouter(t, db)

	// Seed an agent for this tenant.
	agentSvc := agentservice.New(agentrepo.New())
	ctx := context.Background()
	err := db.Tenant(ctx, companyID, func(tctx context.Context) error {
		_, err := agentSvc.Create(tctx, agentdto.CreateAgentRequest{
			Name:         "Link Bot",
			SystemPrompt: "You help with linking.",
			Status:       "active",
		})
		return err
	})
	require.NoError(t, err)

	// List agents to get the seeded agent ID.
	var agentID string
	err = db.Tenant(ctx, companyID, func(tctx context.Context) error {
		agents, err := agentSvc.List(tctx)
		if err != nil {
			return err
		}
		require.Len(t, agents, 1)
		agentID = agents[0].ID.String()
		return nil
	})
	require.NoError(t, err)

	// Create a KB.
	w := do(t, r, http.MethodPost, "/knowledge-bases", gin.H{
		"name": "Link KB", "description": "for linking",
	}, adminTok)
	require.Equal(t, http.StatusCreated, w.Code)
	var kb struct {
		ID string `json:"id"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &kb))

	// 1. GET /agents/:id/knowledge-bases — no links yet, empty list.
	w = do(t, r, http.MethodGet, "/agents/"+agentID+"/knowledge-bases", nil, adminTok)
	require.Equal(t, http.StatusOK, w.Code, "list-kbs-empty: %s", w.Body.String())
	var linked struct {
		KBs []struct {
			ID string `json:"id"`
		} `json:"knowledge_bases"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &linked))
	require.Empty(t, linked.KBs)

	// 2. POST /agents/:id/knowledge-bases/:kbId — link.
	w = do(t, r, http.MethodPost, "/agents/"+agentID+"/knowledge-bases/"+kb.ID, nil, adminTok)
	require.Equal(t, http.StatusNoContent, w.Code, "link: %s", w.Body.String())

	// 3. GET /agents/:id/knowledge-bases — list now includes KB.
	w = do(t, r, http.MethodGet, "/agents/"+agentID+"/knowledge-bases", nil, adminTok)
	require.Equal(t, http.StatusOK, w.Code, "list-kbs-linked: %s", w.Body.String())
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &linked))
	require.Len(t, linked.KBs, 1)
	require.Equal(t, kb.ID, linked.KBs[0].ID)

	// 4. DELETE /agents/:id/knowledge-bases/:kbId — unlink.
	w = do(t, r, http.MethodDelete, "/agents/"+agentID+"/knowledge-bases/"+kb.ID, nil, adminTok)
	require.Equal(t, http.StatusNoContent, w.Code, "unlink: %s", w.Body.String())

	// 5. GET /agents/:id/knowledge-bases — empty after unlink.
	w = do(t, r, http.MethodGet, "/agents/"+agentID+"/knowledge-bases", nil, adminTok)
	require.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &linked))
	require.Empty(t, linked.KBs, "list-after-unlink: %s", w.Body.String())

	// 6. Auth enforcement on link routes.
	w = do(t, r, http.MethodPost, "/agents/"+agentID+"/knowledge-bases/"+kb.ID, nil, "")
	require.Equal(t, http.StatusUnauthorized, w.Code, "link-noauth: %s", w.Body.String())
	w = do(t, r, http.MethodGet, "/agents/"+agentID+"/knowledge-bases", nil, "")
	require.Equal(t, http.StatusUnauthorized, w.Code, "list-noauth: %s", w.Body.String())
}
