package views

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"

	_ "modernc.org/sqlite"
)

var usernamePattern = regexp.MustCompile(`^[a-z0-9_-]+$`)

type Store struct {
	db        *sql.DB
	closeOnce sync.Once
}

func NewStore(path string) (*Store, error) {
	if strings.TrimSpace(path) == "" {
		return nil, errors.New("view database path is required")
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

func (s *Store) init(ctx context.Context) error {
	statements := []string{
		`PRAGMA journal_mode = WAL;`,
		`CREATE TABLE IF NOT EXISTS profile_views (
			username TEXT NOT NULL,
			visitor_hash TEXT NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (username, visitor_hash)
		);`,
		`CREATE INDEX IF NOT EXISTS idx_profile_views_username ON profile_views (username);`,
	}

	for _, statement := range statements {
		if _, err := s.db.ExecContext(ctx, statement); err != nil {
			return err
		}
	}

	return nil
}

func (s *Store) Count(ctx context.Context, username string) (int, error) {
	normalized, err := NormalizeUsername(username)
	if err != nil {
		return 0, err
	}

	var count int
	if err := s.db.QueryRowContext(
		ctx,
		`SELECT COUNT(*) FROM profile_views WHERE username = ?`,
		normalized,
	).Scan(&count); err != nil {
		return 0, err
	}

	return count, nil
}

func (s *Store) Add(ctx context.Context, username, hash string) (bool, int, error) {
	normalizedUsername, err := NormalizeUsername(username)
	if err != nil {
		return false, 0, err
	}

	trimmedHash := strings.TrimSpace(hash)
	if trimmedHash == "" {
		count, countErr := s.Count(ctx, normalizedUsername)
		return false, count, countErr
	}

	result, err := s.db.ExecContext(
		ctx,
		`INSERT OR IGNORE INTO profile_views (username, visitor_hash) VALUES (?, ?)`,
		normalizedUsername,
		trimmedHash,
	)
	if err != nil {
		return false, 0, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, 0, err
	}

	count, err := s.Count(ctx, normalizedUsername)
	if err != nil {
		return false, 0, err
	}

	return rowsAffected > 0, count, nil
}

func (s *Store) HasAny(ctx context.Context, username string, hashes ...string) (bool, error) {
	normalizedUsername, err := NormalizeUsername(username)
	if err != nil {
		return false, err
	}

	uniqueHashes := sanitizeHashes(hashes)
	if len(uniqueHashes) == 0 {
		return false, nil
	}

	placeholders := make([]string, 0, len(uniqueHashes))
	args := make([]any, 0, len(uniqueHashes)+1)
	args = append(args, normalizedUsername)

	for index, hash := range uniqueHashes {
		placeholders = append(placeholders, "?"+strconv.Itoa(index+2))
		args = append(args, hash)
	}

	query := `SELECT EXISTS(
		SELECT 1
		FROM profile_views
		WHERE username = ?1
		  AND visitor_hash IN (` + strings.Join(placeholders, ", ") + `)
	)`

	var exists bool
	if err := s.db.QueryRowContext(ctx, query, args...).Scan(&exists); err != nil {
		return false, err
	}

	return exists, nil
}

func (s *Store) ImportLegacy(ctx context.Context, username string, hashes []string) (int, int, error) {
	normalizedUsername, err := NormalizeUsername(username)
	if err != nil {
		return 0, 0, err
	}

	legacyHashes := sanitizeHashes(hashes)
	if len(legacyHashes) == 0 {
		count, countErr := s.Count(ctx, normalizedUsername)
		return 0, count, countErr
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, 0, err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	statement, err := tx.PrepareContext(
		ctx,
		`INSERT OR IGNORE INTO profile_views (username, visitor_hash) VALUES (?, ?)`,
	)
	if err != nil {
		return 0, 0, err
	}
	defer func() {
		_ = statement.Close()
	}()

	imported := 0
	for _, hash := range legacyHashes {
		result, execErr := statement.ExecContext(ctx, normalizedUsername, hash)
		if execErr != nil {
			return 0, 0, execErr
		}

		rowsAffected, rowsErr := result.RowsAffected()
		if rowsErr != nil {
			return 0, 0, rowsErr
		}

		imported += int(rowsAffected)
	}

	count, err := countViews(ctx, tx, normalizedUsername)
	if err != nil {
		return 0, 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, 0, err
	}

	return imported, count, nil
}

func NormalizeUsername(username string) (string, error) {
	normalized := strings.TrimSpace(strings.ToLower(username))
	if normalized == "" {
		return "", errors.New("username is required")
	}

	if !usernamePattern.MatchString(normalized) {
		return "", errors.New("username must use only letters, numbers, hyphens, or underscores")
	}

	return normalized, nil
}

func HashIP(ip, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(ip))
	return hex.EncodeToString(mac.Sum(nil))
}

func HashViewer(username, ip, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(strings.ToLower(strings.TrimSpace(username))))
	_, _ = mac.Write([]byte{':'})
	_, _ = mac.Write([]byte(ip))
	return hex.EncodeToString(mac.Sum(nil))
}

func (s *Store) Close() error {
	var err error
	s.closeOnce.Do(func() {
		err = s.db.Close()
	})
	return err
}

func countViews(ctx context.Context, querier interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}, username string) (int, error) {
	var count int
	if err := querier.QueryRowContext(
		ctx,
		`SELECT COUNT(*) FROM profile_views WHERE username = ?`,
		username,
	).Scan(&count); err != nil {
		return 0, err
	}

	return count, nil
}

func sanitizeHashes(hashes []string) []string {
	seen := make(map[string]struct{}, len(hashes))
	sanitized := make([]string, 0, len(hashes))

	for _, hash := range hashes {
		trimmed := strings.TrimSpace(hash)
		if trimmed == "" {
			continue
		}

		if _, exists := seen[trimmed]; exists {
			continue
		}

		seen[trimmed] = struct{}{}
		sanitized = append(sanitized, trimmed)
	}

	return sanitized
}
