package ignore

import (
	"bufio"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/electr1fy0/sorta/internal/core"
)

type IgnoreMatcher struct {
	rules []IgnoreRule
}

type IgnoreRule struct {
	Pattern string
	Source  string
}

func LoadIgnoreMatcher(rootDir string, inlinePatterns []string) (*IgnoreMatcher, error) {
	rules := make([]IgnoreRule, 0, len(inlinePatterns)+16)
	rules = append(rules, sanitizeInlinePatterns(inlinePatterns)...)

	paths := []string{
		filepath.Join(rootDir, ".sortaignore"),
		filepath.Join(rootDir, ".sorta", "ignore"),
	}

	sortaDir, err := core.GetSortaDir()
	if err == nil {
		paths = append(paths, filepath.Join(sortaDir, "ignore"))
	}

	for _, p := range paths {
		lines, err := readIgnoreFile(p)
		if err != nil {
			return nil, err
		}
		rules = append(rules, lines...)
	}

	rules = dedupeRules(rules)
	return &IgnoreMatcher{rules: rules}, nil
}

func (m *IgnoreMatcher) Match(rootDir, candidatePath string, isDir bool) bool {
	_, ok := m.Explain(rootDir, candidatePath, isDir)
	return ok
}

func (m *IgnoreMatcher) Explain(rootDir, candidatePath string, isDir bool) (IgnoreRule, bool) {
	if m == nil || len(m.rules) == 0 {
		return IgnoreRule{}, false
	}

	rel, err := filepath.Rel(rootDir, candidatePath)
	if err != nil {
		return IgnoreRule{}, false
	}
	rel = filepath.ToSlash(filepath.Clean(rel))
	if rel == "." {
		return IgnoreRule{}, false
	}
	name := path.Base(rel)

	for _, rule := range m.rules {
		if matchesPattern(rule.Pattern, rel, name, isDir) {
			return rule, true
		}
	}

	return IgnoreRule{}, false
}

func readIgnoreFile(path string) ([]IgnoreRule, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()

	rules := make([]IgnoreRule, 0, 32)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "//") {
			continue
		}
		rules = append(rules, IgnoreRule{
			Pattern: strings.TrimSuffix(line, "/"),
			Source:  path,
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return rules, nil
}

func sanitizeInlinePatterns(patterns []string) []IgnoreRule {
	out := make([]IgnoreRule, 0, len(patterns))
	for _, p := range patterns {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		out = append(out, IgnoreRule{
			Pattern: strings.TrimSuffix(filepath.ToSlash(p), "/"),
			Source:  "config",
		})
	}
	return out
}

func dedupeRules(rules []IgnoreRule) []IgnoreRule {
	seen := make(map[string]struct{}, len(rules))
	out := make([]IgnoreRule, 0, len(rules))
	for _, r := range rules {
		key := r.Source + "::" + r.Pattern
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, r)
	}
	return out
}

func matchesPattern(pattern, rel, name string, isDir bool) bool {
	pattern = filepath.ToSlash(strings.TrimSpace(pattern))
	if pattern == "" {
		return false
	}

	anchored := strings.HasPrefix(pattern, "/")
	p := strings.TrimPrefix(pattern, "/")

	if strings.Contains(p, "/") || anchored {
		if ok, _ := path.Match(p, rel); ok {
			return true
		}
		if isDir && (rel == p || strings.HasPrefix(rel, p+"/")) {
			return true
		}
		return false
	}

	if ok, _ := path.Match(p, name); ok {
		return true
	}
	if isDir && (name == p || strings.HasPrefix(name, p)) {
		return true
	}

	return false
}
