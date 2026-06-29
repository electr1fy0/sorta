package rename

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode"

	"github.com/electr1fy0/sorta/internal/core"
	"github.com/electr1fy0/sorta/internal/llm"
	"github.com/electr1fy0/sorta/templates"
)

var defaultPrompt = templates.DefaultPrompt

var binaryExts = map[string]bool{
	".o": true, ".obj": true, ".a": true, ".lib": true,
	".so": true, ".dylib": true, ".dll": true,
	".class": true, ".jar": true,
	".png": true, ".jpg": true, ".jpeg": true, ".gif": true, ".bmp": true, ".ico": true,
	".mp4": true, ".avi": true, ".mkv": true, ".mov": true, ".webm": true,
	".mp3": true, ".wav": true, ".flac": true, ".aac": true, ".ogg": true,
	".zip": true, ".tar": true, ".gz": true, ".bz2": true, ".xz": true, ".7z": true, ".rar": true,
	".pyc": true, ".pyo": true, ".pyd": true,
	".exe": true, ".bin": true, ".dmg": true, ".iso": true,
	".db": true, ".sqlite": true, ".sqlite3": true,
	".ttf": true, ".otf": true, ".woff": true, ".woff2": true, ".eot": true,
	".pdf": true, ".doc": true, ".docx": true, ".xls": true, ".xlsx": true,
}

func isBinaryExt(name string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	return binaryExts[ext]
}

func hasPathTraversal(name string) bool {
	return strings.Contains(name, "..") || strings.Contains(name, "/") || strings.Contains(name, "\\")
}

type CaseType int

const (
	CaseNone CaseType = iota
	CaseSnake
	CaseKebab
	CaseCamel
	CasePascal
	CaseUpper
	CaseLower
	CaseTitle
)

func ParseCaseType(s string) (CaseType, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "snake":
		return CaseSnake, nil
	case "kebab":
		return CaseKebab, nil
	case "camel":
		return CaseCamel, nil
	case "pascal":
		return CasePascal, nil
	case "upper":
		return CaseUpper, nil
	case "lower":
		return CaseLower, nil
	case "title":
		return CaseTitle, nil
	case "", "none":
		return CaseNone, nil
	default:
		return CaseNone, fmt.Errorf("unknown case type %q: valid options are snake, kebab, camel, pascal, upper, lower, title", s)
	}
}

func (c CaseType) String() string {
	switch c {
	case CaseSnake:
		return "snake"
	case CaseKebab:
		return "kebab"
	case CaseCamel:
		return "camel"
	case CasePascal:
		return "pascal"
	case CaseUpper:
		return "upper"
	case CaseLower:
		return "lower"
	case CaseTitle:
		return "title"
	default:
		return "none"
	}
}

func splitWords(s string) []string {
	var words []string
	var buf strings.Builder

	runes := []rune(s)
	for i, r := range runes {
		if r == '_' || r == '-' || r == ' ' {
			if buf.Len() > 0 {
				words = append(words, buf.String())
				buf.Reset()
			}
			continue
		}

		if unicode.IsUpper(r) && buf.Len() > 0 {
			prev := runes[i-1]
			if unicode.IsLower(prev) {
				words = append(words, buf.String())
				buf.Reset()
			} else if i+1 < len(runes) && unicode.IsLower(runes[i+1]) && buf.Len() > 0 {
				words = append(words, buf.String())
				buf.Reset()
			}
		}

		buf.WriteRune(r)
	}

	if buf.Len() > 0 {
		words = append(words, buf.String())
	}

	return words
}

func (c CaseType) Apply(filename string) string {
	if c == CaseNone {
		return filename
	}

	ext := filepath.Ext(filename)
	name := strings.TrimSuffix(filename, ext)

	words := splitWords(name)
	if len(words) == 0 {
		return filename
	}

	cleaned := make([]string, len(words))
	switch c {
	case CaseSnake:
		for i, w := range words {
			cleaned[i] = strings.ToLower(w)
		}
		return strings.Join(cleaned, "_") + ext
	case CaseKebab:
		for i, w := range words {
			cleaned[i] = strings.ToLower(w)
		}
		return strings.Join(cleaned, "-") + ext
	case CaseCamel:
		for i, w := range words {
			if i == 0 {
				cleaned[i] = strings.ToLower(w)
			} else {
				cleaned[i] = titleWord(w)
			}
		}
		return strings.Join(cleaned, "") + ext
	case CasePascal:
		for i, w := range words {
			cleaned[i] = titleWord(w)
		}
		return strings.Join(cleaned, "") + ext
	case CaseUpper:
		for i, w := range words {
			cleaned[i] = strings.ToUpper(w)
		}
		return strings.Join(cleaned, "_") + ext
	case CaseLower:
		for i, w := range words {
			cleaned[i] = strings.ToLower(w)
		}
		return strings.Join(cleaned, " ") + ext
	case CaseTitle:
		for i, w := range words {
			cleaned[i] = titleWord(w)
		}
		return strings.Join(cleaned, " ") + ext
	default:
		return filename
	}
}

func titleWord(w string) string {
	runes := []rune(w)
	for i, r := range runes {
		if i == 0 {
			runes[i] = unicode.ToTitle(r)
		} else {
			runes[i] = unicode.ToLower(r)
		}
	}
	return string(runes)
}

type Renamer struct {
	client   llm.Client
	model    string
	hints    []string
	caseType CaseType
}

func NewRenamer(client llm.Client, model string, hints []string) *Renamer {
	return &Renamer{
		client: client,
		model:  model,
		hints:  append([]string(nil), hints...),
	}
}

func NewCaseRenamer(caseType CaseType) *Renamer {
	return &Renamer{
		caseType: caseType,
	}
}

func (r *Renamer) SetCaseType(c CaseType) {
	r.caseType = c
}

func (r *Renamer) Decide(ctx context.Context, files []core.FileEntry) ([]core.FileOperation, error) {
	if len(files) == 0 {
		return nil, nil
	}

	useLLM := len(r.hints) > 0 && r.client != nil

	var llmFiles []core.FileEntry
	var ops []core.FileOperation

	if useLLM {
		for _, f := range files {
			if isBinaryExt(f.SourcePath) {
				ops = append(ops, core.FileOperation{OpType: core.OpSkip, File: f})
			} else {
				llmFiles = append(llmFiles, f)
			}
		}
		if len(llmFiles) == 0 {
			return ops, nil
		}
	}

	var newNames []string

	if useLLM {
		var err error
		newNames, err = r.renameWithLLM(ctx, llmFiles)
		if err != nil {
			return nil, err
		}
	} else {
		newNames = make([]string, len(files))
		for i, f := range files {
			newNames[i] = filepath.Base(f.SourcePath)
		}
	}

	if r.caseType != CaseNone {
		for i, n := range newNames {
			newNames[i] = r.caseType.Apply(n)
		}
	}

	seen := make(map[string]bool)

	if useLLM {
		buildOps := func(files []core.FileEntry, names []string) []core.FileOperation {
			out := make([]core.FileOperation, 0, len(files))
			for i, newName := range names {
				if hasPathTraversal(newName) {
					out = append(out, core.FileOperation{OpType: core.OpSkip, File: files[i]})
					continue
				}

				originalName := filepath.Base(files[i].SourcePath)
				if newName == originalName {
					out = append(out, core.FileOperation{OpType: core.OpSkip, File: files[i]})
					continue
				}

				newName = resolveCollision(newName, seen)
				seen[newName] = true
				destPath := filepath.Join(filepath.Dir(files[i].SourcePath), newName)

				if _, err := os.Stat(destPath); err == nil {
					out = append(out, core.FileOperation{OpType: core.OpSkip, File: files[i]})
					continue
				}

				out = append(out, core.FileOperation{
					OpType:   core.OpRename,
					File:     files[i],
					DestPath: destPath,
					Size:     files[i].Size,
				})
			}
			return out
		}
		ops = append(ops, buildOps(llmFiles, newNames)...)
	} else {
		for i, newName := range newNames {
			originalName := filepath.Base(files[i].SourcePath)

			if r.caseType != CaseNone && newName == originalName {
				ops = append(ops, core.FileOperation{OpType: core.OpSkip, File: files[i]})
				continue
			}

			newName = resolveCollision(newName, seen)
			seen[newName] = true
			destPath := filepath.Join(filepath.Dir(files[i].SourcePath), newName)

			if _, err := os.Stat(destPath); err == nil {
				ops = append(ops, core.FileOperation{OpType: core.OpSkip, File: files[i]})
				continue
			}

			ops = append(ops, core.FileOperation{
				OpType:   core.OpRename,
				File:     files[i],
				DestPath: destPath,
				Size:     files[i].Size,
			})
		}
	}

	return ops, nil
}

func resolveCollision(name string, seen map[string]bool) string {
	ext := filepath.Ext(name)
	base := strings.TrimSuffix(name, ext)
	counter := 1
	for seen[name] {
		name = fmt.Sprintf("%s_v%d%s", base, counter, ext)
		counter++
	}
	return name
}

func (r *Renamer) renameWithLLM(ctx context.Context, files []core.FileEntry) ([]string, error) {
	allNewNames := make([]string, 0, len(files))
	batchSize := 10

	for i := 0; i < len(files); i += batchSize {
		end := i + batchSize
		end = min(len(files), end)

		batch := files[i:end]
		batchFilenames := make([]string, len(batch))
		for j, f := range batch {
			batchFilenames[j] = filepath.Base(f.SourcePath)
		}

		marshalledPayload, err := json.Marshal(batchFilenames)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal filenames: %w", err)
		}

		userPrompt := ""
		if len(r.hints) > 0 {
			hintsJSON, err := json.Marshal(r.hints)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal rename hints: %w", err)
			}
			userPrompt += "USER_HINTS: " + string(hintsJSON) + "\n"
		}
		userPrompt += string(marshalledPayload) + "\nOutput the results as a JSON object."

		raw, err := retryLLM(ctx, func() (string, error) {
			return r.client.Run(ctx, llm.Request{
				Model:        r.model,
				SystemPrompt: defaultPrompt + "\n\nCRITICAL: Return exactly one JSON object with a 'filenames' key containing the renamed strings and nothing else.",
				UserPrompt:   userPrompt,
			})
		})
		if err != nil {
			return nil, fmt.Errorf("rename request failed for batch starting at %d after retries: %w", i, err)
		}
		raw = strings.TrimSpace(raw)

		var response struct {
			Filenames []string `json:"filenames"`
		}
		if err := json.Unmarshal([]byte(raw), &response); err != nil {
			return nil, fmt.Errorf("failed to parse AI response for batch starting at %d: %w. Raw output: %s", i, err, raw)
		}

		if len(response.Filenames) != len(batch) {
			return nil, fmt.Errorf("integrity error in batch starting at %d: sent %d files, received %d names. Raw output: %s", i, len(batch), len(response.Filenames), raw)
		}

		allNewNames = append(allNewNames, response.Filenames...)
	}

	return allNewNames, nil
}

type SelectionSorter struct {
	renamer *Renamer
	allowed map[string]struct{}
}

func NewSelectionSorter(files []string, hints []string) (*SelectionSorter, error) {
	client, err := llm.NewClient(llm.DefaultModel)
	if err != nil {
		return nil, err
	}
	allowed := make(map[string]struct{}, len(files))
	for _, file := range files {
		allowed[filepath.Clean(file)] = struct{}{}
	}
	return &SelectionSorter{
		renamer: NewRenamer(client, llm.DefaultModel, hints),
		allowed: allowed,
	}, nil
}

func retryLLM(ctx context.Context, fn func() (string, error)) (string, error) {
	var lastErr error
	backoff := []time.Duration{1 * time.Second, 2 * time.Second, 4 * time.Second}

	for attempt := 0; attempt <= len(backoff); attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return "", ctx.Err()
			case <-time.After(backoff[attempt-1]):
			}
		}

		result, err := fn()
		if err == nil {
			return result, nil
		}
		lastErr = err
	}

	return "", lastErr
}

func (s *SelectionSorter) Decide(ctx context.Context, files []core.FileEntry) ([]core.FileOperation, error) {
	if len(s.allowed) == 0 {
		return s.renamer.Decide(ctx, files)
	}

	selected := make([]core.FileEntry, 0, len(files))
	ops := make([]core.FileOperation, 0, len(files))
	for _, file := range files {
		rel, err := filepath.Rel(file.RootDir, file.SourcePath)
		if err != nil {
			return nil, err
		}
		if _, ok := s.allowed[filepath.Clean(rel)]; ok {
			selected = append(selected, file)
			continue
		}
		ops = append(ops, core.FileOperation{OpType: core.OpSkip, File: file})
	}

	renames, err := s.renamer.Decide(ctx, selected)
	if err != nil {
		return nil, err
	}
	ops = append(ops, renames...)
	return ops, nil
}
