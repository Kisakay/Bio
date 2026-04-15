package web

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"kisakay/server/internal/config"
	"kisakay/server/internal/lastfm"
	"kisakay/server/internal/views"
)

func TestHandlerRootListsAPIRoutes(t *testing.T) {
	t.Parallel()

	store, err := views.NewStore(filepath.Join(t.TempDir(), "views.json"))
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}
	defer func() {
		if err := store.Close(); err != nil {
			t.Fatalf("Close() error = %v", err)
		}
	}()

	server := NewServer(config.Config{
		ViewHashSecret:   "test-secret",
		RateLimitEnabled: false,
	}, lastfm.NewClient("", "Kisakay", &http.Client{}), store)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/", nil)

	server.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}

	var payload rootResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if payload.Name != "Kisakay API" {
		t.Fatalf("expected API name to be set, got %q", payload.Name)
	}

	if len(payload.Routes) != 2 {
		t.Fatalf("expected 2 routes, got %d", len(payload.Routes))
	}

	if payload.Routes[0].Path != "/api/lastfm" {
		t.Fatalf("expected first route to be /api/lastfm, got %q", payload.Routes[0].Path)
	}

	if payload.Routes[1].Path != "/api/views/:username" {
		t.Fatalf("expected second route to be /api/views/:username, got %q", payload.Routes[1].Path)
	}
}

func TestHandlerViewsAreScopedByUsername(t *testing.T) {
	t.Parallel()

	store, err := views.NewStore(filepath.Join(t.TempDir(), "views.db"))
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}
	defer func() {
		if err := store.Close(); err != nil {
			t.Fatalf("Close() error = %v", err)
		}
	}()

	server := NewServer(config.Config{
		ViewHashSecret:   "test-secret",
		RateLimitEnabled: false,
	}, lastfm.NewClient("", "Kisakay", &http.Client{}), store)

	post := httptest.NewRequest(http.MethodPost, "/api/views/kisakay", nil)
	post.RemoteAddr = "127.0.0.1:1234"

	postRecorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(postRecorder, post)

	if postRecorder.Code != http.StatusOK {
		t.Fatalf("expected POST status %d, got %d", http.StatusOK, postRecorder.Code)
	}

	get := httptest.NewRequest(http.MethodGet, "/api/views/kisakay", nil)
	getRecorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(getRecorder, get)

	if getRecorder.Code != http.StatusOK {
		t.Fatalf("expected GET status %d, got %d", http.StatusOK, getRecorder.Code)
	}

	var payload viewsResponse
	if err := json.Unmarshal(getRecorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if payload.Count != 1 {
		t.Fatalf("expected count = 1, got %d", payload.Count)
	}

	other := httptest.NewRequest(http.MethodGet, "/api/views/another-user", nil)
	otherRecorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(otherRecorder, other)

	if err := json.Unmarshal(otherRecorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("json.Unmarshal(other) error = %v", err)
	}

	if payload.Count != 0 {
		t.Fatalf("expected other profile count = 0, got %d", payload.Count)
	}
}

func TestHandlerViewsRejectsInvalidUsername(t *testing.T) {
	t.Parallel()

	store, err := views.NewStore(filepath.Join(t.TempDir(), "views.db"))
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}
	defer func() {
		if err := store.Close(); err != nil {
			t.Fatalf("Close() error = %v", err)
		}
	}()

	server := NewServer(config.Config{
		ViewHashSecret:   "test-secret",
		RateLimitEnabled: false,
	}, lastfm.NewClient("", "Kisakay", &http.Client{}), store)

	request := httptest.NewRequest(http.MethodGet, "/api/views/not.valid", nil)
	recorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, recorder.Code)
	}

	if !strings.Contains(recorder.Body.String(), "username") {
		t.Fatalf("expected username error message, got %q", recorder.Body.String())
	}
}

func TestHandlerViewsLegacyHashesDoNotDoubleCount(t *testing.T) {
	t.Parallel()

	store, err := views.NewStore(filepath.Join(t.TempDir(), "views.db"))
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}
	defer func() {
		if err := store.Close(); err != nil {
			t.Fatalf("Close() error = %v", err)
		}
	}()

	if _, _, err := store.ImportLegacy(
		context.Background(),
		"kisakay",
		[]string{views.HashIP("127.0.0.1", "test-secret")},
	); err != nil {
		t.Fatalf("ImportLegacy() error = %v", err)
	}

	server := NewServer(config.Config{
		ViewHashSecret:   "test-secret",
		RateLimitEnabled: false,
	}, lastfm.NewClient("", "Kisakay", &http.Client{}), store)

	post := httptest.NewRequest(http.MethodPost, "/api/views/kisakay", nil)
	post.RemoteAddr = "127.0.0.1:1234"

	recorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(recorder, post)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected POST status %d, got %d", http.StatusOK, recorder.Code)
	}

	var payload viewsResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if payload.Added {
		t.Fatal("expected legacy migrated visitor not to increment the counter")
	}

	if payload.Count != 1 {
		t.Fatalf("expected count = 1 after legacy check, got %d", payload.Count)
	}
}
