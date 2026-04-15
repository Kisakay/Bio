package config

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"kisakay/server/internal/envfile"
)

const (
	defaultLastfmUser        = "Kisakay"
	defaultPort              = "13873"
	defaultRateLimitRequests = 60
	defaultRateLimitBurst    = 20
	defaultRateLimitWindow   = time.Minute
	defaultRateLimitCleanup  = 5 * time.Minute
)

type Config struct {
	LastfmAPIKey             string
	LastfmUser               string
	ViewHashSecret           string
	ViewStorePath            string
	Port                     string
	RateLimitEnabled         bool
	RateLimitRequests        int
	RateLimitBurst           int
	RateLimitWindow          time.Duration
	RateLimitCleanupInterval time.Duration
}

func Load() Config {
	envfile.Load(".env")
	envfile.Load(".env.local")
	envfile.Load(filepath.Join("..", ".env"))
	envfile.Load(filepath.Join("..", ".env.local"))

	return Config{
		LastfmAPIKey:             firstNonEmpty(os.Getenv("LASTFM_API_KEY"), os.Getenv("VITE_LASTFM_API_KEY")),
		LastfmUser:               firstNonEmpty(os.Getenv("LASTFM_USERNAME"), defaultLastfmUser),
		ViewHashSecret:           firstNonEmpty(os.Getenv("VIEW_HASH_SECRET"), os.Getenv("LASTFM_API_KEY"), os.Getenv("VITE_LASTFM_API_KEY"), "kisakay-dev-view-secret"),
		ViewStorePath:            firstNonEmpty(os.Getenv("VIEW_STORE_PATH"), filepath.Join("server-data", "views.json")),
		Port:                     firstNonEmpty(os.Getenv("PORT"), defaultPort),
		RateLimitEnabled:         parseBoolEnv("RATE_LIMIT_ENABLED", true),
		RateLimitRequests:        parseIntEnv("RATE_LIMIT_REQUESTS", defaultRateLimitRequests),
		RateLimitBurst:           parseIntEnv("RATE_LIMIT_BURST", defaultRateLimitBurst),
		RateLimitWindow:          parseDurationEnv("RATE_LIMIT_WINDOW", defaultRateLimitWindow),
		RateLimitCleanupInterval: parseDurationEnv("RATE_LIMIT_CLEANUP_INTERVAL", defaultRateLimitCleanup),
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

func parseBoolEnv(name string, fallback bool) bool {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return fallback
	}

	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}

	return parsed
}

func parseIntEnv(name string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}

	return parsed
}

func parseDurationEnv(name string, fallback time.Duration) time.Duration {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return fallback
	}

	parsed, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}

	return parsed
}
