package bench

import (
	"context"
	"time"

	"github.com/electr1fy0/sorta/internal/core"
	"github.com/electr1fy0/sorta/internal/dupl"
	"github.com/electr1fy0/sorta/internal/ignore"
	"github.com/electr1fy0/sorta/internal/ops"
)

type Report struct {
	Directory string
	Files     int
	Ops       int
	Dedupes   int
	Stats     core.DuplicateStats
}

func BenchmarkDuplicates(rootDir string) (*Report, error) {
	return BenchmarkDuplicatesCtx(context.Background(), rootDir, nil)
}

func BenchmarkDuplicatesCtx(
	ctx context.Context,
	rootDir string,
	progress func(core.ProgressEvent),
) (*Report, error) {
	start := time.Now()
	files := make([]core.FileEntry, 0, 1024)
	ignoreMatcher, err := ignore.LoadIgnoreMatcher(rootDir, nil)
	if err != nil {
		return nil, err
	}

	walkStart := time.Now()
	err = ops.WalkFilesWithIgnoreCtx(ctx, rootDir, ignoreMatcher, func(file core.FileEntry) error {
		files = append(files, file)
		return nil
	})
	if err != nil {
		return nil, err
	}
	walkDuration := time.Since(walkStart)

	finder := dupl.NewDuplicateFinder()
	finder.SetProgressReporter(progress)
	decideStart := time.Now()
	duplOps, err := finder.Decide(ctx, files)
	if err != nil {
		return nil, err
	}
	decideDuration := time.Since(decideStart)

	dedupes := 0
	for _, op := range duplOps {
		if op.OpType == core.OpDedupe {
			dedupes++
		}
	}

	stats := finder.Stats()
	stats.WalkDuration = walkDuration
	stats.DecideDuration = decideDuration
	stats.TotalDuration = time.Since(start)

	return &Report{
		Directory: rootDir,
		Files:     len(files),
		Ops:       len(duplOps),
		Dedupes:   dedupes,
		Stats:     stats,
	}, nil
}
