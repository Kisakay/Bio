package web

import (
	"context"
	"net/http"

	"kisakay/server/internal/config"
	"kisakay/server/internal/lastfm"
	"kisakay/server/internal/ratelimit"
	"kisakay/server/internal/views"
)

type Server struct {
	config      config.Config
	lastfm      lastfmService
	viewStore   *views.Store
	rateLimiter *ratelimit.Middleware
}

type lastfmService interface {
	GetNowPlaying(context.Context, string) (*lastfm.NowPlaying, error)
}

func NewServer(cfg config.Config, lastfmClient lastfmService, viewStore *views.Store) *Server {
	return &Server{
		config:    cfg,
		lastfm:    lastfmClient,
		viewStore: viewStore,
		rateLimiter: ratelimit.NewMiddleware(ratelimit.Config{
			Enabled:         cfg.RateLimitEnabled,
			Requests:        cfg.RateLimitRequests,
			Window:          cfg.RateLimitWindow,
			Burst:           cfg.RateLimitBurst,
			CleanupInterval: cfg.RateLimitCleanupInterval,
		}, clientIP),
	}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleRoot)
	mux.HandleFunc("/api/lastfm", s.handleLastfm)
	mux.HandleFunc("/api/lastfm/", s.handleLastfm)
	mux.HandleFunc("/api/views/", s.handleViewsByUsername)

	var handler http.Handler = mux
	if s.rateLimiter != nil {
		handler = s.rateLimiter.Wrap(handler)
	}

	return withCommonHeaders(handler)
}
