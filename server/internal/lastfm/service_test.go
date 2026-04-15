package lastfm

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"
)

type stubSource struct {
	tracks map[string][]*NowPlaying
	err    error
	calls  map[string]int
}

func (s *stubSource) HasCredentials() bool {
	return true
}

func (s *stubSource) FetchNowPlayingForUser(_ context.Context, username string) (*NowPlaying, error) {
	if s.calls == nil {
		s.calls = make(map[string]int)
	}

	callIndex := s.calls[username]
	s.calls[username]++

	if s.err != nil {
		return nil, s.err
	}

	if len(s.tracks[username]) == 0 {
		return nil, nil
	}

	if callIndex >= len(s.tracks[username]) {
		return s.tracks[username][len(s.tracks[username])-1], nil
	}

	return s.tracks[username][callIndex], nil
}

func newTestStore(t *testing.T) *Store {
	t.Helper()

	store, err := NewStore(filepath.Join(t.TempDir(), "app.db"))
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

func TestServiceCachesTracksForFourMinutes(t *testing.T) {
	t.Parallel()

	store := newTestStore(t)
	source := &stubSource{
		tracks: map[string][]*NowPlaying{
			"kisakay": {
				&NowPlaying{
					Title:     "Ghost Song",
					Artist:    "Kisakay",
					Timestamp: "live now",
					URL:       "https://example.com/ghost-song",
					IsLive:    true,
				},
			},
		},
	}

	now := time.Date(2026, time.April, 15, 10, 0, 0, 0, time.UTC)
	service := NewService(source, store, DefaultCacheTTL)
	service.now = func() time.Time { return now }

	first, err := service.GetNowPlaying(context.Background(), "Kisakay")
	if err != nil {
		t.Fatalf("GetNowPlaying() first call error = %v", err)
	}

	second, err := service.GetNowPlaying(context.Background(), "kisakay")
	if err != nil {
		t.Fatalf("GetNowPlaying() second call error = %v", err)
	}

	if source.calls["kisakay"] != 1 {
		t.Fatalf("expected one upstream call within cache TTL, got %d", source.calls["kisakay"])
	}

	if first == nil || second == nil || second.Title != first.Title {
		t.Fatalf("expected cached track to be returned, got first=%#v second=%#v", first, second)
	}
}

func TestServiceRefreshesAfterCacheTTLExpires(t *testing.T) {
	t.Parallel()

	store := newTestStore(t)
	source := &stubSource{
		tracks: map[string][]*NowPlaying{
			"kisakay": {
				&NowPlaying{
					Title:     "Track One",
					Artist:    "Kisakay",
					Timestamp: "recent scrobble",
					URL:       "https://example.com/one",
				},
				&NowPlaying{
					Title:     "Track Two",
					Artist:    "Kisakay",
					Timestamp: "live now",
					URL:       "https://example.com/two",
					IsLive:    true,
				},
			},
		},
	}

	now := time.Date(2026, time.April, 15, 10, 0, 0, 0, time.UTC)
	service := NewService(source, store, DefaultCacheTTL)
	service.now = func() time.Time { return now }

	first, err := service.GetNowPlaying(context.Background(), "kisakay")
	if err != nil {
		t.Fatalf("GetNowPlaying() first call error = %v", err)
	}

	now = now.Add(DefaultCacheTTL + time.Second)

	second, err := service.GetNowPlaying(context.Background(), "kisakay")
	if err != nil {
		t.Fatalf("GetNowPlaying() second call error = %v", err)
	}

	if source.calls["kisakay"] != 2 {
		t.Fatalf("expected two upstream calls after TTL expiry, got %d", source.calls["kisakay"])
	}

	if first == nil || second == nil || second.Title != "Track Two" {
		t.Fatalf("expected refreshed track after TTL expiry, got first=%#v second=%#v", first, second)
	}
}

func TestServiceFallsBackToStaleCacheWhenRefreshFails(t *testing.T) {
	t.Parallel()

	store := newTestStore(t)
	source := &stubSource{
		tracks: map[string][]*NowPlaying{
			"kisakay": {
				&NowPlaying{
					Title:     "Cached Track",
					Artist:    "Kisakay",
					Timestamp: "recent scrobble",
					URL:       "https://example.com/cached",
				},
			},
		},
	}

	now := time.Date(2026, time.April, 15, 10, 0, 0, 0, time.UTC)
	service := NewService(source, store, DefaultCacheTTL)
	service.now = func() time.Time { return now }

	if _, err := service.GetNowPlaying(context.Background(), "kisakay"); err != nil {
		t.Fatalf("GetNowPlaying() seed call error = %v", err)
	}

	source.err = errors.New("last.fm unavailable")
	now = now.Add(DefaultCacheTTL + time.Second)

	track, err := service.GetNowPlaying(context.Background(), "kisakay")
	if err != nil {
		t.Fatalf("GetNowPlaying() stale fallback error = %v", err)
	}

	if track == nil || track.Title != "Cached Track" {
		t.Fatalf("expected stale cached track when refresh fails, got %#v", track)
	}

	if source.calls["kisakay"] != 2 {
		t.Fatalf("expected refresh attempt before stale fallback, got %d upstream calls", source.calls["kisakay"])
	}
}
