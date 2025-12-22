package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"google.golang.org/genai"
)

const defaultPrompt = `You are an intelligent file renaming engine. Your goal is to transform filenames to be concise, meaningful, and machine-friendly using Title_Snake_Case.

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

2. **Smart Shortening (The "Key-Bit" Rule):
	- **Identify Value:** If the filename contains more than 6 meaningful tokens, remove generic or
	  structural prefixes ("Department", "University", "Institute", "College", "School",
	  "Faculty", "Part", "Chapter", "Section", "Introduction") from the beginning until a
	  subject/topic/technical token appears.
	- **Keep Specifics:** Extract and keep only the **unique subject**, **topic**, or **specific noun** that describes the file content.
	- **Remove Clutter:** Remove tokens such as "Copy", "Copy of", "copy_of", "final", "draft",
	    "new", "document", "file", "download", "backup", "(1)", "(2)", "(copy)" and
	    other software-generated prefixes (e.g. "Microsoft Word - ").
	- Remove numeric tokens of length ≥6 unless they represent a valid date (YYYYMMDD or YYYY-MM-DD).
	- **Remove Redundancy:** If a word (year/subject) repeats, keep only one.
	- After applying abbreviations, remove any immediate repeated tokens (e.g., "DSA_DSA" → "DSA").
	- **Preserve Semester Tokens:** If a token contains "SEM" or matches patterns like
	  FALLSEM*, WINTERSEM*, FALL_S*, WINTER_S*, or Sem/sem + digits, do NOT remove it during shortening.
	  Instead, compress it into the official semester form (e.g., FALLSEM2025-26 → F25-26,
	  WINTERSEM2025-26 → W25-26, sem_05 → S5). Never drop these tokens entirely.

3. **Strict Abbreviations (Use These Exact Forms):
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

func NewRenamer() *Renamer {
	return &Renamer{}
}

func (r *Renamer) Decide(files []FileEntry) ([]FileOperation, error) {
	if len(files) == 0 {
		return nil, nil
	}

	filenames := make([]string, len(files))
	for i, f := range files {
		filenames[i] = filepath.Base(f.SourcePath)
	}

	marshalledPayload, err := json.Marshal(filenames)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal filenames: %w", err)
	}

	prompt := getPrompt(files[0].RootDir)

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

	if len(newnames) != len(files) {
		return nil, fmt.Errorf("integrity error: sent %d files, received %d names", len(files), len(newnames))
	}

	ops := make([]FileOperation, 0, len(files))
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
		destPath := filepath.Join(filepath.Dir(files[i].SourcePath), newName)

		op := FileOperation{
			OpType:   OpMove,
			File:     files[i],
			DestPath: filepath.Join(files[i].SourcePath, destPath),
			Size:     files[i].Size,
		}
		ops = append(ops, op)
	}

	return ops, nil
}

func getPrompt(localDir string) string {
	localPromptPath := filepath.Join(localDir, ".sorta", "prompt")
	if content, err := os.ReadFile(localPromptPath); err == nil && len(content) > 0 {
		return string(content)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return defaultPrompt
	}
	promptPath := filepath.Join(home, ".sorta", "prompt")

	if _, err := os.Stat(promptPath); os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(promptPath), 0755); err != nil {
			return defaultPrompt
		}
		if err := os.WriteFile(promptPath, []byte(defaultPrompt), 0644); err != nil {
			return defaultPrompt
		}
		return defaultPrompt
	}

	content, err := os.ReadFile(promptPath)
	if err != nil {
		return defaultPrompt
	}
	if len(content) == 0 {
		return defaultPrompt
	}
	return string(content)
}
