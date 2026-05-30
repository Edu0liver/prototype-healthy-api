// Package resilience provides retry-with-backoff and a simple circuit breaker
// for outbound integrations (OpenAI, Evolution).
package resilience

import (
	"context"
	"errors"
	"math"
	"math/rand"
	"sync"
	"time"
)

// RetryConfig controls exponential backoff with jitter.
type RetryConfig struct {
	MaxAttempts int
	BaseDelay   time.Duration
	MaxDelay    time.Duration
}

// DefaultRetry is a sensible default for network calls.
func DefaultRetry() RetryConfig {
	return RetryConfig{MaxAttempts: 4, BaseDelay: 200 * time.Millisecond, MaxDelay: 5 * time.Second}
}

// permanentError wraps an error that must not be retried (e.g. a 4xx response).
type permanentError struct{ err error }

func (e *permanentError) Error() string { return e.err.Error() }
func (e *permanentError) Unwrap() error { return e.err }

// Permanent marks an error as non-retryable; Do returns it immediately.
func Permanent(err error) error {
	if err == nil {
		return nil
	}
	return &permanentError{err: err}
}

// Do runs fn with retries, stopping early on ctx cancellation or a Permanent error.
func Do(ctx context.Context, cfg RetryConfig, fn func() error) error {
	var err error
	for attempt := 0; attempt < cfg.MaxAttempts; attempt++ {
		if err = fn(); err == nil {
			return nil
		}
		var perm *permanentError
		if errors.As(err, &perm) {
			return perm.err
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
		delay := backoff(cfg, attempt)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
		}
	}
	return err
}

func backoff(cfg RetryConfig, attempt int) time.Duration {
	d := float64(cfg.BaseDelay) * math.Pow(2, float64(attempt))
	if d > float64(cfg.MaxDelay) {
		d = float64(cfg.MaxDelay)
	}
	jitter := rand.Float64() * d * 0.2
	return time.Duration(d + jitter)
}

// ErrCircuitOpen is returned when the breaker is open.
var ErrCircuitOpen = errors.New("resilience: circuit open")

// Breaker is a minimal circuit breaker.
type Breaker struct {
	mu        sync.Mutex
	failures  int
	threshold int
	openUntil time.Time
	cooldown  time.Duration
}

// NewBreaker builds a breaker that opens after threshold consecutive failures
// and stays open for cooldown.
func NewBreaker(threshold int, cooldown time.Duration) *Breaker {
	return &Breaker{threshold: threshold, cooldown: cooldown}
}

// Execute runs fn unless the circuit is open.
func (b *Breaker) Execute(fn func() error) error {
	b.mu.Lock()
	if time.Now().Before(b.openUntil) {
		b.mu.Unlock()
		return ErrCircuitOpen
	}
	b.mu.Unlock()

	err := fn()

	b.mu.Lock()
	defer b.mu.Unlock()
	if err != nil {
		b.failures++
		if b.failures >= b.threshold {
			b.openUntil = time.Now().Add(b.cooldown)
			b.failures = 0
		}
		return err
	}
	b.failures = 0
	return nil
}
