package web

import (
	"net/http"

	"kisakay/server/internal/views"
)

type viewsResponse struct {
	Count int  `json:"count"`
	Added bool `json:"added,omitempty"`
}

func (s *Server) handleLastfm(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if !s.lastfmClient.HasCredentials() {
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "missing LASTFM_API_KEY",
		})
		return
	}

	track, err := s.lastfmClient.FetchNowPlaying(r.Context())
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{
			"error": err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, track)
}

func (s *Server) handleViews(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, viewsResponse{
			Count: s.viewStore.Count(),
		})
	case http.MethodPost:
		ip := clientIP(r)
		if ip == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{
				"error": "unable to resolve client ip",
			})
			return
		}

		hash := views.HashIP(ip, s.config.ViewHashSecret)
		added, count, err := s.viewStore.Add(hash)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{
				"error": "unable to persist view",
			})
			return
		}

		writeJSON(w, http.StatusOK, viewsResponse{
			Count: count,
			Added: added,
		})
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}
