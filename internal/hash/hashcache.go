package hash

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"

	"github.com/electr1fy0/sorta/internal/core"
)

const hashCacheFilename = "hash-cache.json"

type hashCacheEntry struct {
	Fingerprint FileFingerprint `json:"fingerprint"`
	Hash        string          `json:"hash"`
}

type HashCache struct {
	mu      sync.Mutex
	path    string
	entries map[string]hashCacheEntry
	dirty   bool
}

func LoadHashCache() (*HashCache, error) {
	sortaDir, err := core.GetSortaDir()
	if err != nil {
		return nil, err
	}

	if err := os.MkdirAll(sortaDir, 0755); err != nil {
		return nil, err
	}

	cache := &HashCache{
		path:    filepath.Join(sortaDir, hashCacheFilename),
		entries: make(map[string]hashCacheEntry),
	}

	data, err := os.ReadFile(cache.path)
	if err != nil {
		if os.IsNotExist(err) {
			return cache, nil
		}
		return nil, err
	}
	if len(data) == 0 {
		return cache, nil
	}

	if err := json.Unmarshal(data, &cache.entries); err != nil {
		return cache, nil
	}

	return cache, nil
}

func (c *HashCache) Get(path string, fp FileFingerprint) (string, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry, ok := c.entries[path]
	if !ok {
		return "", false
	}
	if entry.Fingerprint != fp {
		return "", false
	}
	return entry.Hash, true
}

func (c *HashCache) Put(path string, fp FileFingerprint, hash string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	prev, ok := c.entries[path]
	if ok && prev.Fingerprint == fp && prev.Hash == hash {
		return
	}

	c.entries[path] = hashCacheEntry{
		Fingerprint: fp,
		Hash:        hash,
	}
	c.dirty = true
}

func (c *HashCache) Save() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.dirty {
		return nil
	}

	data, err := json.MarshalIndent(c.entries, "", "  ")
	if err != nil {
		return err
	}
	if err := core.WriteFileAtomic(c.path, data, 0644); err != nil {
		return err
	}
	c.dirty = false
	return nil
}
