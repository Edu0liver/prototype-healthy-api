package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func ginContextFor(r *http.Request) *gin.Context {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = r
	return c
}

func TestOriginChecker(t *testing.T) {
	allowed := []string{"https://panel.lumia.app", "app.tenant.com"}
	check := originChecker(allowed)

	cases := []struct {
		name   string
		host   string
		origin string
		want   bool
	}{
		{"no origin (non-browser)", "api.lumia.app", "", true},
		{"same origin", "api.lumia.app", "https://api.lumia.app", true},
		{"same origin different scheme", "api.lumia.app", "http://api.lumia.app", true},
		{"allowlisted full origin", "api.lumia.app", "https://panel.lumia.app", true},
		{"allowlisted host only", "api.lumia.app", "https://app.tenant.com", true},
		{"foreign origin", "api.lumia.app", "https://evil.example.com", false},
		{"malformed origin", "api.lumia.app", "://not a url", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "http://"+tc.host+"/ws", nil)
			r.Host = tc.host
			if tc.origin != "" {
				r.Header.Set("Origin", tc.origin)
			}
			if got := check(r); got != tc.want {
				t.Errorf("origin=%q host=%q: got %v want %v", tc.origin, tc.host, got, tc.want)
			}
		})
	}
}

func TestBearerToken(t *testing.T) {
	cases := []struct {
		name  string
		proto string
		query string
		want  string
	}{
		{"subprotocol", "bearer, abc.def.ghi", "", "abc.def.ghi"},
		{"subprotocol case-insensitive", "Bearer, tok123", "", "tok123"},
		{"query fallback", "", "qtok", "qtok"},
		{"subprotocol wins over query", "bearer, ptok", "qtok", "ptok"},
		{"malformed subprotocol falls back to query", "bearer", "qtok", "qtok"},
		{"none", "", "", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			target := "/ws"
			if tc.query != "" {
				target += "?token=" + tc.query
			}
			r := httptest.NewRequest(http.MethodGet, target, nil)
			if tc.proto != "" {
				r.Header.Set("Sec-WebSocket-Protocol", tc.proto)
			}
			c := ginContextFor(r)
			if got := bearerToken(c); got != tc.want {
				t.Errorf("proto=%q query=%q: got %q want %q", tc.proto, tc.query, got, tc.want)
			}
		})
	}
}
