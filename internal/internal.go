package internal

import (
	"crypto/sha256"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
)

func (e *Executor) Execute(op FileOperation) error {
	if e.DryRun {
		return nil
	}

	switch op.Type {
	case OpMove:
		destDir := filepath.Dir(op.DestPath)

		// srcData, err := os.ReadFile(op.SourcePath)
		// _, err = os.Stat(op.DestPath)
		// if os.IsNotExist(err) { // file exists, yes = not not
		// 	destData, _ := os.ReadFile(op.DestPath)
		// 	checksum := sha256.Sum256(destData)
		// 	checkStringDest := fmt.Sprintf("%x", checksum)

		// 	checksumSrc := sha256.Sum256(srcData)
		// 	checkstringSrc := fmt.Sprintf("%x", checksumSrc)
		// 	fmt.Printf("src: %s dest: %s\n", checkstringSrc, checkStringDest)
		// 	if checkStringDest == checkstringSrc {
		// 		return nil
		// 	}
		// }
		if op.DestPath == op.SourcePath {
			return nil
		}
		// fmt.Printf("Moving file %s → %s\n", op.Filename, destDir)
		os.MkdirAll(destDir, 0755)
		err := os.Rename(op.SourcePath, op.DestPath)

		result.Moved++
		fmt.Println("moved: ", op.Filename)
		if err != nil {
			log.Fatalln("error executing: ", err)
		}

	case OpDelete:
		os.Remove(op.SourcePath)
	}

	return nil
}

func (r *Reporter) Report(op FileOperation, err error) {
	prefix := ansiGreen + "[OK]" + ansiReset
	if r.DryRun {
		prefix = ansiYellow + "[DRY]" + ansiReset
	}

	if err != nil {
		prefix = ansiRed + "[ERR]" + ansiReset
	}

	if op.Type == OpMove {
		fmt.Printf("%s %s -> %s (%s)\n", prefix, op.Filename, op.DestPath, humanReadable(op.Size))
	}
}

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

func (s *ExtensionSorter) Sort(baseDir, dir, filename string, size int64) (FileOperation, error) {
	ext := strings.ToLower(filepath.Ext(filename))

	for folder, extensions := range s.categories {
		if slices.Contains(extensions, ext) {
			return FileOperation{
				Type:       OpMove,
				SourcePath: filepath.Join(dir, filename),
				DestPath:   filepath.Join(baseDir, folder, filename),
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

func (s *ConfigSorter) Sort(baseDir, dir, filename string, size int64) (FileOperation, error) {
	folder := categorize(s.configData, filename)

	if folder == "" {
		return FileOperation{Type: OpSkip}, nil
	}
	// fmt.Println("moving ", dir, filename, folder)
	return FileOperation{
		Type:       OpMove,
		SourcePath: filepath.Join(dir, filename),
		DestPath:   filepath.Join(baseDir, folder, filename),
		Filename:   filename,
		Size:       size,
	}, nil
}

func NewDuplicateFinder() *DuplicateFinder {
	return &DuplicateFinder{
		hashes: make(map[string]string),
	}
}

func (d *DuplicateFinder) Sort(baseDir, dir, filename string, size int64) (FileOperation, error) {
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

var entries []FileInfo

func FilterFiles(dir string, sorter Sorter, executor *Executor, reporter *Reporter) (SortResult, error) {
	// hashes := make(map[string]string)
	entries = make([]FileInfo, 0, 1000)

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
		// fmt.Println("filtering:", path)
		stat, err := d.Info()
		entries = append(entries, FileInfo{d.Name(), stat.Size()})

		if err != nil {
			return err
		}
		size := stat.Size()
		parentDir := filepath.Dir(path)
		fileOp, err := sorter.Sort(dir, parentDir, filename, size)
		err = executor.Execute(fileOp)
		if err != nil {
			fmt.Println("Error executing: ", err)
			return err
		}

		if fileOp.Type == OpSkip {
			result.Skipped++
		}
		return nil
	})

	return result, nil
}

func TopLargestFiles(dir string, n int) error {

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Size > entries[j].Size
	})

	limit := min(len(entries), n)
	if strings.HasPrefix(entries[0].Name, ".") {
		return nil
	}
	fmt.Printf("Top %d largest files in %s:\n", limit, dir)
	for i := range limit {
		fmt.Printf("%d. %s (%s)\n", i+1, entries[i].Name, humanReadable(entries[i].Size))
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
