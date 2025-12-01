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

func (d *DuplicateFinder) Sort(filepaths []FilePath) ([]FileOperation, error) {
	ops := make([]FileOperation, 0, len(filepaths))

	for _, fp := range filepaths {
		fullPath := filepath.Join(fp.FullDir, fp.Filename)

		if fullPath == filepath.Join(fp.BaseDir, "duplicates", fp.Filename) {
			ops = append(ops, FileOperation{
				Type: OpSkip,
			})
			continue
		}

		data, err := os.ReadFile(fullPath)
		if err != nil {
			return nil, err
		}

		checksum := fmt.Sprintf("%x", sha256.Sum256(data))

		if _, exists := d.hashes[checksum]; !exists {
			d.hashes[checksum] = fullPath
			ops = append(ops, FileOperation{Type: OpSkip})
			continue
		}

		ops = append(ops, FileOperation{
			Type:       OpMove,
			SourcePath: fullPath,
			DestPath:   filepath.Join(fp.BaseDir, "duplicates", fp.Filename),
			Filename:   fp.Filename,
			Size:       fp.Size,
		})
	}

	return ops, nil
}
