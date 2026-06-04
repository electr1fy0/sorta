package core

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHumanReadable(t *testing.T) {
	tests := []struct {
		input int64
		want  string
	}{
		{0, "0 B"},
		{512, "512 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
		{1073741824, "1.0 GB"},
		{1099511627776, "1.0 TB"},
	}
	for _, tc := range tests {
		got := HumanReadable(tc.input)
		if got != tc.want {
			t.Errorf("HumanReadable(%d) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestExpandTildePath(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		input string
		want  string
	}{
		{"~/docs", filepath.Join(home, "docs")},
		{"/abs/path", "/abs/path"},
		{"rel/path", "rel/path"},
	}
	for _, tc := range tests {
		got, err := ExpandTildePath(tc.input)
		if err != nil {
			t.Fatalf("ExpandTildePath(%q): %v", tc.input, err)
		}
		if got != tc.want {
			t.Errorf("ExpandTildePath(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestGetSortaDir(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}
	got, err := GetSortaDir()
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(home, ".sorta")
	if got != want {
		t.Errorf("GetSortaDir() = %q, want %q", got, want)
	}
}

func TestWriteFileAtomic(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sub", "test.txt")
	data := []byte("hello world")

	if err := WriteFileAtomic(path, data, 0644); err != nil {
		t.Fatal(err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(data) {
		t.Errorf("WriteFileAtomic content = %q, want %q", string(got), string(data))
	}
}

func TestAppendLineAtomic(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "log.txt")

	if err := AppendLineAtomic(path, "line1", 0644); err != nil {
		t.Fatal(err)
	}
	if err := AppendLineAtomic(path, "line2", 0644); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 2 || lines[0] != "line1" || lines[1] != "line2" {
		t.Errorf("AppendLineAtomic got lines %v, want [line1 line2]", lines)
	}
}
