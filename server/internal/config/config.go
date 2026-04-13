package config

import (
	"os"
	"path/filepath"
	"strings"

	"kisakay/server/internal/envfile"
)

const (
	defaultLastfmUser = "Kisakay"
	defaultPort       = "13873"
)

type Config struct {
	LastfmAPIKey   string
	LastfmUser     string
	ViewHashSecret string
	ViewStorePath  string
	Port           string
}

func Load() Config {
	envfile.Load(".env")
	envfile.Load(".env.local")
	envfile.Load(filepath.Join("..", ".env"))
	envfile.Load(filepath.Join("..", ".env.local"))

	return Config{
		LastfmAPIKey:   firstNonEmpty(os.Getenv("LASTFM_API_KEY"), os.Getenv("VITE_LASTFM_API_KEY")),
		LastfmUser:     firstNonEmpty(os.Getenv("LASTFM_USERNAME"), defaultLastfmUser),
		ViewHashSecret: firstNonEmpty(os.Getenv("VIEW_HASH_SECRET"), os.Getenv("LASTFM_API_KEY"), os.Getenv("VITE_LASTFM_API_KEY"), "kisakay-dev-view-secret"),
		ViewStorePath:  firstNonEmpty(os.Getenv("VIEW_STORE_PATH"), filepath.Join("server-data", "views.json")),
		Port:           firstNonEmpty(os.Getenv("PORT"), defaultPort),
	}
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
