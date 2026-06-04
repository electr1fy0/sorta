package hash

import (
	"os"
	"path/filepath"
	"testing"
)

func TestHashCacheGetPut(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	cache, err := LoadHashCache()
	if err != nil {
		t.Fatal(err)
	}

	fp := FileFingerprint{Size: 100, ModTime: 12345, Inode: 99}

	// Should miss on empty cache
	_, ok := cache.Get("/some/path", fp)
	if ok {
		t.Error("expected cache miss on empty cache")
	}

	// Put and get
	cache.Put("/some/path", fp, "abc123")
	got, ok := cache.Get("/some/path", fp)
	if !ok {
		t.Fatal("expected cache hit after put")
	}
	if got != "abc123" {
		t.Errorf("got hash %q, want %q", got, "abc123")
	}
}

func TestHashCacheStaleFingerprint(t *testing.T) {
	cache := &HashCache{
		path:    filepath.Join(t.TempDir(), "cache.json"),
		entries: make(map[string]hashCacheEntry),
	}

	oldFp := FileFingerprint{Size: 100, ModTime: 1, Inode: 1}
	newFp := FileFingerprint{Size: 200, ModTime: 2, Inode: 2}

	cache.Put("/path", oldFp, "hash1")
	_, ok := cache.Get("/path", newFp)
	if ok {
		t.Error("expected cache miss when fingerprint changed")
	}
}

func TestHashCacheSave(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	cache, err := LoadHashCache()
	if err != nil {
		t.Fatal(err)
	}

	// No save needed when not dirty
	if err := cache.Save(); err != nil {
		t.Fatal(err)
	}

	cache.Put("/path", FileFingerprint{Size: 1, ModTime: 1, Inode: 1}, "hash1")

	if !cache.dirty {
		t.Error("expected cache to be dirty after put")
	}

	if err := cache.Save(); err != nil {
		t.Fatal(err)
	}

	if cache.dirty {
		t.Error("expected cache to be clean after save")
	}

	// Reload and verify
	cache2, err := LoadHashCache()
	if err != nil {
		t.Fatal(err)
	}
	got, ok := cache2.Get("/path", FileFingerprint{Size: 1, ModTime: 1, Inode: 1})
	if !ok || got != "hash1" {
		t.Errorf("after reload got (%q, %v), want (%q, true)", got, ok, "hash1")
	}
}

func TestLoadHashCacheNonExistent(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	cache, err := LoadHashCache()
	if err != nil {
		t.Fatal(err)
	}
	if cache == nil {
		t.Fatal("expected non-nil cache")
	}
}

func TestLoadHashCacheCorrupt(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	sortaDir := filepath.Join(home, ".sorta")
	os.MkdirAll(sortaDir, 0755)
	os.WriteFile(filepath.Join(sortaDir, "hash-cache.json"), []byte("not json"), 0644)

	cache, err := LoadHashCache()
	if err != nil {
		t.Fatal(err)
	}
	if cache == nil {
		t.Fatal("expected non-nil cache even with corrupt data")
	}
}
