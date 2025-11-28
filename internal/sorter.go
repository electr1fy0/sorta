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

	prompt := `You are an intelligent file renaming engine. Your goal is to transform filenames to be concise, meaningful, and machine-friendly using Title_Snake_Case.

Input: A JSON array of filename strings.
Output: A JSON array of transformed filename strings.

### CRITICAL OUTPUT RULES:
1. Return **ONLY** the raw JSON array. No Markdown.
2. The output array MUST have the exact same number of elements as the input.
3. **Uniqueness:** If two files reduce to the same name, append a discriminator (e.g., "_v1").

### RENAMING LOGIC:

1. **Standardization (Title Case):**
   - **Format:** Capitalize the first letter of every significant word (Title_Case).
   - **Separators:** Use underscores only. Replace spaces, hyphens, and dots with underscores.
   - **CamelCase:** Split attached words (e.g., "MyProjectFile" -> "My_Project_File").
   - **Acronyms:** Keep technical terms in ALL CAPS (e.g., "OS", "DSA", "TCP", "AI", "DBMS", "LAB", "ID", "API", "JSON").

2. **Smart Shortening (The "Key-Bit" Rule):**
   - **Identify Value:** If a filename is too long (>6 words), ignore generic hierarchy (e.g., "Department of...", "Chapter 3...", "Part 2...").
   - **Keep Specifics:** Extract and keep only the **unique subject**, **topic**, or **specific noun** that describes the file content.
   - **Remove Clutter:** "Copy of", "final", "draft", "new", "document", "file", "download", "backup" (unless it's the only word).
   - **Remove Redundancy:** If a word (year/subject) repeats, keep only one.

3. **Strict Abbreviations (Use These Exact Forms):**
   - "Assignment" -> "Asn"
   - "Experiment" -> "Exp"
   - "Laboratory" -> "Lab"
   - "Semester" / "Sem" -> "s" (e.g., "sem_05" -> "S5")
   - "Project" -> "Proj"
   - "Syllabus" -> "Syl"
   - "Question Paper" -> "QP"
   - "Introduction" -> "Intro"
   - "Manual" -> "Man"
   - Years: "2024-2025" -> "24_25", "2024" -> "24"

4. **Safety:**
   - NEVER change the file extension.

### EXAMPLES (Study these patterns):

// 1. Standard Academic (Title Case + Acronyms)
Input:  ["Copy of Operating Systems Sem 5 Final Notes (2024).pdf", "data_structures_assignment_1_FINAL_v2.docx"]
Output: ["OS_s5_Notes_24.pdf", "DSA_Asn_1_v2.docx"]

// 2. Intelligent Shortening (Picking the Important Bits)
Input:  ["Vellore_Institute_of_Technology_Fall_2025_Network_Security_Course_Plan.docx"]
Output: ["Network_Security_Course_Plan_25.docx"]  // Dropped Uni name, kept Subject

Input:  ["Introduction_to_Computer_Science_Part_3_Advanced_Data_Structures_and_Algorithms_Notes.pdf"]
Output: ["Advanced_DSA_Notes.pdf"] // Dropped generic intro, kept specific topic

Input:  ["Department_of_Mechanical_Engineering_Fluid_Mechanics_Lab_Manual_v2.pdf"]
Output: ["Fluid_Mechanics_Lab_Man_v2.pdf"]

// 3. Messy Separators & CamelCase
Input:  ["MyProjectFile_Final_v2.java", "web-development-lab-experiment-1.html", "Abstract_Algebra---Notes.pdf"]
Output: ["My_Proj_File_v2.java", "Web_Dev_Lab_Exp_1.html", "Abstract_Algebra_Notes.pdf"]

// 4. Dates & Receipts
Input:  ["Invoice_2024-12-01_paid.pdf", "receipt jan 24 grocery.jpeg", "Statement 01012025.pdf"]
Output: ["Invoice_2024_12_01_Paid.pdf", "Receipt_Jan_24_Grocery.jpeg", "Statement_01_01_2025.pdf"]

// 5. Developer & Technical Files
Input:  ["main_backup_copy.go", "docker-compose-dev.yml", "API_ENDPOINT_SPEC_v3.json"]
Output: ["Main_Backup.go", "Docker_Compose_Dev.yml", "API_Endpoint_Spec_v3.json"]

// 6. Messy "Downloads" Garbage
Input:  ["scan_29384_Contract_Signed.pdf", "Microsoft Word - Resume_John_Doe_2025.docx"]
Output: ["Contract_Signed.pdf", "Resume_John_Doe_25.docx"]

// 7. Versioning Conflicts
Input:  ["lab_experiment_1.txt", "lab_experiment_1 (1).txt", "lab_experiment_1 (2).txt"]
Output: ["Lab_Exp_1.txt", "Lab_Exp_1_v1.txt", "Lab_Exp_1_v2.txt"]

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
