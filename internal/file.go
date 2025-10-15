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

func FilterFiles(dir string, sorter Sorter, executor *Executor, reporter *Reporter) (SortResult, error) {
	result := SortResult{}

	walkErr := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			result.Errors = append(result.Errors, err)
			return nil
		}

		if d.IsDir() || strings.HasPrefix(d.Name(), ".") {
			return nil
		}

		stat, err := d.Info()
		if err != nil {
			result.Errors = append(result.Errors, err)
			return nil
		}

		size := stat.Size()
		parentDir := filepath.Dir(path)
		fileOp, err := sorter.Sort(dir, parentDir, d.Name(), size)
		if err != nil {
			result.Errors = append(result.Errors, err)
			return nil
		}

		moved, err := executor.Execute(fileOp)

		if err != nil {
			result.Errors = append(result.Errors, err)
			return nil
		}
		if moved {
			result.Moved++
		}
		if fileOp.Type == OpSkip {
			result.Skipped++
		}

		return nil
	})

	if walkErr != nil {
		result.Errors = append(result.Errors, walkErr)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		result.Errors = append(result.Errors, err)
		return result, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			dirPath := filepath.Join(dir, entry.Name())
			f, err := os.Open(dirPath)
			if err != nil {
				result.Errors = append(result.Errors, err)
				continue
			}
			_, err = f.Readdir(1)
			f.Close()
			if err == io.EOF {
				if err := os.Remove(dirPath); err != nil {
					result.Errors = append(result.Errors, err)
				}
			}
		}
	}

	if DuplNuke {
		os.RemoveAll(filepath.Join(dir, "duplicates"))
	}
	return result, nil
}

func TopLargestFiles(dir string, n int) error {
	var entries []FileInfo
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
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
		return err
	}

	if len(entries) < 1 {
		return nil
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Size > entries[j].Size
	})

	limit := min(len(entries), n)
	if limit == 0 || strings.HasPrefix(entries[0].Name, ".") {
		return nil
	}
	fmt.Printf("Top %d largest files in %s:\n", limit, dir)
	for i := range limit {
		fmt.Printf("%d. %s (%s)\n", i+1, entries[i].Name, humanReadable(entries[i].Size))
	}

	return nil
}
