package internal

import (
	"crypto/sha256"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	ansiReset  = "\033[0m"
	ansiRed    = "\033[31m"
	ansiGreen  = "\033[32m"
	ansiYellow = "\033[33m"
	ansiCyan   = "\033[36m"
)

type ConfigData map[string][]string

type Sorter interface {
	Sort(dir string, filename string, size int64) (FileOperation, error)
}

type OperationType int

const (
	OpMove OperationType = iota
	OpDelete
	OpSkip
)

type FileOperation struct {
	Type       OperationType
	SourcePath string
	DestPath   string
	Filename   string
	Size       int64
}

type Executor struct {
	DryRun bool
}

func (e *Executor) Execute(op FileOperation) error {
	if e.DryRun {
		return nil
	}

	switch op.Type {
	case OpMove:
		os.MkdirAll(filepath.Dir(op.DestPath), 0755)
		return os.Rename(op.SourcePath, op.DestPath)
	case OpDelete:
		os.Remove(op.SourcePath)
	}

	return nil
}

type Reporter struct {
	DryRun bool
}

func (r *Reporter) Report(op FileOperation, err error) {
	prefix := ansiGreen + "[OK]" + ansiReset
	if r.DryRun {
		prefix = ansiYellow + "[DRY]" + ansiReset
	}

	if err != nil {
		prefix = ansiRed + "[ERR]" + ansiReset
	}

	switch op.Type {
	case OpMove:
		fmt.Printf("%s %s -> %s (%s)\n", prefix, op.Filename, op.DestPath, humanReadable(op.Size))
	case OpSkip:
		break
	}
}

type ExtensionSorter struct {
	categories map[string][]string
}

func NewExtensionSorter() *ExtensionSorter {
	return &ExtensionSorter{
		categories: map[string][]string{
			"docs":   {".pdf", ".docx", ".pages", ".md", ".txts"},
			"images": {".png", ".jpg", ".jpeg", ".heic", ".heif"},
			"movies": {".mp4", ".mov"},
		},
	}
}

func (s *ExtensionSorter) Sort(dir, filename string, size int64) (FileOperation, error) {
	ext := strings.ToLower(filepath.Ext(filename))

	for folder, extensions := range s.categories {
		for _, validExt := range extensions {
			if ext == validExt {
				return FileOperation{
					Type:       OpMove,
					SourcePath: filepath.Join(dir, filename),
					DestPath:   filepath.Join(dir, folder, filename),
					Filename:   filename,
					Size:       size,
				}, nil
			}
		}
	}

	return FileOperation{Type: OpSkip}, nil
}

type ConfigSorter struct {
	configData ConfigData
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

func (s *ConfigSorter) Sort(dir, filename string, size int64) (FileOperation, error) {
	folder := categorize(s.configData, filename)

	if folder == "" {
		return FileOperation{Type: OpSkip}, nil
	}

	return FileOperation{
		Type:       OpMove,
		SourcePath: filepath.Join(dir, filename),
		DestPath:   filepath.Join(dir, folder, filename),
		Filename:   filename,
		Size:       size,
	}, nil
}

type DuplicateFinder struct {
	hashes map[string]string
}

func NewDuplicateFinder() *DuplicateFinder {
	return &DuplicateFinder{
		hashes: make(map[string]string),
	}
}

func (d *DuplicateFinder) Sort(dir, filename string, size int64) (FileOperation, error) {
	fullPath := filepath.Join(dir, filename)

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
		DestPath:   filepath.Join(dir, "duplicates", filename),
		Filename:   filename,
		Size:       size,
	}, nil
}

type SortResult struct {
	Moved   int
	Skipped int
	Errors  []error
}

func (r *SortResult) Print() {
	fmt.Println("Moved:", r.Moved)
	fmt.Println("Skipped:", r.Skipped)
	fmt.Println("Errors: i don't know how to count them yet so 0 errors for now")
}

func readConfigFile() (string, error) {
	home, _ := os.UserHomeDir()
	configName := ".sorta-config"
	configBytes, err := os.ReadFile(filepath.Join(home, configName))
	if err != nil {
		err = createConfig()
		if err != nil {
			return "", err
		}
		fmt.Println("Config file is empty. Add keywords to .sorta-config in home directory") // i'm lying here
		os.Exit(1)
	}

	config := string(configBytes)
	if strings.TrimSpace(config) == "" {
		fmt.Println("Config file is empty. Add keywords to .sorta-config in home directory")

		os.Exit(1)
	}
	return config, nil
}

func ParseConfig() (ConfigData, error) {
	var configData ConfigData
	config, err := readConfigFile()
	if err != nil {
		return configData, err
	}

	lineCount := strings.Count(config, "\n")
	if []rune(config)[len(config)-1] != rune('\n') {
		lineCount++
	}

	configData = make(map[string][]string)

	for line := range strings.Lines(config) {
		if strings.HasPrefix(line, "//") {
			continue
		}
		input := strings.Split(line, ",")
		last := strings.TrimSpace(input[len(input)-1])

		lastSplit := strings.Split(last, "=")

		input[len(input)-1] = lastSplit[0]
		output := lastSplit[1]
		configData[output] = input
	}

	return configData, nil
}

func createConfig() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	content := []byte(`// Config file for 'sorta'
//
// Each line defines how files should be sorted.
// Format: key1,key2,key3 = folderName
//
// - key1, key2, key3, etc are keywords to match in file names.
// - folderName is the target folder for those files.
// - You can list one or many keywords before the '='.
// - Lines starting with '//' are comments and ignored.
// - Make sure no spaces exist between the keys and values
// - * as a keyword matches all filenames which don't contain the other keywords
// Example:
// invoice,bill,txt=Finance
// track,song=Music
// notes,book=Study
// *=others`)

	path := filepath.Join(home, ".sorta-config")

	err = os.WriteFile(path, content, 0600)
	if err != nil {
		return err
	}
	return nil
}

type State struct {
	CliDir              string
	DryRun              bool
	MoveCnt, skippedCnt int
}

type FileInfo struct {
	Name string
	Size int64
}

func createFolder(dir, foldername string) error {
	return os.MkdirAll(filepath.Join(dir, foldername), 0700)
}

func handleMove(dir, foldername, filename string, size int64) error {
	srcPath := filepath.Join(dir, filename)
	destDir := filepath.Join(dir, foldername)
	destPath := filepath.Join(destDir, filename)

	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("failed to create folder %s: %w", destDir, err)
	}

	if err := os.Rename(srcPath, destPath); err != nil {
		return fmt.Errorf("failed to move %s → %s: %w", srcPath, destPath, err)
	}

	fmt.Printf("%s %s → %s (%s)\n", ansiGreen+"[OK]"+ansiReset, filename, destPath, humanReadable(size))
	return nil
}

func FilterFiles(dir string, sorter Sorter, executor *Executor, reporter *Reporter) (SortResult, error) {
	// hashes := make(map[string]string)
	result := SortResult{}

	_ = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil // skip directories
		}

		filename := d.Name()
		if strings.HasPrefix(filename, ".") {
			return nil // skip hidden files
		}

		stat, err := d.Info()
		if err != nil {
			return err
		}
		size := stat.Size()

		fileOp, err := sorter.Sort(dir, filename, size)
		err = executor.Execute(fileOp)

		if fileOp.Type == OpMove {
			result.Moved++
		} else {
			result.Skipped++
		}
		return nil
	})

	return result, nil

}
func TopLargestFiles(dir string, n int) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	var files []FileInfo
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		fullpath := filepath.Join(dir, entry.Name())
		stat, err := os.Stat(fullpath)
		if err != nil {
			return err
		}

		files = append(files, FileInfo{entry.Name(), stat.Size()})
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].Size > files[j].Size
	})

	limit := min(len(files), n)

	fmt.Printf("Top %d largest files in %s:\n", limit, dir)
	for i := 0; i < limit; i++ {
		fmt.Printf("%d. %s (%s)\n", i+1, files[i].Name, humanReadable(files[i].Size))
	}
	return nil
}

func humanReadable(n int64) string {
	const unit int64 = 1024
	if n < unit {
		return fmt.Sprintf("%d B", n)
	}

	div, exp := unit, 0
	for i := n / unit; i >= unit; i /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.1f %cB", float64(n)/float64(div), "KMGTPE"[exp])

}

func moveFile(folder, subfolder, filename string) error {
	err := os.Rename(filepath.Join(folder, filename), filepath.Join(folder, subfolder, filename))
	return err
}

func categorize(configData ConfigData, filename string) string {
	var hasStar bool
	var fallback string
	for foldername, keywords := range configData {
		for _, keyword := range keywords {
			if keyword == "*" {
				hasStar = true
				fallback = foldername
			}
			if strings.Contains(filename, keyword) {
				return foldername
			}
		}
	}

	if hasStar {
		return fallback
	}
	return ""
}
