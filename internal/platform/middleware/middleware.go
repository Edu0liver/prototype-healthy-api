// Package middleware provides cross-cutting Gin middleware: request id, panic
// recovery, JWT auth, per-request tenant transaction (RLS), RBAC and rate limit.
package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/Edu0liver/prototype-healthy-api/internal/platform/appctx"
	"github.com/Edu0liver/prototype-healthy-api/internal/platform/config"
	"github.com/Edu0liver/prototype-healthy-api/internal/platform/database"
	"github.com/Edu0liver/prototype-healthy-api/internal/platform/httputil"
	"github.com/Edu0liver/prototype-healthy-api/internal/platform/redisx"
	"github.com/Edu0liver/prototype-healthy-api/internal/platform/token"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// RBAC roles.
const (
	RoleAdmin            = "admin"
	RoleOperator         = "operator"
	RoleKnowledgeManager = "knowledge_manager"
)

// Middleware bundles dependencies shared by the middleware functions.
type Middleware struct {
	cfg    *config.Config
	db     *database.DB
	rdb    *redisx.Client
	tokens *token.Manager
	log    *zap.Logger
}

// New constructs the middleware bundle.
func New(cfg *config.Config, db *database.DB, rdb *redisx.Client, tokens *token.Manager, log *zap.Logger) *Middleware {
	return &Middleware{cfg: cfg, db: db, rdb: rdb, tokens: tokens, log: log}
}

const requestIDHeader = "X-Request-ID"

// RequestID ensures every request carries a trace id (echoed in the response).
func (m *Middleware) RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		rid := c.GetHeader(requestIDHeader)
		if rid == "" {
			rid = uuid.NewString()
		}
		c.Set("request_id", rid)
		c.Writer.Header().Set(requestIDHeader, rid)
		c.Next()
	}
}

// Recovery converts panics into 500s and logs them with zap.
func (m *Middleware) Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				m.log.Error("panic recovered",
					zap.Any("panic", r),
					zap.String("path", c.FullPath()),
					zap.String("request_id", c.GetString("request_id")),
				)
				c.AbortWithStatusJSON(http.StatusInternalServerError, httputil.ErrorResponse{
					Error: "internal_server_error", Message: "unexpected error",
				})
			}
		}()
		c.Next()
	}
}

// Auth validates the Bearer access token and stores the identity in context.
func (m *Middleware) Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		raw := c.GetHeader("Authorization")
		if !strings.HasPrefix(raw, "Bearer ") {
			httputil.Fail(c, httputil.Unauthorized("missing bearer token"))
			c.Abort()
			return
		}
		claims, err := m.tokens.Parse(strings.TrimPrefix(raw, "Bearer "))
		if err != nil || claims.Type != token.TypeAccess {
			httputil.Fail(c, httputil.Unauthorized("invalid or expired token"))
			c.Abort()
			return
		}
		companyID, err1 := uuid.Parse(claims.CompanyID)
		userID, err2 := uuid.Parse(claims.UserID)
		if err1 != nil || err2 != nil {
			httputil.Fail(c, httputil.Unauthorized("malformed token claims"))
			c.Abort()
			return
		}
		id := appctx.Identity{CompanyID: companyID, UserID: userID, Role: claims.Role}
		c.Request = c.Request.WithContext(appctx.With(c.Request.Context(), id))
		c.Set("identity", id)
		c.Next()
	}
}

// Tenant opens a tenant-scoped transaction (SET LOCAL app.current_company_id)
// for the request and commits on success / rolls back on error or panic. Must
// run after Auth. Repositories pick up the handle via database.MustTx(ctx).
func (m *Middleware) Tenant() gin.HandlerFunc {
	return func(c *gin.Context) {
		companyID := appctx.CompanyID(c.Request.Context())
		if companyID == uuid.Nil {
			httputil.Fail(c, httputil.Unauthorized("no tenant context"))
			c.Abort()
			return
		}
		if !m.allowTenant(c, companyID) {
			httputil.Fail(c, httputil.NewDomainError(http.StatusTooManyRequests, "rate_limited", "rate limit exceeded"))
			c.Abort()
			return
		}
		tx, err := m.db.BeginTenant(c.Request.Context(), companyID)
		if err != nil {
			httputil.Fail(c, err)
			c.Abort()
			return
		}
		committed := false
		defer func() {
			if !committed {
				tx.Rollback()
			}
		}()

		c.Request = c.Request.WithContext(database.WithTx(c.Request.Context(), tx))
		c.Next()

		if len(c.Errors) == 0 && c.Writer.Status() < http.StatusBadRequest {
			if err := tx.Commit().Error; err != nil {
				m.log.Error("tenant tx commit failed", zap.Error(err))
			} else {
				committed = true
			}
		}
	}
}

// allowTenant applies a per-tenant fixed-window (1 min) rate limit via Redis.
// Fails open if Redis is unavailable.
func (m *Middleware) allowTenant(c *gin.Context, companyID uuid.UUID) bool {
	limit := int64(m.cfg.Security.RateLimitPerMinute)
	if limit <= 0 {
		return true
	}
	key := "ratelimit:company:" + companyID.String()
	n, err := m.rdb.Incr(c.Request.Context(), key).Result()
	if err != nil {
		return true // fail open
	}
	if n == 1 {
		m.rdb.Expire(c.Request.Context(), key, time.Minute)
	}
	return n <= limit
}

// RBAC restricts a route to the given roles.
func (m *Middleware) RBAC(roles ...string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(roles))
	for _, r := range roles {
		allowed[r] = struct{}{}
	}
	return func(c *gin.Context) {
		role := appctx.Role(c.Request.Context())
		if _, ok := allowed[role]; !ok {
			httputil.Fail(c, httputil.Forbidden("insufficient role"))
			c.Abort()
			return
		}
		c.Next()
	}
}

// RateLimit applies a fixed-window per-tenant limit using Redis counters.
func (m *Middleware) RateLimit(limit int64, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		companyID := appctx.CompanyID(c.Request.Context())
		if companyID == uuid.Nil {
			c.Next()
			return
		}
		key := "ratelimit:company:" + companyID.String()
		n, err := m.rdb.Incr(c.Request.Context(), key).Result()
		if err == nil {
			if n == 1 {
				m.rdb.Expire(c.Request.Context(), key, window)
			}
			if n > limit {
				httputil.Fail(c, httputil.NewDomainError(http.StatusTooManyRequests, "rate_limited", "rate limit exceeded"))
				c.Abort()
				return
			}
		}
		c.Next()
	}
}
