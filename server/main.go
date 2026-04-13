package main

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const (
	defaultLastfmUser = "Kisakay"
	defaultPort       = "8080"
)

type config struct {
	LastfmAPIKey   string
	LastfmUser     string
	ViewHashSecret string
	ViewStorePath  string
	Port           string
}

type apiServer struct {
	config    config
	client    *http.Client
	viewStore *viewStore
}

type viewStore struct {
	mu     sync.Mutex
	path   string
	hashes map[string]struct{}
}

type persistedViews struct {
	Hashes []string `json:"hashes"`
}

type recentTracksResponse struct {
	RecentTracks struct {
		Track json.RawMessage `json:"track"`
	} `json:"recenttracks"`
}

type lastfmTrackPayload struct {
	Name   string              `json:"name"`
	URL    string              `json:"url"`
	Artist lastfmTextPayload   `json:"artist"`
	Image  []lastfmTextPayload `json:"image"`
	Attr   struct {
		NowPlaying string `json:"nowplaying"`
	} `json:"@attr"`
	Date struct {
		Text string `json:"#text"`
	} `json:"date"`
}

type lastfmTextPayload struct {
	Text string `json:"#text"`
	Name string `json:"name"`
}

type nowPlayingResponse struct {
	Title     string  `json:"title"`
	Artist    string  `json:"artist"`
	Artwork   *string `json:"artwork"`
	Timestamp string  `json:"timestamp"`
	URL       string  `json:"url"`
	IsLive    bool    `json:"isLive"`
}

type viewsResponse struct {
	Count int  `json:"count"`
	Added bool `json:"added,omitempty"`
}

func main() {
	loadEnvFile(".env")
	loadEnvFile(".env.local")
	loadEnvFile(filepath.Join("..", ".env"))
	loadEnvFile(filepath.Join("..", ".env.local"))

	cfg := config{
		LastfmAPIKey:   firstNonEmpty(os.Getenv("LASTFM_API_KEY"), os.Getenv("VITE_LASTFM_API_KEY")),
		LastfmUser:     firstNonEmpty(os.Getenv("LASTFM_USERNAME"), defaultLastfmUser),
		ViewHashSecret: firstNonEmpty(os.Getenv("VIEW_HASH_SECRET"), os.Getenv("LASTFM_API_KEY"), os.Getenv("VITE_LASTFM_API_KEY"), "kisakay-dev-view-secret"),
		ViewStorePath:  firstNonEmpty(os.Getenv("VIEW_STORE_PATH"), filepath.Join("server-data", "views.json")),
		Port:           firstNonEmpty(os.Getenv("PORT"), defaultPort),
	}

	store, err := newViewStore(cfg.ViewStorePath)
	if err != nil {
		log.Fatalf("unable to initialize view store: %v", err)
	}

	server := &apiServer{
		config: cfg,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		viewStore: store,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/lastfm", server.handleLastfm)
	mux.HandleFunc("/api/views", server.handleViews)

	addr := ":" + cfg.Port
	log.Printf("API listening on http://127.0.0.1%s", addr)

	if err := http.ListenAndServe(addr, withCommonHeaders(mux)); err != nil {
		log.Fatal(err)
	}
}

func (s *apiServer) handleLastfm(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if s.config.LastfmAPIKey == "" {
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "missing LASTFM_API_KEY",
		})
		return
	}

	track, err := s.fetchNowPlaying(r.Context())
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{
			"error": err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, track)
}

func (s *apiServer) handleViews(w http.ResponseWriter, r *http.Request) {
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

		hash := hashIP(ip, s.config.ViewHashSecret)
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

func (s *apiServer) fetchNowPlaying(ctx context.Context) (*nowPlayingResponse, error) {
	endpoint := url.URL{
		Scheme: "https",
		Host:   "ws.audioscrobbler.com",
		Path:   "/2.0/",
	}

	query := endpoint.Query()
	query.Set("method", "user.getrecenttracks")
	query.Set("user", s.config.LastfmUser)
	query.Set("api_key", s.config.LastfmAPIKey)
	query.Set("format", "json")
	query.Set("limit", "1")
	query.Set("extended", "1")
	endpoint.RawQuery = query.Encode()

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, err
	}

	response, err := s.client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("last.fm returned %d", response.StatusCode)
	}

	var payload recentTracksResponse
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return nil, err
	}

	track, err := parseTrack(payload.RecentTracks.Track)
	if err != nil {
		return nil, err
	}

	if track == nil {
		return nil, nil
	}

	return track, nil
}

func parseTrack(raw json.RawMessage) (*nowPlayingResponse, error) {
	if len(raw) == 0 {
		return nil, nil
	}

	var tracks []lastfmTrackPayload
	if err := json.Unmarshal(raw, &tracks); err != nil {
		var single lastfmTrackPayload
		if errSingle := json.Unmarshal(raw, &single); errSingle != nil {
			return nil, errors.New("unable to decode last.fm track")
		}
		tracks = []lastfmTrackPayload{single}
	}

	if len(tracks) == 0 {
		return nil, nil
	}

	track := tracks[0]
	title := strings.TrimSpace(track.Name)
	artist := strings.TrimSpace(firstNonEmpty(track.Artist.Name, track.Artist.Text))
	if title == "" || artist == "" {
		return nil, errors.New("last.fm track payload missing title or artist")
	}

	var artwork *string
	for index := len(track.Image) - 1; index >= 0; index-- {
		candidate := strings.TrimSpace(track.Image[index].Text)
		if candidate == "" {
			continue
		}

		artwork = &candidate
		break
	}

	isLive := strings.TrimSpace(track.Attr.NowPlaying) == "true"
	timestamp := strings.TrimSpace(track.Date.Text)
	if timestamp == "" {
		timestamp = "recent scrobble"
	}
	if isLive {
		timestamp = "live now"
	}

	trackURL := strings.TrimSpace(track.URL)
	if trackURL != "" && !strings.HasPrefix(trackURL, "http") {
		trackURL = "https://www.last.fm" + trackURL
	}
	if trackURL == "" {
		trackURL = "https://www.last.fm/user/" + defaultLastfmUser
	}

	return &nowPlayingResponse{
		Title:     title,
		Artist:    artist,
		Artwork:   artwork,
		Timestamp: timestamp,
		URL:       trackURL,
		IsLive:    isLive,
	}, nil
}

func newViewStore(path string) (*viewStore, error) {
	store := &viewStore{
		path:   path,
		hashes: make(map[string]struct{}),
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return store, nil
		}
		return nil, err
	}

	if len(data) == 0 {
		return store, nil
	}

	var persisted persistedViews
	if err := json.Unmarshal(data, &persisted); err != nil {
		return nil, err
	}

	for _, hash := range persisted.Hashes {
		trimmed := strings.TrimSpace(hash)
		if trimmed == "" {
			continue
		}
		store.hashes[trimmed] = struct{}{}
	}

	return store, nil
}

func (v *viewStore) Count() int {
	v.mu.Lock()
	defer v.mu.Unlock()

	return len(v.hashes)
}

func (v *viewStore) Add(hash string) (bool, int, error) {
	v.mu.Lock()
	defer v.mu.Unlock()

	if _, exists := v.hashes[hash]; exists {
		return false, len(v.hashes), nil
	}

	v.hashes[hash] = struct{}{}
	if err := v.persistLocked(); err != nil {
		delete(v.hashes, hash)
		return false, len(v.hashes), err
	}

	return true, len(v.hashes), nil
}

func (v *viewStore) persistLocked() error {
	hashes := make([]string, 0, len(v.hashes))
	for hash := range v.hashes {
		hashes = append(hashes, hash)
	}

	payload := persistedViews{Hashes: hashes}
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	tempPath := v.path + ".tmp"
	if err := os.WriteFile(tempPath, data, 0o600); err != nil {
		return err
	}

	return os.Rename(tempPath, v.path)
}

func hashIP(ip, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(ip))
	return hex.EncodeToString(mac.Sum(nil))
}

func clientIP(r *http.Request) string {
	for _, header := range []string{"CF-Connecting-IP", "X-Forwarded-For", "X-Real-IP"} {
		value := strings.TrimSpace(r.Header.Get(header))
		if value == "" {
			continue
		}

		if header == "X-Forwarded-For" {
			value = strings.TrimSpace(strings.Split(value, ",")[0])
		}

		if parsed := normalizeIP(value); parsed != "" {
			return parsed
		}
	}

	host, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr))
	if err == nil {
		return normalizeIP(host)
	}

	return normalizeIP(r.RemoteAddr)
}

func normalizeIP(value string) string {
	ip := net.ParseIP(strings.TrimSpace(value))
	if ip == nil {
		return ""
	}

	return ip.String()
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func withCommonHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		next.ServeHTTP(w, r)
	})
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			return trimmed
		}
	}

	return ""
}

func loadEnvFile(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		key, value, found := strings.Cut(trimmed, "=")
		if !found {
			continue
		}

		key = strings.TrimSpace(key)
		if key == "" || os.Getenv(key) != "" {
			continue
		}

		value = strings.Trim(strings.TrimSpace(value), `"'`)
		_ = os.Setenv(key, value)
	}
}
