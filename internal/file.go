package internal

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

var DuplNuke = false
var RecurseLevel int = -1

type FilePath struct {
	BaseDir  string `json:"BaseDir"`
	FullDir  string `json:"FullDir"`
	Filename string `json:"Filename"`
	Size     int64  `json:"size"`
}

func FilterFiles(dir string, sorter Sorter, executor *Executor, reporter *Reporter) (*SortResult, error) {
	result := &SortResult{}
	var operations []FileOperation
	var filePaths []FilePath
	walkErr := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		relPath, _ := filepath.Rel(dir, filepath.Dir(path))
		relPath = filepath.Clean(relPath)
		fmt.Println("relpath", relPath)
		slashCnt := strings.Count(relPath, "/")
		if RecurseLevel >= 0 && slashCnt > RecurseLevel {
			return nil
		}
		if err != nil {
			return err
		}

		if d.IsDir() || strings.HasPrefix(d.Name(), ".") {
			return nil
		}
		stat, err := d.Info()
		if err != nil {
			return err
		}

		size := stat.Size()
		parentDir := filepath.Dir(path)

		filePaths = append(filePaths, FilePath{dir, parentDir, d.Name(), size})
		return nil
	})

	if walkErr != nil {
		return nil, walkErr
	}

	operations, _ = sorter.Sort(filePaths)

	for _, op := range operations {
		moved, err := executor.Execute(op)
		if moved || err != nil {
			reporter.Report(op, err)
		}

		if err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("%s: %w", op.Filename, err))
			continue
		}
		if moved {
			result.Moved++
		}
		if op.Type == OpSkip {
			result.Skipped++
		}
	}

	if err := cleanEmptyFolders(dir); err != nil {
		return nil, err
	}

	if DuplNuke {
		if err := os.RemoveAll(filepath.Join(dir, "duplicates")); err != nil {
			return nil, err
		}
	}
	return result, nil
}

func cleanEmptyFolders(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			dirPath := filepath.Join(dir, entry.Name())
			f, err := os.Open(dirPath)
			if err != nil {
				return fmt.Errorf("failed to open directory: %w", err)
			}
			defer f.Close()

			_, err = f.Readdir(1)
			if err == io.EOF {
				if err := os.Remove(dirPath); err != nil {
					return fmt.Errorf("failed to remove empty directory: %w", err)
				}
			} else if err != nil {
				return fmt.Errorf("failed to read directory contents: %w", err)
			}
		}
	}
	return nil
}

func TopLargestFiles(dir string, n int) error {
	var entries []FileInfo
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || strings.HasPrefix(d.Name(), ".") {
			return nil
		}
		f, err := d.Info()
		if err != nil {
			return err
		}
		entries = append(entries, FileInfo{d.Name(), f.Size()})
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to walk directory: %w", err)
	}

	if len(entries) == 0 {
		fmt.Println("No files found.")
		return nil
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Size > entries[j].Size
	})

	limit := min(len(entries), n)
	fmt.Printf("Top %d largest files in %s:\n", limit, dir)
	for i := range limit {
		fmt.Printf("%d. %s (%s)\n", i+1, entries[i].Name, humanReadable(entries[i].Size))
	}

	return nil
}
