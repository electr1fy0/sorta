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
	result := &SortResult{}
	var operations []FileOperation
	var files []FileEntry

	if RecurseLevel >= 0 && runtime.GOOS == "windows" {
		return result, fmt.Errorf("--recurselevel is only available on Unix")
	}

	walkErr := filepath.WalkDir(rootDir, func(path string, d fs.DirEntry, err error) error {
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

		files = append(files, FileEntry{rootDir, path, size})
		return nil
	})

	if walkErr != nil {
		return nil, walkErr
	}

	operations, _ = sorter.Decide(files)

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
			result.Moved++
		}
		if op.OpType == OpSkip {
			result.Skipped++
		}
	}
	id := time.Now().String()
	transaction := Transaction{TType: TAction, Operations: operations, ID: id}
	logToHistory(transaction)
	if err := cleanEmptyFolders(rootDir); err != nil {
		return nil, err
	}

	if DuplNuke {
		if err := os.RemoveAll(filepath.Join(rootDir, "duplicates")); err != nil {
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
	err := filepath.WalkDir(rootDir, func(path string, d fs.DirEntry, err error) error {
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
		entries = append(entries, FileEntry{rootDir, path, f.Size()})
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
	for i := range limit {
		fmt.Printf("%d. %s (%s)\n", i+1, filepath.Base(entries[i].SourcePath), humanReadable(entries[i].Size))
	}

	return nil
}
