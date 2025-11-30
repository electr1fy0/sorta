package internal

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

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

func NewConfigSorter(folderPath, configPath string) (*ConfigSorter, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("error determining home directory: %w", err)
	}
	defaultPath := filepath.Join(home, ".sorta", "config")
	var localPath string
	if configPath == defaultPath {
		localPath = filepath.Join(folderPath, ".sorta", "config")
	}

	var confData *ConfigData
	_, err = os.Open(localPath)
	if err != nil {
		if os.IsNotExist(err) {
			confData, err = ParseConfig(configPath)
		} else {
			return nil, err
		}
	} else {
		confData, err = ParseConfig(localPath)
	}

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

func (s *ConfigSorter) GetBlacklist() []string {
	return s.configData.Blacklist
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
1. Return **ONLY** the raw JSON array. No Markdown, no prose, no explanation, no code fences.
2. The output array MUST have the exact same number of elements as the input.
3. Preserve array order exactly. Do not sort or reorder entries.
4. **Uniqueness:** If two transformed names collide, append "_v1", "_v2", ... to the later items. Keep any original version markers; append only if collision still remains.
5. Never modify or remove the file extension. Copy it exactly from input.

### RENAMING LOGIC:
Apply the following rules strictly in the order they appear. Do not skip steps or reorder steps.
All transformations must be derived strictly from the input tokens. Do not invent, guess, or infer content.

1. **Standardization (Title Case):**
   - **Format:** Capitalize the first letter of every significant word (Title_Case).
   - Title Case means: Capitalize only the first letter of each word, keep the rest lowercase (unless acronym/abbreviation rules override).
   - **Separators:** Use underscores only. Replace spaces, hyphens, and dots with underscores.
   - **CamelCase:** Split attached words (e.g., "MyProjectFile" -> "My_Project_File"). Apply camelCase splitting before any other replacement steps.
   - **Acronyms:** If a token (match whole token, case-insensitive) equals any of: OS, DSA, TCP, AI, DBMS, LAB, ID, API, JSON — output it in ALL CAPS.
     Perform acronym normalization **before** applying Title Case. Do not transform these acronyms when they appear as substrings inside larger words.

2. **Smart Shortening (The "Key-Bit" Rule):**
	- **Identify Value:** If the filename contains more than 6 meaningful tokens, remove generic or
	  structural prefixes ("Department", "University", "Institute", "College", "School",
	  "Faculty", "Part", "Chapter", "Section", "Introduction") from the beginning until a
	  subject/topic/technical token appears.
	- **Keep Specifics:** Extract and keep only the **unique subject**, **topic**, or **specific noun** that describes the file content.
	- **Remove Clutter:** Remove tokens such as "Copy", "Copy of", "copy_of", "final", "draft",
	    "new", "document", "file", "download", "backup", "(1)", "(2)", "(copy)" and
	    other software-generated prefixes (e.g. "Microsoft Word -").
	- Remove numeric tokens of length ≥6 unless they represent a valid date (YYYYMMDD or YYYY-MM-DD).
	- **Remove Redundancy:** If a word (year/subject) repeats, keep only one.
	- After applying abbreviations, remove any immediate repeated tokens (e.g., "DSA_DSA" → "DSA").
	- **Preserve Semester Tokens:** If a token contains "SEM" or matches patterns like
	  FALLSEM*, WINTERSEM*, FALL_S*, WINTER_S*, or Sem/sem + digits, do NOT remove it during shortening.
	  Instead, compress it into the official semester form (e.g., FALLSEM2025-26 → F25-26,
	  WINTERSEM2025-26 → W25-26, sem_05 → S5). Never drop these tokens entirely.

3. **Strict Abbreviations (Use These Exact Forms):**
   - "Assignment" -> "Asn"
   - "Experiment" -> "Exp"
   - "Laboratory" -> "Lab"
   - "Semester" / "Sem" -> "S" (e.g., "sem_05" -> "S5", "FALLSEM2025-26" -> "F25-26", "Fall_s25-26" -> "F25-26", "WINTERSEM2025-26"->"W25-26")
   - "Project" -> "Proj"
   - "Syllabus" -> "Syl"
   - "Question Paper" -> "QP"
   - "Introduction" -> "Intro"
   - Years: For ranges ("2024-2025"), output "2024_25". Keep original ordering.
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

// 8. Unnecessary metadata
Input: ["Mooc_File_VL374892378278273_FAT-1"]
Output:["MOOC_File_FAT-1.pdf"]

PAYLOAD:`

	home, _ := os.UserHomeDir()
	promptPath := filepath.Join(home, ".sorta", "prompt")
	promptFile, err := os.Open(promptPath)
	if err != nil {
		if os.IsNotExist(err) {
			os.WriteFile(promptPath, []byte(prompt), 0644)
			time.Sleep(200 * time.Millisecond)
		}
	}

	filePromptBytes, _ := io.ReadAll(promptFile)
	if string(filePromptBytes) != "" {
		prompt = string(filePromptBytes)
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create genai client: %w", err)
	}
	status := make(chan struct{})
	go func() {
		dots := []string{"", ".", "..", "..."}
		i := 0

		for {
			select {
			case <-status:
				return

			default:
				fmt.Printf("\rConversing with Gemini%s     ", dots[i])
				i = (i + 1) % len(dots)
				time.Sleep(200 * time.Millisecond)
			}
		}
	}()

	resp, err := client.Models.GenerateContent(ctx, "gemini-2.5-flash-lite", genai.Text(prompt+"\n"+string(marshalledPayload)), nil)
	close(status)
	fmt.Print("\r                             \n")

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
