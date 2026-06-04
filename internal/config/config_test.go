package config

import (
	"regexp"
	"strings"
	"testing"
)

func TestParseConfigFromReader(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		input := `Images = jpg, png, .gif
Documents = pdf, docx
`
		cfg, err := parseConfigFromReader(strings.NewReader(input))
		if err != nil {
			t.Fatal(err)
		}
		if len(cfg.Foldernames) != 2 {
			t.Fatalf("expected 2 folders, got %d", len(cfg.Foldernames))
		}
		if cfg.Foldernames[0] != "Images" || cfg.Foldernames[1] != "Documents" {
			t.Errorf("unexpected folder names: %v", cfg.Foldernames)
		}
		if len(cfg.Matchers[0]) != 3 || len(cfg.Matchers[1]) != 2 {
			t.Errorf("unexpected matcher counts: %d %d", len(cfg.Matchers[0]), len(cfg.Matchers[1]))
		}
	})

	t.Run("comments and blank lines", func(t *testing.T) {
		input := "# this is a comment\n// also a comment\n\nImages = jpg\n"
		cfg, err := parseConfigFromReader(strings.NewReader(input))
		if err != nil {
			t.Fatal(err)
		}
		if len(cfg.Foldernames) != 1 || cfg.Foldernames[0] != "Images" {
			t.Errorf("expected 1 folder, got %v", cfg.Foldernames)
		}
	})

	t.Run("blacklist", func(t *testing.T) {
		input := "Images = jpg\n!*.bak\n!temp/\n"
		cfg, err := parseConfigFromReader(strings.NewReader(input))
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

	t.Run("warnings on malformed lines", func(t *testing.T) {
		input := "Images = jpg\nbadline\n"
		cfg, err := parseConfigFromReader(strings.NewReader(input))
		if err != nil {
			t.Fatal(err)
		}
		if len(cfg.Warnings) == 0 {
			t.Error("expected warnings for bad lines")
		}
	})

	t.Run("empty config", func(t *testing.T) {
		_, err := parseConfigFromReader(strings.NewReader(""))
		if err == nil || !strings.Contains(err.Error(), "empty") {
			t.Errorf("expected empty config error, got %v", err)
		}
	})

	t.Run("regex matcher", func(t *testing.T) {
		input := "Code = regex(\\.go$), regex(\\.ts$)\n"
		cfg, err := parseConfigFromReader(strings.NewReader(input))
		if err != nil {
			t.Fatal(err)
		}
		if len(cfg.Matchers[0]) != 2 {
			t.Fatalf("expected 2 matchers, got %d", len(cfg.Matchers[0]))
		}
		if cfg.Matchers[0][0].Regex == nil || cfg.Matchers[0][1].Regex == nil {
			t.Error("expected regex matchers")
		}
	})
}

func TestParseInline(t *testing.T) {
	cfg, err := ParseInline(`Images = jpg\nDocuments = pdf`)
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Foldernames) != 2 {
		t.Errorf("expected 2 folders, got %d", len(cfg.Foldernames))
	}
}

func TestParseMatcher(t *testing.T) {
	t.Run("raw keyword", func(t *testing.T) {
		m, err := parseMatcher("jpg")
		if err != nil {
			t.Fatal(err)
		}
		if m.Raw != "jpg" || m.Regex != nil {
			t.Errorf("expected raw matcher, got %+v", m)
		}
	})

	t.Run("valid regex", func(t *testing.T) {
		m, err := parseMatcher("regex(\\.go$)")
		if err != nil {
			t.Fatal(err)
		}
		if m.Regex == nil {
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
		_, err := parseMatcher("regex([invalid)")
		if err == nil {
			t.Error("expected error for invalid regex")
		}
	})

	t.Run("malformed regex wrapper", func(t *testing.T) {
		_, err := parseMatcher("regex(jpg")
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
		if len(cfg.Foldernames) != 2 {
			t.Fatalf("expected 2 folders, got %d", len(cfg.Foldernames))
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

	t.Run("no keywords", func(t *testing.T) {
		_, err := BuildConfigData([]RuleSpec{
			{Folder: "Images", Keywords: []string{}},
		})
		if err == nil {
			t.Error("expected error for no keywords")
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
		Foldernames: []string{"Images", "Documents"},
		Matchers: [][]Matcher{
			{{Raw: "jpg"}, {Raw: "png"}},
			{{Raw: "pdf"}, {Raw: "docx"}},
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

	t.Run("fallback wildcard", func(t *testing.T) {
		cfg2 := &ConfigData{
			Foldernames: []string{"Other", "Images"},
			Matchers: [][]Matcher{
				{{Raw: "*"}},
				{{Raw: "jpg"}},
			},
		}
		got := Categorize(*cfg2, "unknown.txt")
		if got != "Other" {
			t.Errorf("expected fallback to 'Other', got %q", got)
		}
	})

	t.Run("exact match before fallback", func(t *testing.T) {
		cfg2 := &ConfigData{
			Foldernames: []string{"Other", "Images"},
			Matchers: [][]Matcher{
				{{Raw: "*"}},
				{{Raw: "jpg"}},
			},
		}
		got := Categorize(*cfg2, "photo.jpg")
		if got != "Images" {
			t.Errorf("expected exact match 'Images', got %q", got)
		}
	})

	t.Run("regex match", func(t *testing.T) {
		re := regexp.MustCompile(`\.go$`)
		cfg3 := &ConfigData{
			Foldernames: []string{"Code"},
			Matchers:    [][]Matcher{{{Regex: re}}},
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
}
