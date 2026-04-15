package ratelimit

import (
	"encoding/json"
	"math"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	defaultRequests        = 60
	defaultWindow          = time.Minute
	defaultCleanupInterval = 5 * time.Minute
)

type Config struct {
	Enabled         bool
	Requests        int
	Window          time.Duration
	Burst           int
	CleanupInterval time.Duration
}

type KeyFunc func(*http.Request) string

type Middleware struct {
	config      Config
	keyFunc     KeyFunc
	now         func() time.Time
	mu          sync.Mutex
	buckets     map[string]*bucket
	nextCleanup time.Time
}

type bucket struct {
	tokens   float64
	last     time.Time
	lastSeen time.Time
}

type decision struct {
	allowed    bool
	remaining  int
	retryAfter time.Duration
}

func NewMiddleware(cfg Config, keyFunc KeyFunc) *Middleware {
	normalized := normalizeConfig(cfg)
	if !normalized.Enabled {
		return nil
	}

	return &Middleware{
		config:  normalized,
		keyFunc: keyFunc,
		now:     time.Now,
		buckets: make(map[string]*bucket),
	}
}

func (m *Middleware) Wrap(next http.Handler) http.Handler {
	if m == nil {
		return next
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			next.ServeHTTP(w, r)
			return
		}

		key := ""
		if m.keyFunc != nil {
			key = strings.TrimSpace(m.keyFunc(r))
		}

		if key == "" {
			next.ServeHTTP(w, r)
			return
		}

		decision := m.allow(key)
		w.Header().Set("X-RateLimit-Limit", strconv.Itoa(m.config.Requests))
		w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(decision.remaining))

		if !decision.allowed {
			retryAfterSeconds := int(math.Ceil(decision.retryAfter.Seconds()))
			if retryAfterSeconds < 1 {
				retryAfterSeconds = 1
			}

			w.Header().Set("Retry-After", strconv.Itoa(retryAfterSeconds))
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusTooManyRequests)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"error": "rate limit exceeded",
			})
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (m *Middleware) allow(key string) decision {
	now := m.now()

	m.mu.Lock()
	defer m.mu.Unlock()

	m.cleanup(now)

	b := m.buckets[key]
	if b == nil {
		b = &bucket{
			tokens:   float64(m.config.Burst),
			last:     now,
			lastSeen: now,
		}
		m.buckets[key] = b
	}

	ratePerSecond := float64(m.config.Requests) / m.config.Window.Seconds()
	elapsed := now.Sub(b.last).Seconds()
	if elapsed > 0 {
		b.tokens = minFloat(float64(m.config.Burst), b.tokens+(elapsed*ratePerSecond))
		b.last = now
	}

	b.lastSeen = now

	if b.tokens < 1 {
		missingTokens := 1 - b.tokens
		retryAfter := time.Duration(math.Ceil((missingTokens / ratePerSecond) * float64(time.Second)))
		return decision{
			allowed:    false,
			remaining:  0,
			retryAfter: retryAfter,
		}
	}

	b.tokens--

	return decision{
		allowed:   true,
		remaining: int(math.Floor(b.tokens)),
	}
}

func (m *Middleware) cleanup(now time.Time) {
	if !m.nextCleanup.IsZero() && now.Before(m.nextCleanup) {
		return
	}

	staleAfter := m.config.Window * 2
	if staleAfter < m.config.CleanupInterval {
		staleAfter = m.config.CleanupInterval
	}

	cutoff := now.Add(-staleAfter)
	for key, bucket := range m.buckets {
		if bucket.lastSeen.Before(cutoff) {
			delete(m.buckets, key)
		}
	}

	m.nextCleanup = now.Add(m.config.CleanupInterval)
}

func normalizeConfig(cfg Config) Config {
	if cfg.Window <= 0 {
		cfg.Window = defaultWindow
	}

	if cfg.Requests <= 0 {
		cfg.Requests = defaultRequests
	}

	if cfg.Burst <= 0 {
		cfg.Burst = cfg.Requests
	}

	if cfg.CleanupInterval <= 0 {
		cfg.CleanupInterval = defaultCleanupInterval
	}

	if !cfg.Enabled && cfg == (Config{}) {
		cfg.Enabled = true
	}

	return cfg
}

func minFloat(a, b float64) float64 {
	if a < b {
		return a
	}

	return b
}
