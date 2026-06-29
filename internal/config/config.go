package config

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/electr1fy0/sorta/internal/core"
	"github.com/electr1fy0/sorta/templates"
	"gopkg.in/yaml.v3"
)

type MatcherType int

const (
	MatcherKeyword   MatcherType = iota
	MatcherExtension
	MatcherRegex
	MatcherCatchAll
)

type TypedMatcher struct {
	Raw    string
	Regex  *regexp.Regexp
	Negate bool
	Type   MatcherType
}

type Rule struct {
	Folder   string
	Priority int
	MatchAll bool
	Matchers []TypedMatcher
}

type ConfigData struct {
	Rules     []Rule
	Blacklist []string
	Warnings  []string
}

type RuleSpec struct {
	Folder    string
	Keywords  []string
	Regex     []string
	Priority  int
	MatchAll  bool
}

type yamlConfig struct {
	Rules  []yamlRule  `yaml:"rules"`
	Ignore []string    `yaml:"ignore,omitempty"`
}

type yamlRule struct {
	Folder    string   `yaml:"folder"`
	Keywords  []string `yaml:"keywords,omitempty"`
	Regex     []string `yaml:"regex,omitempty"`
	Priority  int      `yaml:"priority,omitempty"`
	Match     string   `yaml:"match,omitempty"`
	CatchAll  bool     `yaml:"catch_all,omitempty"`
}

func LoadConfig(explicitPath, targetDir string) (*ConfigData, string, error) {
	path, err := ResolveConfigPath(explicitPath, targetDir)
	if err != nil {
		return nil, "", err
	}

	cfg, err := ParseConfig(path)
	if err != nil {
		return nil, path, err
	}
	for _, warning := range cfg.Warnings {
		fmt.Fprintf(os.Stderr, "config warning: %s\n", warning)
	}
	return cfg, path, nil
}

func ResolveConfigPath(explicitPath, targetDir string) (string, error) {
	if explicitPath != "" {
		return explicitPath, nil
	}

	if targetDir != "" {
		localPath := filepath.Join(targetDir, ".sorta", "config")
		if _, err := os.Stat(localPath); err == nil {
			return localPath, nil
		}
	}

	globalDir, err := core.GetSortaDir()
	if err != nil {
		return "", err
	}
	globalPath := filepath.Join(globalDir, "config")

	if _, err := os.Stat(globalPath); os.IsNotExist(err) {
		if err := createGlobalConfig(globalDir, globalPath); err != nil {
			return "", fmt.Errorf("failed to create global config: %w", err)
		}
	}

	return globalPath, nil
}

func createGlobalConfig(dir, path string) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(templates.DefaultConfig), 0644)
}

func ParseInline(s string) (*ConfigData, error) {
	normalized := strings.ReplaceAll(s, `\n`, "\n")
	return parseYAML([]byte(normalized))
}

func BuildConfigData(rules []RuleSpec) (*ConfigData, error) {
	configData := &ConfigData{
		Rules:     make([]Rule, 0, len(rules)),
		Blacklist: make([]string, 0),
		Warnings:  make([]string, 0),
	}

	for _, rs := range rules {
		folder := strings.TrimSpace(rs.Folder)
		if folder == "" {
			return nil, fmt.Errorf("rule folder cannot be empty")
		}

		var matchers []TypedMatcher
		for _, kw := range rs.Keywords {
			kw = strings.TrimSpace(kw)
			if kw == "" {
				continue
			}
			matcher, err := parseMatcherStr(kw)
			if err != nil {
				return nil, err
			}
			matchers = append(matchers, matcher)
		}
		for _, re := range rs.Regex {
			re = strings.TrimSpace(re)
			if re == "" {
				continue
			}
			matcher, err := parseMatcherStr("regex(" + re + ")")
			if err != nil {
				return nil, err
			}
			matchers = append(matchers, matcher)
		}

		if len(matchers) == 0 {
			return nil, fmt.Errorf("rule %q must include at least one matcher", folder)
		}

		configData.Rules = append(configData.Rules, Rule{
			Folder:   folder,
			Priority: rs.Priority,
			MatchAll: rs.MatchAll,
			Matchers: matchers,
		})
	}

	if len(configData.Rules) == 0 {
		return nil, fmt.Errorf("at least one rule is required")
	}

	return configData, nil
}

func ParseConfig(configPath string) (*ConfigData, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("config file not found at %s", configPath)
		}
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	return parseYAML(data)
}

func parseConfigFromReader(r io.Reader) (*ConfigData, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}
	return parseYAML(data)
}

func parseYAML(data []byte) (*ConfigData, error) {
	var yc yamlConfig
	if err := yaml.Unmarshal(data, &yc); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	cfg := &ConfigData{
		Rules:     make([]Rule, 0, len(yc.Rules)),
		Blacklist: make([]string, 0),
		Warnings:  make([]string, 0),
	}

	for i, yr := range yc.Rules {
		r := Rule{
			Folder:   yr.Folder,
			Priority: yr.Priority,
		}

		switch strings.ToLower(yr.Match) {
		case "all":
			r.MatchAll = true
		case "any", "":
			r.MatchAll = false
		default:
			cfg.Warnings = append(cfg.Warnings, fmt.Sprintf("rule %d: unknown match %q, using 'any'", i+1, yr.Match))
		}

		if yr.CatchAll {
			r.Matchers = append(r.Matchers, TypedMatcher{Type: MatcherCatchAll})
		}

		if len(yr.Keywords) == 0 && len(yr.Regex) == 0 && !yr.CatchAll {
			cfg.Warnings = append(cfg.Warnings, fmt.Sprintf("rule %d (%q): no matchers specified", i+1, yr.Folder))
		}

		for _, kw := range yr.Keywords {
			kw = strings.TrimSpace(kw)
			if kw == "" {
				continue
			}
			m, err := parseMatcherStr(kw)
			if err != nil {
				cfg.Warnings = append(cfg.Warnings, fmt.Sprintf("rule %d: %v", i+1, err))
				continue
			}
			r.Matchers = append(r.Matchers, m)
		}

		for _, re := range yr.Regex {
			re = strings.TrimSpace(re)
			if re == "" {
				continue
			}
			negate := strings.HasPrefix(re, "!")
			if negate {
				re = strings.TrimSpace(re[1:])
			}
			compiled, err := regexp.Compile(re)
			if err != nil {
				cfg.Warnings = append(cfg.Warnings, fmt.Sprintf("rule %d: invalid regex %q: %v", i+1, re, err))
				continue
			}
			r.Matchers = append(r.Matchers, TypedMatcher{
				Type:   MatcherRegex,
				Raw:    re,
				Regex:  compiled,
				Negate: negate,
			})
		}

		if len(r.Matchers) == 0 {
			cfg.Warnings = append(cfg.Warnings, fmt.Sprintf("rule %d (%q): no valid matchers, skipping", i+1, yr.Folder))
			continue
		}

		cfg.Rules = append(cfg.Rules, r)
	}

	for _, ig := range yc.Ignore {
		ig = strings.TrimSpace(ig)
		if ig != "" {
			cfg.Blacklist = append(cfg.Blacklist, ig)
		}
	}

	if len(cfg.Rules) == 0 {
		return nil, fmt.Errorf("config must contain at least one rule")
	}

	return cfg, nil
}

func parseMatcherStr(s string) (TypedMatcher, error) {
	s = strings.TrimSpace(s)

	negate := strings.HasPrefix(s, "!")
	if negate {
		s = strings.TrimSpace(s[1:])
	}

	if trimmed, ok := strings.CutPrefix(s, "regex("); ok {
		if trimmed, ok = strings.CutSuffix(trimmed, ")"); ok {
			re, err := regexp.Compile(trimmed)
			if err != nil {
				return TypedMatcher{}, fmt.Errorf("invalid regex %q: %w", trimmed, err)
			}
			return TypedMatcher{Type: MatcherRegex, Raw: trimmed, Regex: re, Negate: negate}, nil
		}
		return TypedMatcher{}, fmt.Errorf("malformed regex matcher %q", s)
	}

	if s == "*" {
		return TypedMatcher{Type: MatcherCatchAll, Negate: negate}, nil
	}

	if strings.HasPrefix(s, ".") {
		return TypedMatcher{Type: MatcherExtension, Raw: s, Negate: negate}, nil
	}

	return TypedMatcher{Type: MatcherKeyword, Raw: s, Negate: negate}, nil
}

func Categorize(configData ConfigData, filename string) string {
	if len(configData.Rules) == 0 {
		return ""
	}

	bestFolder := ""
	bestPriority := -1

	for _, rule := range configData.Rules {
		matched := evaluateRule(rule, filename)
		if matched {
			if rule.Priority >= bestPriority {
				bestFolder = rule.Folder
				bestPriority = rule.Priority
			}
		}
	}

	return bestFolder
}

func evaluateRule(rule Rule, filename string) bool {
	lower := strings.ToLower(filename)

	for _, m := range rule.Matchers {
		hit := matchTyped(m, lower, filename)
		if m.Negate {
			hit = !hit
		}

		if rule.MatchAll && !hit {
			return false
		}
		if !rule.MatchAll && hit {
			return true
		}
	}

	return rule.MatchAll
}

func matchTyped(m TypedMatcher, lower, original string) bool {
	switch m.Type {
	case MatcherCatchAll:
		return true
	case MatcherKeyword:
		return strings.Contains(lower, strings.ToLower(m.Raw))
	case MatcherExtension:
		return strings.HasSuffix(strings.ToLower(original), strings.ToLower(m.Raw))
	case MatcherRegex:
		return m.Regex.MatchString(original)
	default:
		return false
	}
}
