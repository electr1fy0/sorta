package ignore

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMatchesPattern(t *testing.T) {
	tests := []struct {
		pattern string
		rel     string
		name    string
		isDir   bool
		want    bool
	}{
		{"*.txt", "file.txt", "file.txt", false, true},
		{"*.txt", "file.go", "file.go", false, false},
		{"build", "build", "build", true, true},
		{"build", "build/output.o", "output.o", true, false},
		{"/absolute/path", "absolute/path", "path", false, true},
		{"dir/file", "dir/file", "file", false, true},
		{"", "anything", "anything", false, false},
	}
	for _, tc := range tests {
		got := matchesPattern(tc.pattern, tc.rel, tc.name, tc.isDir)
		if got != tc.want {
			t.Errorf("matchesPattern(%q, %q, %q, %v) = %v, want %v",
				tc.pattern, tc.rel, tc.name, tc.isDir, got, tc.want)
		}
	}
}

func TestDedupeRules(t *testing.T) {
	rules := []IgnoreRule{
		{Pattern: "*.bak", Source: "file1"},
		{Pattern: "*.bak", Source: "file1"},
		{Pattern: "*.tmp", Source: "file2"},
	}
	deduped := dedupeRules(rules)
	if len(deduped) != 2 {
		t.Errorf("expected 2 rules after dedup, got %d", len(deduped))
	}
}

func TestSanitizeInlinePatterns(t *testing.T) {
	patterns := []string{" *.bak ", "", "temp/"}
	result := sanitizeInlinePatterns(patterns)
	if len(result) != 2 {
		t.Fatalf("expected 2 patterns, got %d", len(result))
	}
	if result[0].Pattern != "*.bak" || result[0].Source != "config" {
		t.Errorf("unexpected first rule: %+v", result[0])
	}
	if result[1].Pattern != "temp" || result[1].Source != "config" {
		t.Errorf("unexpected second rule: %+v", result[1])
	}
}

func TestIgnoreMatcher(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, ".sortaignore"), []byte("*.bak\nsecret/\n"), 0644); err != nil {
		t.Fatal(err)
	}

	m, err := LoadIgnoreMatcher(dir, []string{"*.tmp"})
	if err != nil {
		t.Fatal(err)
	}

	testCases := []struct {
		path  string
		isDir bool
		want  bool
	}{
		{filepath.Join(dir, "file.txt"), false, false},
		{filepath.Join(dir, "file.bak"), false, true},
		{filepath.Join(dir, "file.tmp"), false, true},
		{filepath.Join(dir, "secret"), true, true},
	}
	for _, tc := range testCases {
		got := m.Match(dir, tc.path, tc.isDir)
		if got != tc.want {
			t.Errorf("Match(%q, isDir=%v) = %v, want %v", tc.path, tc.isDir, got, tc.want)
		}
	}
}

func TestIgnoreMatcherExplain(t *testing.T) {
	m := &IgnoreMatcher{
		rules: []IgnoreRule{
			{Pattern: "*.bak", Source: "test"},
		},
	}
	rule, ok := m.Explain("/root", "/root/file.bak", false)
	if !ok {
		t.Fatal("expected match")
	}
	if rule.Pattern != "*.bak" {
		t.Errorf("expected pattern *.bak, got %q", rule.Pattern)
	}

	_, ok = m.Explain("/root", "/root/file.txt", false)
	if ok {
		t.Error("expected no match")
	}
}

func TestNilIgnoreMatcher(t *testing.T) {
	var m *IgnoreMatcher
	if m.Match("/root", "/root/file.txt", false) {
		t.Error("nil matcher should not match")
	}
}
