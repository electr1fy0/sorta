package internal

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"google.golang.org/genai"
)

func NewExtensionSorter() *ExtensionSorter {
	return &ExtensionSorter{
		categories: map[string][]string{
			"docs":   {".pdf", ".docx", ".pages", ".md", ".txts"},
			"images": {".png", ".jpg", ".jpeg", ".heic", ".heif"},
			"movies": {".mp4", ".mov"},
			"slides": {".pptx"},
		},
	}
}

func (s *ExtensionSorter) Sort(BaseDir, dir, filename string, size int64) (FileOperation, error) {
	ext := strings.ToLower(filepath.Ext(filename))

	for folder, extensions := range s.categories {
		if slices.Contains(extensions, ext) {
			return FileOperation{
				Type:       OpMove,
				SourcePath: filepath.Join(dir, filename),
				DestPath:   filepath.Join(BaseDir, folder, filename),
				Filename:   filename,
				Size:       size,
			}, nil
		}
	}

	return FileOperation{Type: OpSkip}, nil
}

func NewConfigSorter() (*ConfigSorter, error) {
	confData, err := ParseConfig()
	if err != nil {
		return nil, err
	}
	return &ConfigSorter{
		configData: confData,
	}, nil
}

func (s *ConfigSorter) Sort(filePaths []FilePath) ([]FileOperation, error) {
	ops := make([]FileOperation, 0, 10)

	for _, filePath := range filePaths {
		srcPath := filepath.Join(filePath.FullDir, filePath.Filename)
		folder := categorize(*s.configData, filePath.Filename, filepath.Ext(srcPath))

		if folder == "" {
			ops = append(ops, FileOperation{Type: OpSkip})
		} else {
			ops = append(ops, FileOperation{
				Type:       OpMove,
				SourcePath: srcPath,
				DestPath:   filepath.Join(filePath.BaseDir, folder, filePath.Filename),
				Filename:   filePath.Filename,
				Size:       filePath.Size,
			})
		}
	}
	return ops, nil
}

func (r *Renamer) Sort(filePaths []FilePath) ([]FileOperation, error) {
	filenames := make([]string, len(filePaths))
	for i, f := range filePaths {
		filenames[i] = f.Filename
	}

	marshalledPayload, err := json.Marshal(filenames)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal filenames: %w", err)
	}

	prompt := `You are an intelligent file renaming engine. Your goal is to transform filenames to be concise, meaningful, and machine-friendly (snake_case).

Input: A JSON array of filename strings.
Output: A JSON array of transformed filename strings.

### CRITICAL OUTPUT RULES (Zero Tolerance):
1. Return **ONLY** the raw JSON array. No Markdown, no code blocks, no explanations.
2. The output array MUST have the exact same number of elements as the input.
3. **Uniqueness is Mandatory:** If two files reduce to the same name, you MUST append a discriminator (e.g., "_v1", "_v2", or keep the original number).

### TRANSFORMATION LOGIC (Execute in Order):

1. **Standardization:**
   - Convert to 'snake_case' (lowercase, underscores).
   - Replace spaces, hyphens, and dots (except the extension dot) with underscores.
   - Remove special characters like parentheses '()'.

2. **Semantic Cleaning (The "Human" Rule):**
   - **Remove Clutter:** Strip generic words that add no value: "copy", "final", "draft", "new", "converted", "document", "file".
   - **Remove Redundancy:** If a word (like a year, subject, or name) appears multiple times in the string, keep only one instance.
   - **Focus on Purpose:** Ensure the filename reflects *what* the file is. If the name is extremely long, shorten it to the 3-4 most significant keywords.

3. **Academic & Technical Abbreviations (Strict Mapping):**
   - "Assignment" -> "asn"
   - "Experiment" / "Exp" -> "exp"
   - "Laboratory" / "Lab" -> "lab"
   - "Semester" / "Sem" -> "s" (e.g., "sem_05" -> "s5")
   - "Project" -> "proj"
   - "Syllabus" -> "syl"
   - "Question Paper" / "QP" -> "qp"
   - "Introduction" -> "intro"
   - Years: "2024-2025" -> "24_25", "2024" -> "24"

4. **Safety:**
   - NEVER change the file extension.

### FEW-SHOT EXAMPLES (Mimic this style):

Input:  ["Copy of Operating Systems Sem 5 Final Notes (2024).pdf", "Data_Structures_Assignment_1_FINAL_v2.docx"]
Output: ["os_s5_notes_24.pdf", "dsa_asn_1_v2.docx"]

Input:  ["project_report_final_final_print.pdf", "my_resume_engineering_2025_updated.pdf"]
Output: ["proj_report_print.pdf", "resume_eng_25.pdf"]

Input:  ["lab_experiment_1.txt", "lab_experiment_2.txt"]
Output: ["lab_exp_1.txt", "lab_exp_2.txt"]

PAYLOAD:`

	ctx := context.Background()
	client, err := genai.NewClient(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create genai client: %w", err)
	}

	resp, err := client.Models.GenerateContent(ctx, "gemini-2.5-flash-lite", genai.Text(prompt+"\n"+string(marshalledPayload)), nil)
	if err != nil {
		return nil, fmt.Errorf("gemini request failed: %w", err)
	}

	raw := resp.Text()
	raw = strings.TrimSpace(raw)

	var newnames []string
	if err := json.Unmarshal([]byte(raw), &newnames); err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w. Raw output: %s", err, raw)
	}

	if len(newnames) != len(filePaths) {
		return nil, fmt.Errorf("integrity error: sent %d files, received %d names", len(filePaths), len(newnames))
	}

	ops := make([]FileOperation, 0, len(filePaths))
	seen := make(map[string]bool)

	for i, newName := range newnames {
		if strings.TrimSpace(newName) == "" {
			newName = filenames[i]
		}

		base := newName
		ext := filepath.Ext(newName)
		nameNoExt := strings.TrimSuffix(base, ext)
		counter := 1

		for seen[newName] {
			newName = fmt.Sprintf("%s_v%d%s", nameNoExt, counter, ext)
			counter++
		}
		seen[newName] = true

		op := FileOperation{
			Type:       OpMove,
			SourcePath: filepath.Join(filePaths[i].FullDir, filePaths[i].Filename),
			DestPath:   filepath.Join(filePaths[i].FullDir, newName),
			Filename:   newName,
			Size:       filePaths[i].Size,
		}
		ops = append(ops, op)
	}

	return ops, nil
}

func NewDuplicateFinder() *DuplicateFinder {
	return &DuplicateFinder{
		hashes: make(map[string]string),
	}
}

func NewRenamer() *Renamer {
	return &Renamer{}
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
