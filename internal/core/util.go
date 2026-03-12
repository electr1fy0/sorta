package core

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func HumanReadable(n int64) string {
	const unit int64 = 1024
	if n < unit {
		return fmt.Sprintf("%d B", n)
	}

	div, exp := unit, 0
	for i := n / unit; i >= unit; i /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.1f %cB", float64(n)/float64(div), "KMGTPE"[exp])
}

func ExpandPath(path string) (string, error) {
	if strings.HasPrefix(path, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("cannot determine home directory: %w", err)
		}
		path = filepath.Join(home, path[1:])
	}
	return path, nil
}

func GetSortaDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home directory: %w", err)
	}
	return filepath.Join(home, ".sorta"), nil
}

func WriteFileAtomic(path string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	tmp, err := os.CreateTemp(dir, filepath.Base(path)+".tmp-*")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Chmod(perm); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}

	return os.Rename(tmpPath, path)
}

func AppendLineAtomic(path string, line string, perm os.FileMode) error {
	var existing []byte
	data, err := os.ReadFile(path)
	if err == nil {
		existing = data
	} else if !os.IsNotExist(err) {
		return err
	}

	buf := make([]byte, 0, len(existing)+len(line)+1)
	buf = append(buf, existing...)
	buf = append(buf, line...)
	if len(line) == 0 || line[len(line)-1] != '\n' {
		buf = append(buf, '\n')
	}

	return WriteFileAtomic(path, buf, perm)
}
