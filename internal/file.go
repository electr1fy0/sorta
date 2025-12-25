package internal

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"
)

var (
	DuplNuke         = false
	RecurseLevel int = -1
)

func FilterFiles(rootDir string, sorter Sorter, executor *Executor, reporter *Reporter) (*SortResult, error) {
	operations, err := PlanOperations(rootDir, sorter)
	if err != nil {
		return nil, err
	}

	return ApplyOperations(rootDir, operations, executor, reporter)
}

func PlanOperations(rootDir string, sorter Sorter) ([]FileOperation, error) {
	var files []FileEntry

	if RecurseLevel >= 0 && runtime.GOOS == "windows" {
		return nil, fmt.Errorf("--recurselevel is currently only available on unix")
	}

	walkErr := WalkFiles(rootDir, func(file FileEntry) error {
		files = append(files, file)
		return nil
	})

	if walkErr != nil {
		return nil, walkErr
	}

	operations, err := sorter.Decide(files)
	if err != nil {
		return nil, err
	}
	return operations, nil
}

func ApplyOperations(rootDir string, operations []FileOperation, executor *Executor, reporter *Reporter) (*SortResult, error) {
	result := &SortResult{}
	for _, op := range operations {
		moved, err := executor.Execute(op)
		if moved || err != nil {
			reporter.Report(op, err)
		}

		if err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("%s: %w", filepath.Base(op.File.SourcePath), err))
			continue
		}
		if moved {
			switch op.OpType {
			case OpMove:
				result.Moved++
			case OpDelete:
				result.Deleted++
			}
		}
		if op.OpType == OpSkip {
			result.Skipped++
		}
	}
	id := time.Now().String()
	transaction := Transaction{TType: TAction, Operations: operations, ID: id}
	LogToHistory(transaction)
	if err := cleanEmptyFolders(rootDir); err != nil {
		return result, err
	}

	if DuplNuke {
		if err := os.RemoveAll(filepath.Join(rootDir, "duplicates")); err != nil {
			return result, err
		}
		result.Deleted++
	}
	return result, nil
}

func cleanEmptyFolders(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}

	for _, entry := range entries {
		path := filepath.Join(dir, entry.Name())

		if entry.IsDir() {
			if err := cleanEmptyFolders(path); err != nil {
				return err
			}

			subEntries, err := os.ReadDir(path)
			if err != nil {
				continue
			}

			if len(subEntries) == 0 {

				if err := os.Remove(path); err != nil && !os.IsNotExist(err) {

					return fmt.Errorf("failed to remove empty dir, %q: %w", path, err)
				}
			} else if len(subEntries) == 1 && subEntries[0].Name() == ".DS_Store" {
				err := os.Remove(filepath.Join(path, ".DS_Store"))
				if err != nil {
					return fmt.Errorf("failed to remove .DS_Store: %v", err)
				}
				if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
					return fmt.Errorf("failed to remove empty dir, %q: %w", path, err)
				}

			}
		}
	}

	return nil
}

func TopLargestFiles(rootDir string, n int) error {
	var entries []FileEntry
	err := WalkFiles(rootDir, func(file FileEntry) error {
		entries = append(entries, file)
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
	fmt.Printf("Top %d largest files in %s:\n", limit, rootDir)
	return nil
}

func WalkFiles(rootDir string, fn func(FileEntry) error) error {
	return filepath.WalkDir(rootDir, func(path string, d fs.DirEntry, err error) error {
		relFolder, _ := filepath.Rel(rootDir, filepath.Dir(path))
		relFolder = filepath.Clean(relFolder)
		slashCnt := strings.Count(relFolder, "/")
		if RecurseLevel >= 0 && slashCnt > RecurseLevel {
			return nil
		}
		if err != nil {
			return err
		}
		if strings.Contains(path, "/.") || strings.Contains(path, "\\.") || d.IsDir() {
			return nil
		}

		stat, err := d.Info()
		if err != nil {
			return err
		}

		size := stat.Size()
		return fn(FileEntry{rootDir, path, size})
	})
}
