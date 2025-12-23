package internal

import (
	"crypto/sha256"
	"fmt"
	"io"
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
	validFiles := make([]FileEntry, 0, len(files))

	for _, f := range files {
		if f.SourcePath == filepath.Join(f.RootDir, "duplicates", (filepath.Base(f.SourcePath))) {
			ops = append(ops, FileOperation{OpType: OpSkip})
			continue
		}
		validFiles = append(validFiles, f)
	}

	bySize := make(map[int64][]FileEntry)
	for _, f := range validFiles {
		bySize[f.Size] = append(bySize[f.Size], f)
	}

	for _, group := range bySize {
		if len(group) == 1 {
			ops = append(ops, FileOperation{OpType: OpSkip})
			continue
		}

		byPartial := make(map[string][]FileEntry)
		for _, f := range group {
			h, err := partialHash(f.SourcePath)
			if err != nil {
				return nil, err
			}
			byPartial[h] = append(byPartial[h], f)
		}

		for _, pGroup := range byPartial {
			if len(pGroup) == 1 {
				ops = append(ops, FileOperation{OpType: OpSkip})
				continue
			}

			byFull := make(map[string][]FileEntry)
			for _, f := range pGroup {
				h, err := fullHash(f.SourcePath)
				if err != nil {
					return nil, err
				}
				byFull[h] = append(byFull[h], f)
			}

			for _, fGroup := range byFull {
				ops = append(ops, FileOperation{OpType: OpSkip})
				for i := 1; i < len(fGroup); i++ {
					f := fGroup[i]
					ops = append(ops, FileOperation{
						OpType:   OpMove,
						File:     f,
						DestPath: filepath.Join(f.RootDir, "duplicates", (filepath.Base(f.SourcePath))),
					})
				}
			}
		}
	}

	return ops, nil
}

func partialHash(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	buf := make([]byte, 4096)
	n, err := io.ReadFull(f, buf)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return "", err
	}
	return fmt.Sprintf("%x", sha256.Sum256(buf[:n])), nil
}

func fullHash(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}
