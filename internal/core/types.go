package core

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"
)

const (
	ansiReset  = "\x1b[0m"
	ansiRed    = "\x1b[31m"
	ansiGreen  = "\x1b[32m"
	ansiYellow = "\x1b[33m"
)

type OperationType int

const (
	OpMove OperationType = iota
	OpRename
	OpDedupe
	OpDelete
	OpSkip
	OpUndo
)

type TransactionType int

const (
	TAction TransactionType = iota
	TUndo
)

type Transaction struct {
	ID           string
	TType        TransactionType
	Operations   []FileOperation
	Irreversible bool
}

type FileEntry struct {
	RootDir    string
	SourcePath string
	Size       int64
}

type FileOperation struct {
	OpType   OperationType
	File     FileEntry
	DestPath string
	Size     int64
}

type Sorter interface {
	Decide(ctx context.Context, filePaths []FileEntry) ([]FileOperation, error)
}

type DuplicateStats struct {
	FilesSeen      int
	PartialHashed  int
	FullHashed     int
	CacheHits      int
	CacheMisses    int
	BytesHashed    int64
	WalkDuration   time.Duration
	DecideDuration time.Duration
	TotalDuration  time.Duration
}

type ProgressEvent struct {
	Stage     string
	Completed int
	Total     int
}

type SortResult struct {
	Moved   int
	Renamed int
	Deduped int
	Skipped int
	Deleted int
	Errors  []error
}

func (r *SortResult) PrintSummary() {
	fmt.Println("--------------------------------------------------")
	if r.Moved > 0 {
		fmt.Printf("  %sMoved:%s %d\n", ansiGreen, ansiReset, r.Moved)
	} else if r.Deduped > 0 {
		fmt.Printf("  %sDeduped:%s %d\n", ansiGreen, ansiReset, r.Deduped)
	} else if r.Renamed > 0 {
		fmt.Printf("  %sRenamed:%s %d\n", ansiGreen, ansiReset, r.Renamed)
	}

	fmt.Printf("  %sDeleted:%s %d\n", ansiRed, ansiReset, r.Deleted)
	fmt.Printf("  %sSkipped:%s %d\n", ansiYellow, ansiReset, r.Skipped)
	if len(r.Errors) > 0 {
		fmt.Printf("  %sErrors:%s  %d\n", ansiRed, ansiReset, len(r.Errors))

		counts := make(map[string]int)
		examples := make(map[string]string)
		for _, err := range r.Errors {
			cat := classifyError(err)
			counts[cat]++
			if _, ok := examples[cat]; !ok {
				examples[cat] = err.Error()
			}
		}

		keys := make([]string, 0, len(counts))
		for k := range counts {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, k := range keys {
			fmt.Printf("    - %s: %d\n", k, counts[k])
			fmt.Printf("      e.g. %s\n", examples[k])
		}
	}
	fmt.Println("--------------------------------------------------")
}

func classifyError(err error) string {
	if errors.Is(err, os.ErrNotExist) {
		return "not_found"
	}
	if errors.Is(err, os.ErrPermission) {
		return "permission"
	}

	msg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(msg, "permission denied"):
		return "permission"
	case strings.Contains(msg, "file exists"):
		return "collision"
	case strings.Contains(msg, "no such file"):
		return "not_found"
	case strings.Contains(msg, "directory not empty"):
		return "dir_not_empty"
	case strings.Contains(msg, "invalid"):
		return "invalid"
	default:
		return "other"
	}
}
