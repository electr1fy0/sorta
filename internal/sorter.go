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
	ops := make([]FileOperation, 0, 20)
	filenames := make([]string, 0, 20)
	newnames := make([]string, 0, 20)

	for _, filename := range filePaths {
		filenames = append(filenames, filename.Filename)
	}
	prompt := `Return the **exact same JSON structure**.
Only modify the **final filename segment** of each path.

Rules for renaming:
- Each filename must be **unique** and must reflect the file’s actual purpose or meaning. The more specific the better.
- Remove generic or redundant clutter from filenames.
- If multiple files have the same substring and you thing it does not help much shorten that part.
- If a filename has redundant info in one way or another. Strip off the redundancy. e.g. Repeated mentions of a year, name, subject, title, etc.
- Given College/School documents, try to trim down the Semester/Class year name / number to a more concise representation.
- Each file should have its unique determiner.
- Do not change JSON structure.
- If the name is too long, definitely try to shorten it.
- Do not add any new fields.
- Have a consistent usage of _ or - or spaces all across.
- If filename really does not need to be changed, at least style it better.
- All filenames should have a similar consistent modern styling.
- If you cannot safely rename, return the JSON unchanged.


Output **only** the raw JSON as a plain string. Nothing else. Don't even add the code blocks or syntax highlight at all. Zero formatting.`

	marshalled, _ := json.Marshal(filenames)

	ctx := context.Background()

	client, _ := genai.NewClient(ctx, nil)

	resp, _ := client.Models.GenerateContent(ctx, "gemini-2.5-flash", genai.Text(prompt+"\n"+string(marshalled)), nil)

	raw := resp.Text()

	raw = strings.TrimSpace(raw)
	raw = strings.TrimPrefix(raw, "```json")
	raw = strings.TrimPrefix(raw, "```")
	raw = strings.TrimPrefix(raw, "```")

	err := json.Unmarshal([]byte(raw), &newnames)
	if err != nil {
		fmt.Println(err)

	}

	for i, filePath := range filePaths {
		op := FileOperation{OpMove, filepath.Join(filePath.FullDir, filePath.Filename), filepath.Join(filePaths[i].FullDir, newnames[i]), newnames[i], filePath.Size}
		ops = append(ops, op)
	}
	return ops, err
}

func NewDuplicateFinder() *DuplicateFinder {
	return &DuplicateFinder{
		hashes: make(map[string]string),
	}
}

func NewRenamer() *Renamer {
	return &Renamer{}
}

func (d *DuplicateFinder) Sort(BaseDir, dir, filename string, size int64) (FileOperation, error) {
	fullPath := filepath.Join(dir, filename)
	if fullPath == filepath.Join(BaseDir, "duplicates", filename) {
		return FileOperation{
			Type: OpSkip,
		}, nil
	}

	data, err := os.ReadFile(fullPath)
	if err != nil {
		return FileOperation{Type: OpSkip}, err
	}

	checksum := fmt.Sprintf("%x", sha256.Sum256(data))
	if _, exists := d.hashes[checksum]; !exists {
		d.hashes[checksum] = fullPath
		return FileOperation{
			Type: OpSkip}, nil
	}

	return FileOperation{
		Type:       OpMove,
		SourcePath: fullPath,
		DestPath:   filepath.Join(BaseDir, "duplicates", filename),
		Filename:   filename,
		Size:       size,
	}, nil
}
