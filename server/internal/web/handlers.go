package web

import (
	"errors"
	"net/http"
	"strings"

	"kisakay/server/internal/views"
)

type viewsResponse struct {
	Count int  `json:"count"`
	Added bool `json:"added,omitempty"`
}

func (s *Server) handleRoot(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	writeJSON(w, http.StatusOK, rootResponse{
		Name:   "Kisakay API",
		Routes: apiRoutes(),
	})
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

func (s *Server) handleViewsByUsername(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	username, err := viewUsernameFromPath(r.URL.Path)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
		return
	}

	switch r.Method {
	case http.MethodGet:
		count, err := s.viewStore.Count(r.Context(), username)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{
				"error": "unable to read views",
			})
			return
		}

		writeJSON(w, http.StatusOK, viewsResponse{
			Count: count,
		})
	case http.MethodPost:
		ip := clientIP(r)
		if ip == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{
				"error": "unable to resolve client ip",
			})
			return
		}

		legacyHash := views.HashIP(ip, s.config.ViewHashSecret)
		hash := views.HashViewer(username, ip, s.config.ViewHashSecret)

		exists, err := s.viewStore.HasAny(r.Context(), username, legacyHash, hash)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{
				"error": "unable to read views",
			})
			return
		}

		if exists {
			count, countErr := s.viewStore.Count(r.Context(), username)
			if countErr != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{
					"error": "unable to read views",
				})
				return
			}

			writeJSON(w, http.StatusOK, viewsResponse{
				Count: count,
				Added: false,
			})
			return
		}

		added, count, err := s.viewStore.Add(r.Context(), username, hash)
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

func viewUsernameFromPath(path string) (string, error) {
	prefix := "/api/views/"
	if !strings.HasPrefix(path, prefix) {
		return "", errors.New("username is required")
	}

	rawUsername := strings.Trim(strings.TrimPrefix(path, prefix), "/")
	if strings.Contains(rawUsername, "/") {
		return "", errors.New("username is required")
	}

	return views.NormalizeUsername(rawUsername)
}
