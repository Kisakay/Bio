package ratelimit

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestMiddlewareBlocksWhenLimitIsExceeded(t *testing.T) {
	t.Parallel()

	middleware := NewMiddleware(Config{
		Enabled:  true,
		Requests: 2,
		Window:   time.Second,
		Burst:    2,
	}, func(*http.Request) string {
		return "203.0.113.10"
	})

	handler := middleware.Wrap(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	for range 2 {
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/test", nil))

		if recorder.Code != http.StatusNoContent {
			t.Fatalf("expected request to pass, got status %d", recorder.Code)
		}
	}

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/test", nil))

	if recorder.Code != http.StatusTooManyRequests {
		t.Fatalf("expected rate limited status, got %d", recorder.Code)
	}

	if got := recorder.Header().Get("Retry-After"); got == "" {
		t.Fatal("expected Retry-After header to be set")
	}
}

func TestMiddlewareRefillsTokensOverTime(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.April, 15, 10, 0, 0, 0, time.UTC)
	middleware := NewMiddleware(Config{
		Enabled:  true,
		Requests: 1,
		Window:   time.Second,
		Burst:    1,
	}, func(*http.Request) string {
		return "203.0.113.10"
	})
	middleware.now = func() time.Time {
		return now
	}

	handler := middleware.Wrap(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	first := httptest.NewRecorder()
	handler.ServeHTTP(first, httptest.NewRequest(http.MethodGet, "/api/test", nil))
	if first.Code != http.StatusNoContent {
		t.Fatalf("expected first request to pass, got status %d", first.Code)
	}

	blocked := httptest.NewRecorder()
	handler.ServeHTTP(blocked, httptest.NewRequest(http.MethodGet, "/api/test", nil))
	if blocked.Code != http.StatusTooManyRequests {
		t.Fatalf("expected second immediate request to be blocked, got status %d", blocked.Code)
	}

	now = now.Add(1100 * time.Millisecond)

	afterRefill := httptest.NewRecorder()
	handler.ServeHTTP(afterRefill, httptest.NewRequest(http.MethodGet, "/api/test", nil))
	if afterRefill.Code != http.StatusNoContent {
		t.Fatalf("expected request after refill to pass, got status %d", afterRefill.Code)
	}
}

func TestMiddlewareSkipsOptionsRequests(t *testing.T) {
	t.Parallel()

	middleware := NewMiddleware(Config{
		Enabled:  true,
		Requests: 1,
		Window:   time.Minute,
		Burst:    1,
	}, func(*http.Request) string {
		return "203.0.113.10"
	})

	calls := 0
	handler := middleware.Wrap(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(http.StatusNoContent)
	}))

	for range 2 {
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodOptions, "/api/test", nil))

		if recorder.Code != http.StatusNoContent {
			t.Fatalf("expected OPTIONS request to pass, got status %d", recorder.Code)
		}
	}

	if calls != 2 {
		t.Fatalf("expected wrapped handler to be called twice, got %d", calls)
	}
}
