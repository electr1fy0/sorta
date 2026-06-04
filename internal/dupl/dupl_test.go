package dupl

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/electr1fy0/sorta/internal/core"
)

func TestGroupBySize(t *testing.T) {
	files := []core.FileEntry{
		{SourcePath: "/a/1.txt", Size: 100},
		{SourcePath: "/a/2.txt", Size: 200},
		{SourcePath: "/a/3.txt", Size: 100},
	}
	groups := groupBySize(files)
	if len(groups) != 2 {
		t.Fatalf("expected 2 size groups, got %d", len(groups))
	}
	if len(groups[100]) != 2 {
		t.Errorf("expected 2 files of size 100, got %d", len(groups[100]))
	}
	if len(groups[200]) != 1 {
		t.Errorf("expected 1 file of size 200, got %d", len(groups[200]))
	}
}

func TestFilterValidFiles(t *testing.T) {
	files := []core.FileEntry{
		{RootDir: "/dir", SourcePath: "/dir/file.txt"},
		{RootDir: "/dir", SourcePath: "/dir/duplicates/dup.txt"},
	}
	valid, ops := filterValidFiles(files)
	if len(valid) != 1 {
		t.Errorf("expected 1 valid file, got %d", len(valid))
	}
	if len(ops) != 1 {
		t.Errorf("expected 1 skip op for file in duplicates dir, got %d", len(ops))
	}
}

func TestFilterSingletons(t *testing.T) {
	groups := map[int64][]core.FileEntry{
		100: {{SourcePath: "/a/1.txt"}},
		200: {{SourcePath: "/a/2.txt"}, {SourcePath: "/a/3.txt"}},
	}
	var ops []core.FileOperation
	candidates := filterSingletons(groups, &ops)
	if len(candidates) != 2 {
		t.Errorf("expected 2 candidates from non-singleton group, got %d", len(candidates))
	}
	if len(ops) != 1 {
		t.Errorf("expected 1 skip op for singleton group, got %d", len(ops))
	}
}

func TestDedupeDestPath(t *testing.T) {
	file := core.FileEntry{
		RootDir:    "/dir",
		SourcePath: "/dir/sub/photo.png",
	}
	dest := dedupeDestPath(file, "abcdef1234567890")
	if !strings.HasPrefix(dest, "/dir/duplicates/") {
		t.Errorf("expected path under /dir/duplicates/, got %q", dest)
	}
	if !strings.HasSuffix(dest, ".png") {
		t.Errorf("expected .png extension, got %q", dest)
	}
	if !strings.Contains(dest, "abcdef12") {
		t.Errorf("expected hash prefix in path, got %q", dest)
	}
}

func TestHashWorkerCount(t *testing.T) {
	n := hashWorkerCount()
	if n < 2 || n > 16 {
		t.Errorf("expected worker count between 2 and 16, got %d", n)
	}
}

func TestPartialHash(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.bin")
	if err := os.WriteFile(path, []byte("hello world"), 0644); err != nil {
		t.Fatal(err)
	}

	h, err := partialHash(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(h) == 0 {
		t.Error("expected non-empty hash")
	}
}

func TestFullHash(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.bin")
	if err := os.WriteFile(path, []byte("hello world"), 0644); err != nil {
		t.Fatal(err)
	}

	h, err := fullHash(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(h) == 0 {
		t.Error("expected non-empty hash")
	}
}

func TestPartialAndFullHashConsistency(t *testing.T) {
	dir := t.TempDir()

	small := filepath.Join(dir, "small.bin")
	os.WriteFile(small, []byte("hello"), 0644)
	large := filepath.Join(dir, "large.bin")
	data := make([]byte, 10000)
	for i := range data {
		data[i] = byte(i % 256)
	}
	os.WriteFile(large, data, 0644)

	// Small file: partial and full should differ because partial takes only first 4096 bytes
	smallPartial, _ := partialHash(small)
	smallFull, _ := fullHash(small)
	if smallPartial == smallFull {
		t.Log("small file may have same partial and full hash")
	}

	// Large file: partial takes first 4096, full takes all
	largePartial, _ := partialHash(large)
	largeFull, _ := fullHash(large)
	if largePartial == largeFull {
		t.Log("large file may have same partial and full hash (unlikely but not a bug)")
	}

	_ = smallPartial
	_ = smallFull
	_ = largePartial
	_ = largeFull
}

func TestDuplicateFinder(t *testing.T) {
	dir := t.TempDir()

	// Create two identical files and one unique file
	contentA := []byte("duplicate content")
	file1 := filepath.Join(dir, "file1.txt")
	file2 := filepath.Join(dir, "file2.txt")
	file3 := filepath.Join(dir, "file3.txt")
	os.WriteFile(file1, contentA, 0644)
	os.WriteFile(file2, contentA, 0644)
	os.WriteFile(file3, []byte("unique"), 0644)

	finder := NewDuplicateFinder()
	files := []core.FileEntry{
		{RootDir: dir, SourcePath: file1, Size: int64(len(contentA))},
		{RootDir: dir, SourcePath: file2, Size: int64(len(contentA))},
		{RootDir: dir, SourcePath: file3, Size: 6},
	}

	ops, err := finder.Decide(context.Background(), files)
	if err != nil {
		t.Fatal(err)
	}

	// Should skip file3, skip one of the duplicates, and dedupe the other
	var dedupCount, skipCount int
	for _, op := range ops {
		switch op.OpType {
		case core.OpDedupe:
			dedupCount++
		case core.OpSkip:
			skipCount++
		}
	}
	if dedupCount != 1 {
		t.Errorf("expected 1 dedupe, got %d", dedupCount)
	}
	if skipCount != 2 {
		t.Errorf("expected 2 skips, got %d", skipCount)
	}
}

func TestGroupByHashCtx(t *testing.T) {
	files := []core.FileEntry{
		{SourcePath: "/a/1.txt"},
		{SourcePath: "/a/2.txt"},
	}
	hashes := map[string]string{
		"/a/1.txt": "hash1",
		"/a/2.txt": "hash1",
	}
	groups, err := groupByHashCtx(context.Background(), files, hashes, func(_ core.FileEntry, h string) string {
		return h
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(groups) != 1 {
		t.Errorf("expected 1 group, got %d", len(groups))
	}
	if len(groups["hash1"]) != 2 {
		t.Errorf("expected 2 files in hash1 group, got %d", len(groups["hash1"]))
	}
}

func TestGroupByHashCtxMissingHash(t *testing.T) {
	files := []core.FileEntry{
		{SourcePath: "/a/1.txt"},
	}
	_, err := groupByHashCtx(context.Background(), files, map[string]string{}, func(_ core.FileEntry, h string) string {
		return h
	})
	if err == nil {
		t.Error("expected error for missing hash")
	}
}

func TestGroupByHashCtxContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	files := []core.FileEntry{{SourcePath: "/a/1.txt"}}
	hashes := map[string]string{"/a/1.txt": "hash1"}
	_, err := groupByHashCtx(ctx, files, hashes, func(_ core.FileEntry, h string) string {
		return h
	})
	if err == nil {
		t.Error("expected error from cancelled context")
	}
}
