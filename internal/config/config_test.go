package config

import (
	"regexp"
	"strings"
	"testing"
)

func TestParseYAML(t *testing.T) {
	t.Run("basic rules", func(t *testing.T) {
		input := `
rules:
  - folder: Images
    keywords: [jpg, png, .gif]
  - folder: Documents
    keywords: [pdf, docx]
`
		cfg, err := parseYAML([]byte(input))
		if err != nil {
			t.Fatal(err)
		}
		if len(cfg.Rules) != 2 {
			t.Fatalf("expected 2 rules, got %d", len(cfg.Rules))
		}
		if cfg.Rules[0].Folder != "Images" || cfg.Rules[1].Folder != "Documents" {
			t.Errorf("unexpected folders: %q %q", cfg.Rules[0].Folder, cfg.Rules[1].Folder)
		}
		if len(cfg.Rules[0].Matchers) != 3 || len(cfg.Rules[1].Matchers) != 2 {
			t.Errorf("unexpected matcher counts: %d %d", len(cfg.Rules[0].Matchers), len(cfg.Rules[1].Matchers))
		}
	})

	t.Run("priority and match all", func(t *testing.T) {
		input := `
rules:
  - folder: Finance
    keywords: [.pdf, invoice]
    priority: 10
    match: all
  - folder: Docs
    keywords: [.pdf]
    priority: 5
`
		cfg, err := parseYAML([]byte(input))
		if err != nil {
			t.Fatal(err)
		}
		if len(cfg.Rules) != 2 {
			t.Fatalf("expected 2 rules, got %d", len(cfg.Rules))
		}
		if cfg.Rules[0].Priority != 10 || !cfg.Rules[0].MatchAll {
			t.Errorf("expected priority 10, match_all=true, got %d %v", cfg.Rules[0].Priority, cfg.Rules[0].MatchAll)
		}
		if cfg.Rules[1].Priority != 5 || cfg.Rules[1].MatchAll {
			t.Errorf("expected priority 5, match_all=false, got %d %v", cfg.Rules[1].Priority, cfg.Rules[1].MatchAll)
		}
	})

	t.Run("catch all", func(t *testing.T) {
		input := `
rules:
  - folder: Other
    catch_all: true
  - folder: Images
    keywords: [jpg]
    priority: 10
`
		cfg, err := parseYAML([]byte(input))
		if err != nil {
			t.Fatal(err)
		}
		if len(cfg.Rules) != 2 {
			t.Fatalf("expected 2 rules, got %d", len(cfg.Rules))
		}
		if len(cfg.Rules[0].Matchers) != 1 || cfg.Rules[0].Matchers[0].Type != MatcherCatchAll {
			t.Errorf("expected catch-all matcher on first rule")
		}
	})

	t.Run("ignore list", func(t *testing.T) {
		input := `
rules:
  - folder: Images
    keywords: [jpg]
ignore:
  - "*.bak"
  - temp/
`
		cfg, err := parseYAML([]byte(input))
		if err != nil {
			t.Fatal(err)
		}
		if len(cfg.Blacklist) != 2 {
			t.Fatalf("expected 2 blacklist entries, got %d", len(cfg.Blacklist))
		}
		if cfg.Blacklist[0] != "*.bak" || cfg.Blacklist[1] != "temp/" {
			t.Errorf("unexpected blacklist: %v", cfg.Blacklist)
		}
	})

	t.Run("regex matchers", func(t *testing.T) {
		input := `
rules:
  - folder: Code
    regex: ["\\.go$", "\\.ts$"]
`
		cfg, err := parseYAML([]byte(input))
		if err != nil {
			t.Fatal(err)
		}
		if len(cfg.Rules[0].Matchers) != 2 {
			t.Fatalf("expected 2 matchers, got %d", len(cfg.Rules[0].Matchers))
		}
		if cfg.Rules[0].Matchers[0].Type != MatcherRegex || cfg.Rules[0].Matchers[1].Type != MatcherRegex {
			t.Error("expected regex matchers")
		}
	})

	t.Run("negated keyword", func(t *testing.T) {
		input := `
rules:
  - folder: Work
    keywords: [.pdf, "!draft"]
`
		cfg, err := parseYAML([]byte(input))
		if err != nil {
			t.Fatal(err)
		}
		if len(cfg.Rules[0].Matchers) != 2 {
			t.Fatalf("expected 2 matchers, got %d", len(cfg.Rules[0].Matchers))
		}
		if !cfg.Rules[0].Matchers[1].Negate {
			t.Error("expected negate on 'draft'")
		}
	})

	t.Run("comments", func(t *testing.T) {
		input := `
# this is a comment
rules:
  - folder: Images
    keywords: [jpg]
`
		cfg, err := parseYAML([]byte(input))
		if err != nil {
			t.Fatal(err)
		}
		if len(cfg.Rules) != 1 || cfg.Rules[0].Folder != "Images" {
			t.Errorf("expected 1 rule, got %+v", cfg.Rules)
		}
	})

	t.Run("empty config", func(t *testing.T) {
		_, err := parseYAML([]byte(""))
		if err == nil || !strings.Contains(err.Error(), "at least one rule") {
			t.Errorf("expected empty config error, got %v", err)
		}
	})
}

func TestParseInline(t *testing.T) {
	cfg, err := ParseInline("rules:\n  - folder: Images\n    keywords: [jpg]\n  - folder: Documents\n    keywords: [pdf]")
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Rules) != 2 {
		t.Errorf("expected 2 rules, got %d", len(cfg.Rules))
	}
}

func TestParseMatcherStr(t *testing.T) {
	t.Run("keyword", func(t *testing.T) {
		m, err := parseMatcherStr("jpg")
		if err != nil {
			t.Fatal(err)
		}
		if m.Type != MatcherKeyword || m.Raw != "jpg" || m.Negate {
			t.Errorf("expected keyword matcher, got %+v", m)
		}
	})

	t.Run("extension", func(t *testing.T) {
		m, err := parseMatcherStr(".pdf")
		if err != nil {
			t.Fatal(err)
		}
		if m.Type != MatcherExtension || m.Raw != ".pdf" {
			t.Errorf("expected extension matcher, got %+v", m)
		}
	})

	t.Run("catch all", func(t *testing.T) {
		m, err := parseMatcherStr("*")
		if err != nil {
			t.Fatal(err)
		}
		if m.Type != MatcherCatchAll {
			t.Errorf("expected catch-all matcher, got %+v", m)
		}
	})

	t.Run("negated", func(t *testing.T) {
		m, err := parseMatcherStr("!draft")
		if err != nil {
			t.Fatal(err)
		}
		if m.Type != MatcherKeyword || m.Raw != "draft" || !m.Negate {
			t.Errorf("expected negated keyword, got %+v", m)
		}
	})

	t.Run("valid regex", func(t *testing.T) {
		m, err := parseMatcherStr("regex(\\.go$)")
		if err != nil {
			t.Fatal(err)
		}
		if m.Type != MatcherRegex || m.Regex == nil {
			t.Fatal("expected regex matcher")
		}
		if !m.Regex.MatchString("main.go") {
			t.Error("regex should match main.go")
		}
		if m.Regex.MatchString("main.txt") {
			t.Error("regex should not match main.txt")
		}
	})

	t.Run("invalid regex", func(t *testing.T) {
		_, err := parseMatcherStr("regex([invalid)")
		if err == nil {
			t.Error("expected error for invalid regex")
		}
	})

	t.Run("malformed regex wrapper", func(t *testing.T) {
		_, err := parseMatcherStr("regex(jpg")
		if err == nil {
			t.Error("expected error for malformed regex")
		}
	})
}

func TestBuildConfigData(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		cfg, err := BuildConfigData([]RuleSpec{
			{Folder: "Images", Keywords: []string{"jpg", "png"}},
			{Folder: "Docs", Keywords: []string{"pdf"}},
		})
		if err != nil {
			t.Fatal(err)
		}
		if len(cfg.Rules) != 2 {
			t.Fatalf("expected 2 rules, got %d", len(cfg.Rules))
		}
	})

	t.Run("empty folder", func(t *testing.T) {
		_, err := BuildConfigData([]RuleSpec{
			{Folder: "", Keywords: []string{"jpg"}},
		})
		if err == nil {
			t.Error("expected error for empty folder")
		}
	})

	t.Run("no matchers", func(t *testing.T) {
		_, err := BuildConfigData([]RuleSpec{
			{Folder: "Images", Keywords: []string{}},
		})
		if err == nil {
			t.Error("expected error for no matchers")
		}
	})

	t.Run("no rules", func(t *testing.T) {
		_, err := BuildConfigData([]RuleSpec{})
		if err == nil {
			t.Error("expected error for no rules")
		}
	})
}

func TestCategorize(t *testing.T) {
	cfg := &ConfigData{
		Rules: []Rule{
			{Folder: "Images", Matchers: []TypedMatcher{
				{Type: MatcherKeyword, Raw: "jpg"},
				{Type: MatcherKeyword, Raw: "png"},
			}},
			{Folder: "Documents", Matchers: []TypedMatcher{
				{Type: MatcherKeyword, Raw: "pdf"},
				{Type: MatcherKeyword, Raw: "docx"},
			}},
		},
	}

	tests := []struct {
		filename string
		want     string
	}{
		{"photo.jpg", "Images"},
		{"image.png", "Images"},
		{"report.pdf", "Documents"},
		{"notes.docx", "Documents"},
		{"script.go", ""},
	}
	for _, tc := range tests {
		got := Categorize(*cfg, tc.filename)
		if got != tc.want {
			t.Errorf("Categorize(%q) = %q, want %q", tc.filename, got, tc.want)
		}
	}

	t.Run("case insensitive", func(t *testing.T) {
		got := Categorize(*cfg, "PHOTO.JPG")
		if got != "Images" {
			t.Errorf("expected case-insensitive match, got %q", got)
		}
	})

	t.Run("catch all fallback", func(t *testing.T) {
		cfg2 := &ConfigData{
			Rules: []Rule{
				{Folder: "Other", Priority: 0, Matchers: []TypedMatcher{
					{Type: MatcherCatchAll},
				}},
				{Folder: "Images", Priority: 10, Matchers: []TypedMatcher{
					{Type: MatcherKeyword, Raw: "jpg"},
				}},
			},
		}
		got := Categorize(*cfg2, "unknown.txt")
		if got != "Other" {
			t.Errorf("expected fallback to 'Other', got %q", got)
		}
	})

	t.Run("exact match before catch all", func(t *testing.T) {
		cfg2 := &ConfigData{
			Rules: []Rule{
				{Folder: "Other", Priority: 0, Matchers: []TypedMatcher{
					{Type: MatcherCatchAll},
				}},
				{Folder: "Images", Priority: 10, Matchers: []TypedMatcher{
					{Type: MatcherKeyword, Raw: "jpg"},
				}},
			},
		}
		got := Categorize(*cfg2, "photo.jpg")
		if got != "Images" {
			t.Errorf("expected exact match 'Images', got %q", got)
		}
	})

	t.Run("same priority: later rule wins", func(t *testing.T) {
		cfg2 := &ConfigData{
			Rules: []Rule{
				{Folder: "Images", Matchers: []TypedMatcher{
					{Type: MatcherKeyword, Raw: "jpg"},
				}},
				{Folder: "Photos", Matchers: []TypedMatcher{
					{Type: MatcherKeyword, Raw: "jpg"},
				}},
			},
		}
		got := Categorize(*cfg2, "photo.jpg")
		if got != "Photos" {
			t.Errorf("expected later rule 'Photos', got %q", got)
		}
	})

	t.Run("priority overrides order", func(t *testing.T) {
		cfg2 := &ConfigData{
			Rules: []Rule{
				{Folder: "Images", Priority: 5, Matchers: []TypedMatcher{
					{Type: MatcherKeyword, Raw: "jpg"},
				}},
				{Folder: "Photos", Priority: 1, Matchers: []TypedMatcher{
					{Type: MatcherKeyword, Raw: "jpg"},
				}},
			},
		}
		got := Categorize(*cfg2, "photo.jpg")
		if got != "Images" {
			t.Errorf("expected higher priority 'Images', got %q", got)
		}
	})

	t.Run("regex match", func(t *testing.T) {
		re := regexp.MustCompile(`\.go$`)
		cfg3 := &ConfigData{
			Rules: []Rule{
				{Folder: "Code", Matchers: []TypedMatcher{
					{Type: MatcherRegex, Raw: "\\.go$", Regex: re},
				}},
			},
		}
		got := Categorize(*cfg3, "main.go")
		if got != "Code" {
			t.Errorf("expected 'Code', got %q", got)
		}
		got2 := Categorize(*cfg3, "main.txt")
		if got2 != "" {
			t.Errorf("expected no match, got %q", got2)
		}
	})

	t.Run("AND match requires all matchers", func(t *testing.T) {
		cfg4 := &ConfigData{
			Rules: []Rule{
				{Folder: "Docs", MatchAll: true, Matchers: []TypedMatcher{
					{Type: MatcherKeyword, Raw: "pdf"},
					{Type: MatcherKeyword, Raw: "report"},
				}},
			},
		}
		if got := Categorize(*cfg4, "report.pdf"); got != "Docs" {
			t.Errorf("expected 'Docs', got %q", got)
		}
		if got := Categorize(*cfg4, "notes.pdf"); got != "" {
			t.Errorf("expected no match for pdf without report, got %q", got)
		}
		if got := Categorize(*cfg4, "report.txt"); got != "" {
			t.Errorf("expected no match for report without pdf, got %q", got)
		}
	})

	t.Run("negated matcher in AND mode", func(t *testing.T) {
		cfg5 := &ConfigData{
			Rules: []Rule{
				{Folder: "Docs", MatchAll: true, Priority: 10, Matchers: []TypedMatcher{
					{Type: MatcherKeyword, Raw: "pdf"},
					{Type: MatcherKeyword, Raw: "draft", Negate: true},
				}},
				{Folder: "Drafts", Priority: 5, Matchers: []TypedMatcher{
					{Type: MatcherKeyword, Raw: "pdf"},
					{Type: MatcherKeyword, Raw: "draft"},
				}},
			},
		}
		if got := Categorize(*cfg5, "report.pdf"); got != "Docs" {
			t.Errorf("expected 'Docs' for pdf without draft, got %q", got)
		}
		if got := Categorize(*cfg5, "draft.pdf"); got != "Drafts" {
			t.Errorf("expected 'Drafts' for draft.pdf (OR match), got %q", got)
		}
	})

	t.Run("extension match", func(t *testing.T) {
		cfg6 := &ConfigData{
			Rules: []Rule{
				{Folder: "Images", Matchers: []TypedMatcher{
					{Type: MatcherExtension, Raw: ".jpg"},
				}},
			},
		}
		if got := Categorize(*cfg6, "photo.jpg"); got != "Images" {
			t.Errorf("expected 'Images', got %q", got)
		}
		if got := Categorize(*cfg6, "photo.jpg.bak"); got != "" {
			t.Errorf("expected no match for .jpg.bak, got %q", got)
		}
	})
}
