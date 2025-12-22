package internal

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
)

func NewDuplicateFinder() *DuplicateFinder {
	return &DuplicateFinder{
		hashes: make(map[string]string),
	}
}

func (d *DuplicateFinder) Decide(files []FileEntry) ([]FileOperation, error) {
	ops := make([]FileOperation, 0, len(files))

	for _, f := range files {

		if f.SourcePath == filepath.Join(f.RootDir, "duplicates", (filepath.Base(f.SourcePath))) {
			ops = append(ops, FileOperation{
				OpType: OpSkip,
			})
			continue
		}

		data, err := os.ReadFile(f.SourcePath)
		if err != nil {
			return nil, err
		}

		checksum := fmt.Sprintf("%x", sha256.Sum256(data))

		if _, exists := d.hashes[checksum]; !exists {
			d.hashes[checksum] = f.SourcePath
			ops = append(ops, FileOperation{OpType: OpSkip})
			continue
		}

		ops = append(ops, FileOperation{
			OpType:   OpMove,
			File:     f,
			DestPath: filepath.Join(f.RootDir, "duplicates", (filepath.Base(f.SourcePath))),
		})
	}

	return ops, nil
}
