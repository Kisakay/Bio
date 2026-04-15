package lastfm

import (
	"context"
	"errors"
	"time"
)

const DefaultCacheTTL = 4 * time.Minute

var ErrMissingCredentials = errors.New("missing LASTFM_API_KEY")

type Source interface {
	HasCredentials() bool
	FetchNowPlayingForUser(context.Context, string) (*NowPlaying, error)
}

type Service struct {
	source Source
	store  *Store
	ttl    time.Duration
	now    func() time.Time
}

func NewService(source Source, store *Store, ttl time.Duration) *Service {
	if ttl <= 0 {
		ttl = DefaultCacheTTL
	}

	return &Service{
		source: source,
		store:  store,
		ttl:    ttl,
		now:    time.Now,
	}
}

func (s *Service) GetNowPlaying(ctx context.Context, username string) (*NowPlaying, error) {
	normalized, err := NormalizeUsername(username)
	if err != nil {
		return nil, err
	}

	cached, err := s.store.Get(ctx, normalized)
	if err != nil {
		return nil, err
	}

	if cached != nil && s.now().UTC().Sub(cached.FetchedAt) < s.ttl {
		return cached.Track, nil
	}

	if !s.source.HasCredentials() {
		if cached != nil {
			return cached.Track, nil
		}

		return nil, ErrMissingCredentials
	}

	track, err := s.source.FetchNowPlayingForUser(ctx, normalized)
	if err != nil {
		if cached != nil {
			return cached.Track, nil
		}

		return nil, err
	}

	if err := s.store.Upsert(ctx, normalized, track, s.now().UTC()); err != nil {
		return nil, err
	}

	return track, nil
}
