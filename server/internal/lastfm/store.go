package lastfm

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

var usernamePattern = regexp.MustCompile(`^[a-z0-9._-]+$`)

type CacheEntry struct {
	Track     *NowPlaying
	FetchedAt time.Time
}

type Store struct {
	db        *sql.DB
	closeOnce sync.Once
}

func NewStore(path string) (*Store, error) {
	if strings.TrimSpace(path) == "" {
		return nil, errors.New("lastfm database path is required")
	}

	if err := ensureParentDir(path); err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}

	store := &Store{db: db}
	if err := store.init(context.Background()); err != nil {
		_ = db.Close()
		return nil, err
	}

	return store, nil
}

func NormalizeUsername(username string) (string, error) {
	normalized := strings.TrimSpace(strings.ToLower(username))
	if normalized == "" {
		return "", errors.New("last.fm username is required")
	}

	if !usernamePattern.MatchString(normalized) {
		return "", errors.New("last.fm username must use only letters, numbers, dots, hyphens, or underscores")
	}

	return normalized, nil
}

func (s *Store) Get(ctx context.Context, username string) (*CacheEntry, error) {
	normalized, err := NormalizeUsername(username)
	if err != nil {
		return nil, err
	}

	var payload sql.NullString
	var fetchedAtUnix int64
	err = s.db.QueryRowContext(
		ctx,
		`SELECT payload, fetched_at_unix FROM lastfm_cache WHERE username = ?`,
		normalized,
	).Scan(&payload, &fetchedAtUnix)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var track *NowPlaying
	if payload.Valid && strings.TrimSpace(payload.String) != "" {
		var decoded NowPlaying
		if err := json.Unmarshal([]byte(payload.String), &decoded); err != nil {
			return nil, err
		}
		track = &decoded
	}

	return &CacheEntry{
		Track:     track,
		FetchedAt: time.Unix(fetchedAtUnix, 0).UTC(),
	}, nil
}

func (s *Store) Upsert(ctx context.Context, username string, track *NowPlaying, fetchedAt time.Time) error {
	normalized, err := NormalizeUsername(username)
	if err != nil {
		return err
	}

	var payload any
	if track != nil {
		encoded, err := json.Marshal(track)
		if err != nil {
			return err
		}
		payload = string(encoded)
	}

	_, err = s.db.ExecContext(
		ctx,
		`INSERT INTO lastfm_cache (username, payload, fetched_at_unix)
		VALUES (?, ?, ?)
		ON CONFLICT(username) DO UPDATE SET
			payload = excluded.payload,
			fetched_at_unix = excluded.fetched_at_unix`,
		normalized,
		payload,
		fetchedAt.UTC().Unix(),
	)
	return err
}

func (s *Store) Close() error {
	var err error
	s.closeOnce.Do(func() {
		err = s.db.Close()
	})
	return err
}

func (s *Store) init(ctx context.Context) error {
	statements := []string{
		`PRAGMA journal_mode = WAL;`,
		`CREATE TABLE IF NOT EXISTS lastfm_cache (
			username TEXT NOT NULL PRIMARY KEY,
			payload TEXT,
			fetched_at_unix INTEGER NOT NULL
		);`,
	}

	for _, statement := range statements {
		if _, err := s.db.ExecContext(ctx, statement); err != nil {
			return err
		}
	}

	return nil
}

func ensureParentDir(path string) error {
	if path == ":memory:" || strings.HasPrefix(path, "file:") {
		return nil
	}

	parent := filepath.Dir(path)
	if parent == "." || parent == "" {
		return nil
	}

	return os.MkdirAll(parent, 0o755)
}
