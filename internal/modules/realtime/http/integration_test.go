package http_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	realtimehttp "github.com/Edu0liver/prototype-healthy-api/internal/modules/realtime/http"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/config"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/redisx"
	"github.com/Edu0liver/prototype-healthy-api/pkg/token"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func newServer(t *testing.T) (*httptest.Server, *token.Manager) {
	t.Helper()
	tok := token.New("test-secret-please-change", 15*time.Minute, time.Hour)
	rdb := &redisx.Client{Client: redis.NewClient(&redis.Options{Addr: "127.0.0.1:0"})}
	h := realtimehttp.NewHandler(tok, rdb, &config.Config{}, zap.NewNop())

	gin.SetMode(gin.TestMode)
	r := gin.New()
	realtimehttp.RegisterRoutes(r.Group(""), h)
	return httptest.NewServer(r), tok
}

func TestWS_RejectsMissingToken(t *testing.T) {
	srv, _ := newServer(t)
	defer srv.Close()
	resp, err := http.Get(srv.URL + "/ws")
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestWS_RejectsInvalidToken(t *testing.T) {
	srv, _ := newServer(t)
	defer srv.Close()
	resp, err := http.Get(srv.URL + "/ws?token=garbage")
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestWS_RejectsNonAccessToken(t *testing.T) {
	srv, tok := newServer(t)
	defer srv.Close()
	// A refresh token must not authorize the realtime socket.
	refresh, err := tok.GenerateRefresh(uuid.New(), uuid.New())
	require.NoError(t, err)
	resp, err := http.Get(srv.URL + "/ws?token=" + refresh)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestWS_UpgradesWithValidAccessToken(t *testing.T) {
	srv, tok := newServer(t)
	defer srv.Close()
	access, err := tok.GenerateAccess(uuid.New(), uuid.New(), "admin")
	require.NoError(t, err)

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "/ws?token=" + access
	conn, resp, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err, "valid access token must upgrade to a websocket")
	require.Equal(t, http.StatusSwitchingProtocols, resp.StatusCode)
	conn.Close()
}
