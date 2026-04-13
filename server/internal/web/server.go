package web

import (
	"net/http"

	"kisakay/server/internal/config"
	"kisakay/server/internal/lastfm"
	"kisakay/server/internal/views"
)

type Server struct {
	config       config.Config
	lastfmClient *lastfm.Client
	viewStore    *views.Store
}

func NewServer(cfg config.Config, lastfmClient *lastfm.Client, viewStore *views.Store) *Server {
	return &Server{
		config:       cfg,
		lastfmClient: lastfmClient,
		viewStore:    viewStore,
	}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/lastfm", s.handleLastfm)
	mux.HandleFunc("/api/views", s.handleViews)

	return withCommonHeaders(mux)
}
