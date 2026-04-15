package web

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"kisakay/server/internal/config"
	"kisakay/server/internal/lastfm"
	"kisakay/server/internal/views"
)

type stubLastfmService struct {
	track             *lastfm.NowPlaying
	err               error
	receivedUsernames []string
}

func (s *stubLastfmService) GetNowPlaying(_ context.Context, username string) (*lastfm.NowPlaying, error) {
	s.receivedUsernames = append(s.receivedUsernames, username)
	return s.track, s.err
}

func newTestStore(t *testing.T) *views.Store {
	t.Helper()

	store, err := views.NewStore(filepath.Join(t.TempDir(), "app.db"))
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}

	t.Cleanup(func() {
		if err := store.Close(); err != nil {
			t.Fatalf("Close() error = %v", err)
		}
	})

	return store
}

func TestHandlerRootListsAPIRoutes(t *testing.T) {
	t.Parallel()

	store := newTestStore(t)

	server := NewServer(config.Config{
		ViewHashSecret:   "test-secret",
		RateLimitEnabled: false,
	}, &stubLastfmService{}, store)

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

	if len(payload.Routes) != 3 {
		t.Fatalf("expected 3 routes, got %d", len(payload.Routes))
	}

	if payload.Routes[0].Path != "/api/lastfm/:username" {
		t.Fatalf("expected first route to be /api/lastfm/:username, got %q", payload.Routes[0].Path)
	}

	if payload.Routes[1].Path != "/api/lastfm" {
		t.Fatalf("expected second route to be /api/lastfm, got %q", payload.Routes[1].Path)
	}

	if payload.Routes[2].Path != "/api/views/:username" {
		t.Fatalf("expected third route to be /api/views/:username, got %q", payload.Routes[2].Path)
	}
}

func TestHandlerViewsAreScopedByUsername(t *testing.T) {
	t.Parallel()

	store := newTestStore(t)

	server := NewServer(config.Config{
		ViewHashSecret:   "test-secret",
		RateLimitEnabled: false,
	}, &stubLastfmService{}, store)

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

	store := newTestStore(t)

	server := NewServer(config.Config{
		ViewHashSecret:   "test-secret",
		RateLimitEnabled: false,
	}, &stubLastfmService{}, store)

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

	store := newTestStore(t)

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
	}, &stubLastfmService{}, store)

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

func TestHandlerLastfmUsesUsernameFromPath(t *testing.T) {
	t.Parallel()

	store := newTestStore(t)
	lastfmService := &stubLastfmService{
		track: &lastfm.NowPlaying{
			Title:     "Ghost Song",
			Artist:    "Kisakay",
			Timestamp: "live now",
			URL:       "https://www.last.fm/music/Kisakay/_/Ghost+Song",
			IsLive:    true,
		},
	}

	server := NewServer(config.Config{
		LastfmUser:       "fallback-user",
		ViewHashSecret:   "test-secret",
		RateLimitEnabled: false,
	}, lastfmService, store)

	request := httptest.NewRequest(http.MethodGet, "/api/lastfm/Test.User", nil)
	recorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}

	if len(lastfmService.receivedUsernames) != 1 || lastfmService.receivedUsernames[0] != "test.user" {
		t.Fatalf("expected normalized username to be passed to service, got %#v", lastfmService.receivedUsernames)
	}

	var payload lastfm.NowPlaying
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if payload.Title != "Ghost Song" {
		t.Fatalf("expected track title to be returned, got %q", payload.Title)
	}
}

func TestHandlerLastfmUsesDefaultUsernameWhenPathHasNoParam(t *testing.T) {
	t.Parallel()

	store := newTestStore(t)
	lastfmService := &stubLastfmService{}

	server := NewServer(config.Config{
		LastfmUser:       "Default.User",
		ViewHashSecret:   "test-secret",
		RateLimitEnabled: false,
	}, lastfmService, store)

	request := httptest.NewRequest(http.MethodGet, "/api/lastfm", nil)
	recorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}

	if len(lastfmService.receivedUsernames) != 1 || lastfmService.receivedUsernames[0] != "default.user" {
		t.Fatalf("expected default username to be passed to service, got %#v", lastfmService.receivedUsernames)
	}
}

func TestHandlerLastfmReturnsInternalServerErrorWhenCredentialsAreMissing(t *testing.T) {
	t.Parallel()

	store := newTestStore(t)
	server := NewServer(config.Config{
		LastfmUser:       "kisakay",
		ViewHashSecret:   "test-secret",
		RateLimitEnabled: false,
	}, &stubLastfmService{err: lastfm.ErrMissingCredentials}, store)

	request := httptest.NewRequest(http.MethodGet, "/api/lastfm/kisakay", nil)
	recorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, recorder.Code)
	}
}

func TestHandlerLastfmReturnsBadGatewayForUpstreamFailures(t *testing.T) {
	t.Parallel()

	store := newTestStore(t)
	server := NewServer(config.Config{
		LastfmUser:       "kisakay",
		ViewHashSecret:   "test-secret",
		RateLimitEnabled: false,
	}, &stubLastfmService{err: errors.New("upstream failed")}, store)

	request := httptest.NewRequest(http.MethodGet, "/api/lastfm/kisakay", nil)
	recorder := httptest.NewRecorder()
	server.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadGateway {
		t.Fatalf("expected status %d, got %d", http.StatusBadGateway, recorder.Code)
	}
}
