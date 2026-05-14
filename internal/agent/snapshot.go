package agent

import (
	"context"
	"path/filepath"
	"slices"
	"strings"

	"github.com/electr1fy0/sorta/internal/core"
	"github.com/electr1fy0/sorta/internal/ops"
)

func SummarizeDirectory(ctx context.Context, dir string) (DirectorySnapshot, error) {
	const maxFiles = 120
	const maxFolders = 80

	observedFiles := make([]string, 0, maxFiles)
	observedFolders := make(map[string]struct{})
	extensionCounts := make(map[string]int)
	totalFiles := 0

	err := ops.WalkFilesWithIgnoreCtx(ctx, dir, nil, func(fileEntry core.FileEntry) error {
		totalFiles++
		rel, err := filepath.Rel(dir, fileEntry.SourcePath)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if len(observedFiles) < maxFiles {
			observedFiles = append(observedFiles, rel)
		}

		folder := filepath.ToSlash(filepath.Dir(rel))
		if folder != "." {
			observedFolders[folder] = struct{}{}
		}

		ext := strings.ToLower(filepath.Ext(rel))
		if ext == "" {
			ext = "(none)"
		}
		extensionCounts[ext]++
		return nil
	})

	if err != nil {
		return DirectorySnapshot{}, err
	}

	slices.Sort(observedFiles)

	folders := make([]string, 0, len(observedFolders))
	for folder := range observedFolders {
		folders = append(folders, folder)
	}
	slices.Sort(folders)
	if len(folders) > maxFolders {
		folders = folders[:maxFolders]
	}

	return DirectorySnapshot{
		FileCount:       totalFiles,
		FolderCount:     len(observedFolders),
		ObservedFiles:   observedFiles,
		ObservedFolders: folders,
		ExtensionCounts: extensionCounts,
	}, nil
}
