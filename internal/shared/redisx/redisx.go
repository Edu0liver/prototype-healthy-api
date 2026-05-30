// Package redisx wraps go-redis and provides the platform's stateful primitives:
// idempotency (SETNX), distributed per-conversation locks (Redlock-style),
// debounce buffers, conversation-state mirror, and stream queue helpers.
package redisx

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Edu0liver/prototype-healthy-api/internal/shared/config"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// Client wraps *redis.Client with domain helpers.
type Client struct {
	*redis.Client
}

// New parses the Redis URL, connects, and registers shutdown.
func New(cfg *config.Config, lc fx.Lifecycle, log *zap.Logger) (*Client, error) {
	opts, err := redis.ParseURL(cfg.Redis.URL)
	if err != nil {
		return nil, fmt.Errorf("redisx: parse url: %w", err)
	}
	rdb := redis.NewClient(opts)

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error { return rdb.Ping(ctx).Err() },
		OnStop: func(context.Context) error {
			log.Info("closing redis client")
			return rdb.Close()
		},
	})
	return &Client{rdb}, nil
}

// ---- Idempotency ----------------------------------------------------------

// FirstSeen returns true if key was set now (i.e. not seen before) within ttl.
func (c *Client) FirstSeen(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	return c.SetNX(ctx, key, "1", ttl).Result()
}

// ---- Distributed lock (Redlock single-node) -------------------------------

// ErrLockNotAcquired is returned when the lock is held by someone else.
var ErrLockNotAcquired = errors.New("redisx: lock not acquired")

// Lock holds a per-key lock with a unique token for safe release.
type Lock struct {
	key   string
	token string
	c     *Client
}

var releaseScript = redis.NewScript(`
if redis.call("get", KEYS[1]) == ARGV[1] then
  return redis.call("del", KEYS[1])
else
  return 0
end`)

// AcquireLock attempts to grab key for ttl. Returns ErrLockNotAcquired if busy.
func (c *Client) AcquireLock(ctx context.Context, key string, ttl time.Duration) (*Lock, error) {
	token := uuid.NewString()
	ok, err := c.SetNX(ctx, key, token, ttl).Result()
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrLockNotAcquired
	}
	return &Lock{key: key, token: token, c: c}, nil
}

// Release frees the lock only if we still own it.
func (l *Lock) Release(ctx context.Context) error {
	return releaseScript.Run(ctx, l.c, []string{l.key}, l.token).Err()
}

// ---- Debounce buffer ------------------------------------------------------

// PushBuffer appends a message fragment to the conversation buffer.
func (c *Client) PushBuffer(ctx context.Context, convID string, fragment string, ttl time.Duration) error {
	key := bufferKey(convID)
	pipe := c.TxPipeline()
	pipe.RPush(ctx, key, fragment)
	pipe.Expire(ctx, key, ttl)
	_, err := pipe.Exec(ctx)
	return err
}

// DrainBuffer returns and clears all buffered fragments for a conversation.
func (c *Client) DrainBuffer(ctx context.Context, convID string) ([]string, error) {
	key := bufferKey(convID)
	pipe := c.TxPipeline()
	get := pipe.LRange(ctx, key, 0, -1)
	pipe.Del(ctx, key)
	if _, err := pipe.Exec(ctx); err != nil {
		return nil, err
	}
	return get.Val(), nil
}

func bufferKey(convID string) string { return "buffer:conv:" + convID }

// ---- Conversation state mirror -------------------------------------------

// State values mirror conversations.state.
const (
	StateAI     = "ai"
	StateHuman  = "human"
	StateClosed = "closed"
)

func stateKey(convID string) string { return "conv:state:" + convID }
func blockKey(convID string) string { return "block:conv:" + convID }

// SetState stores the conversation state (no expiry — operational source of truth).
func (c *Client) SetState(ctx context.Context, convID, state string) error {
	return c.Set(ctx, stateKey(convID), state, 0).Err()
}

// GetState returns the cached state or redis.Nil error on miss.
func (c *Client) GetState(ctx context.Context, convID string) (string, error) {
	return c.Get(ctx, stateKey(convID)).Result()
}

// Block sets the passive-handover flag (fromMe human takeover) with a TTL.
func (c *Client) Block(ctx context.Context, convID string, ttl time.Duration) error {
	return c.Set(ctx, blockKey(convID), "1", ttl).Err()
}

// Unblock clears the passive-handover flag.
func (c *Client) Unblock(ctx context.Context, convID string) error {
	return c.Del(ctx, blockKey(convID)).Err()
}

// IsBlocked reports whether the conversation is under passive human control.
func (c *Client) IsBlocked(ctx context.Context, convID string) (bool, error) {
	n, err := c.Exists(ctx, blockKey(convID)).Result()
	return n > 0, err
}

// LockKey builds the per-conversation lock key.
func LockKey(convID string) string { return "lock:conv:" + convID }

// DedupeKey builds the idempotency key for a message external id.
func DedupeKey(externalID string) string { return "dedupe:msg:" + externalID }
