package views

import (
	"context"
	"path/filepath"
	"testing"
)

func TestStoreCountsViewsPerUsername(t *testing.T) {
	t.Parallel()

	store, err := NewStore(filepath.Join(t.TempDir(), "views.db"))
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}
	defer func() {
		if err := store.Close(); err != nil {
			t.Fatalf("Close() error = %v", err)
		}
	}()

	ctx := context.Background()

	added, count, err := store.Add(ctx, "kisakay", "hash-a")
	if err != nil {
		t.Fatalf("Add(kisakay, hash-a) error = %v", err)
	}

	if !added || count != 1 {
		t.Fatalf("expected first insert to increment count, got added=%v count=%d", added, count)
	}

	added, count, err = store.Add(ctx, "kisakay", "hash-a")
	if err != nil {
		t.Fatalf("Add duplicate error = %v", err)
	}

	if added || count != 1 {
		t.Fatalf("expected duplicate insert to be ignored, got added=%v count=%d", added, count)
	}

	added, count, err = store.Add(ctx, "another-user", "hash-a")
	if err != nil {
		t.Fatalf("Add(another-user, hash-a) error = %v", err)
	}

	if !added || count != 1 {
		t.Fatalf("expected per-user counter isolation, got added=%v count=%d", added, count)
	}

	kisakayCount, err := store.Count(ctx, "kisakay")
	if err != nil {
		t.Fatalf("Count(kisakay) error = %v", err)
	}

	if kisakayCount != 1 {
		t.Fatalf("expected kisakay count = 1, got %d", kisakayCount)
	}
}

func TestNormalizeUsername(t *testing.T) {
	t.Parallel()

	got, err := NormalizeUsername("  KiSaKaY_123  ")
	if err != nil {
		t.Fatalf("NormalizeUsername() error = %v", err)
	}

	if got != "kisakay_123" {
		t.Fatalf("expected normalized username to be kisakay_123, got %q", got)
	}

	if _, err := NormalizeUsername("not valid!"); err == nil {
		t.Fatal("expected invalid username to return an error")
	}
}

func TestHashViewerIncludesUsername(t *testing.T) {
	t.Parallel()

	hashA := HashViewer("kisakay", "127.0.0.1", "secret")
	hashB := HashViewer("another-user", "127.0.0.1", "secret")

	if hashA == hashB {
		t.Fatal("expected username-scoped hashes to differ")
	}
}

func TestStoreImportLegacyAndHasAny(t *testing.T) {
	t.Parallel()

	store, err := NewStore(filepath.Join(t.TempDir(), "views.db"))
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}
	defer func() {
		if err := store.Close(); err != nil {
			t.Fatalf("Close() error = %v", err)
		}
	}()

	ctx := context.Background()
	legacyHash := HashIP("127.0.0.1", "secret")

	imported, count, err := store.ImportLegacy(ctx, "kisakay", []string{legacyHash, legacyHash})
	if err != nil {
		t.Fatalf("ImportLegacy() error = %v", err)
	}

	if imported != 1 || count != 1 {
		t.Fatalf("expected imported=1 count=1, got imported=%d count=%d", imported, count)
	}

	exists, err := store.HasAny(ctx, "kisakay", legacyHash, HashViewer("kisakay", "127.0.0.1", "secret"))
	if err != nil {
		t.Fatalf("HasAny() error = %v", err)
	}

	if !exists {
		t.Fatal("expected legacy hash lookup to succeed")
	}
}
