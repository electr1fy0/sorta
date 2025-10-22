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

func FilterFiles(dir string, sorter Sorter, executor *Executor, reporter *Reporter) (*SortResult, error) {
	result := &SortResult{}
	var operations []FileOperation

	walkErr := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
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
		fileOp, err := sorter.Sort(dir, parentDir, d.Name(), size)
		if err != nil {
			return err
		}
		operations = append(operations, fileOp)
		return nil
	})

	if walkErr != nil {
		return nil, walkErr
	}

	for _, op := range operations {
		moved, err := executor.Execute(op)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error executing operation for %s: %v\n", op.Filename, err)
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
	for i := 0; i < limit; i++ {
		fmt.Printf("%d. %s (%s)\n", i+1, entries[i].Name, humanReadable(entries[i].Size))
	}

	return nil
}
