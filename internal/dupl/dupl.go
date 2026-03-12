package dupl

import (
	"context"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/electr1fy0/sorta/internal/core"
	"github.com/electr1fy0/sorta/internal/hash"
)

type sizeHash struct {
	size int64
	hash string
}

type DuplicateFinder struct {
	cache      *hash.HashCache
	stats      core.DuplicateStats
	statsMu    sync.Mutex
	progressFn func(core.ProgressEvent)
}

func NewDuplicateFinder() *DuplicateFinder {
	cache, err := hash.LoadHashCache()
	if err != nil {
		cache = nil
	}
	return &DuplicateFinder{cache: cache}
}

func (d *DuplicateFinder) SetProgressReporter(fn func(core.ProgressEvent)) {
	d.statsMu.Lock()
	defer d.statsMu.Unlock()
	d.progressFn = fn
}

func (d *DuplicateFinder) Decide(ctx context.Context, files []core.FileEntry) ([]core.FileOperation, error) {
	d.setFilesSeen(len(files))

	validFiles, ops := filterValidFiles(files)

	bySize := groupBySize(validFiles)
	partialCandidates := filterSingletons(bySize, &ops)

	partialHashes, err := d.hashFiles(ctx, partialCandidates, "partial", func(f core.FileEntry) (string, error) {
		h, err := partialHash(f.SourcePath)
		if err == nil {
			d.addPartialHashed(1)
		}
		return h, err
	})
	if err != nil {
		return nil, err
	}

	byPartial, err := groupByHashCtx(ctx, partialCandidates, partialHashes, func(f core.FileEntry, h string) sizeHash {
		return sizeHash{f.Size, h}
	})
	if err != nil {
		return nil, err
	}
	fullCandidates := filterSingletons(byPartial, &ops)

	fullHashes, err := d.hashFiles(ctx, fullCandidates, "full", d.fullHashWithCache)
	if err != nil {
		return nil, err
	}

	byFull, err := groupByHashCtx(ctx, fullCandidates, fullHashes, func(_ core.FileEntry, h string) string { return h })
	if err != nil {
		return nil, err
	}

	for contentHash, fGroup := range byFull {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		ops = append(ops, core.FileOperation{OpType: core.OpSkip})
		for _, f := range fGroup[1:] {
			ops = append(ops, core.FileOperation{
				OpType:   core.OpDedupe,
				File:     f,
				DestPath: dedupeDestPath(f, contentHash),
			})
		}
	}

	if d.cache != nil {
		if err := d.cache.Save(); err != nil {
			return nil, err
		}
	}

	return ops, nil
}

func filterSingletons[K comparable](groups map[K][]core.FileEntry, ops *[]core.FileOperation) []core.FileEntry {
	var candidates []core.FileEntry
	for _, group := range groups {
		if len(group) == 1 {
			*ops = append(*ops, core.FileOperation{OpType: core.OpSkip})
			continue
		}
		candidates = append(candidates, group...)
	}
	return candidates
}

func groupBySize(files []core.FileEntry) map[int64][]core.FileEntry {
	result := make(map[int64][]core.FileEntry, len(files))
	for _, f := range files {
		result[f.Size] = append(result[f.Size], f)
	}
	return result
}

func groupByHashCtx[K comparable](ctx context.Context, files []core.FileEntry, hashes map[string]string, keyFn func(core.FileEntry, string) K) (map[K][]core.FileEntry, error) {
	result := make(map[K][]core.FileEntry, len(files))
	for _, f := range files {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		h, ok := hashes[f.SourcePath]
		if !ok {
			return nil, fmt.Errorf("hash missing for %s", f.SourcePath)
		}
		k := keyFn(f, h)
		result[k] = append(result[k], f)
	}
	return result, nil
}

func filterValidFiles(files []core.FileEntry) ([]core.FileEntry, []core.FileOperation) {
	var valid []core.FileEntry
	var ops []core.FileOperation
	for _, f := range files {
		if f.SourcePath == filepath.Join(f.RootDir, "duplicates", filepath.Base(f.SourcePath)) {
			ops = append(ops, core.FileOperation{OpType: core.OpSkip})
			continue
		}
		valid = append(valid, f)
	}
	return valid, ops
}

func (d *DuplicateFinder) Stats() core.DuplicateStats {
	d.statsMu.Lock()
	defer d.statsMu.Unlock()
	return d.stats
}

func (d *DuplicateFinder) reportProgress(event core.ProgressEvent) {
	d.statsMu.Lock()
	fn := d.progressFn
	d.statsMu.Unlock()
	if fn != nil {
		fn(event)
	}
}

func (d *DuplicateFinder) setFilesSeen(n int) {
	d.statsMu.Lock()
	defer d.statsMu.Unlock()
	d.stats.FilesSeen = n
}

func (d *DuplicateFinder) addPartialHashed(n int) {
	d.statsMu.Lock()
	defer d.statsMu.Unlock()
	d.stats.PartialHashed += n
}

func (d *DuplicateFinder) addCacheHit() {
	d.statsMu.Lock()
	defer d.statsMu.Unlock()
	d.stats.CacheHits++
}

func (d *DuplicateFinder) addFullHashMiss(bytes int64) {
	d.statsMu.Lock()
	defer d.statsMu.Unlock()
	d.stats.CacheMisses++
	d.stats.FullHashed++
	d.stats.BytesHashed += bytes
}

func (d *DuplicateFinder) fullHashWithCache(file core.FileEntry) (string, error) {
	fp, err := hash.GetFingerprint(file.SourcePath)
	if err != nil {
		return "", err
	}

	if d.cache != nil {
		if h, ok := d.cache.Get(file.SourcePath, fp); ok {
			d.addCacheHit()
			return h, nil
		}
	}

	h, err := fullHash(file.SourcePath)
	if err != nil {
		return "", err
	}

	d.addFullHashMiss(file.Size)

	if d.cache != nil {
		d.cache.Put(file.SourcePath, fp, h)
	}
	return h, nil
}

func hashWorkerCount() int {
	workers := runtime.NumCPU() * 2
	workers = max(2, workers)
	workers = min(16, workers)

	return workers
}

func (d *DuplicateFinder) hashFiles(
	ctx context.Context,
	files []core.FileEntry,
	stage string,
	hashFn func(core.FileEntry) (string, error),
) (map[string]string, error) {
	if len(files) == 0 {
		return map[string]string{}, nil
	}

	type hashResult struct {
		path string
		hash string
		err  error
	}

	workers := hashWorkerCount()
	jobs := make(chan core.FileEntry, workers*2)
	results := make(chan hashResult, workers*2)

	var wg sync.WaitGroup
	wg.Add(workers)
	for range workers {
		go func() {
			defer wg.Done()
			for f := range jobs {
				if err := ctx.Err(); err != nil {
					results <- hashResult{path: f.SourcePath, err: err}
					continue
				}
				h, err := hashFn(f)
				results <- hashResult{path: f.SourcePath, hash: h, err: err}
			}
		}()
	}

	go func() {
		for _, f := range files {
			if ctx.Err() != nil {
				break
			}
			jobs <- f
		}
		close(jobs)
		wg.Wait()
		close(results)
	}()

	hashed := make(map[string]string, len(files))
	var firstErr error
	completed, lastReported := 0, 0
	for res := range results {
		completed++
		if completed-lastReported >= 250 || completed == len(files) {
			d.reportProgress(core.ProgressEvent{Stage: stage, Completed: completed, Total: len(files)})
			lastReported = completed
		}
		if res.err != nil {
			if firstErr == nil {
				firstErr = res.err
			}
			continue
		}
		hashed[res.path] = res.hash
	}

	if firstErr != nil {
		return nil, firstErr
	}
	return hashed, nil
}

func dedupeDestPath(file core.FileEntry, contentHash string) string {
	base := filepath.Base(file.SourcePath)
	ext := filepath.Ext(base)
	stem := base[:len(base)-len(ext)]

	hashPart := contentHash
	if len(hashPart) > 8 {
		hashPart = hashPart[:8]
	}

	pathHash := sha1.Sum([]byte(filepath.Clean(file.SourcePath)))
	pathPart := hex.EncodeToString(pathHash[:])[:6]

	return filepath.Join(file.RootDir, "duplicates", fmt.Sprintf("%s_%s_%s%s", stem, hashPart, pathPart, ext))
}

func partialHash(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	buf := make([]byte, 4096)
	n, err := io.ReadFull(f, buf)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return "", err
	}
	return fmt.Sprintf("%x", sha256.Sum256(buf[:n])), nil
}

func fullHash(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}
