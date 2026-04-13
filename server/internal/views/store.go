package views

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type Store struct {
	mu     sync.Mutex
	path   string
	hashes map[string]struct{}
}

type persisted struct {
	Hashes []string `json:"hashes"`
}

func NewStore(path string) (*Store, error) {
	store := &Store{
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

	var payload persisted
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, err
	}

	for _, hash := range payload.Hashes {
		trimmed := strings.TrimSpace(hash)
		if trimmed == "" {
			continue
		}
		store.hashes[trimmed] = struct{}{}
	}

	return store, nil
}

func (s *Store) Count() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	return len(s.hashes)
}

func (s *Store) Add(hash string) (bool, int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.hashes[hash]; exists {
		return false, len(s.hashes), nil
	}

	s.hashes[hash] = struct{}{}
	if err := s.persistLocked(); err != nil {
		delete(s.hashes, hash)
		return false, len(s.hashes), err
	}

	return true, len(s.hashes), nil
}

func HashIP(ip, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(ip))
	return hex.EncodeToString(mac.Sum(nil))
}

func (s *Store) persistLocked() error {
	hashes := make([]string, 0, len(s.hashes))
	for hash := range s.hashes {
		hashes = append(hashes, hash)
	}

	payload := persisted{Hashes: hashes}
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	tempPath := s.path + ".tmp"
	if err := os.WriteFile(tempPath, data, 0o600); err != nil {
		return err
	}

	return os.Rename(tempPath, s.path)
}
