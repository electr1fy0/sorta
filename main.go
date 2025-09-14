package main

import (
	"bufio"
	"crypto/sha256"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

type ConfigData map[string][]string

type State struct {
	cliDir              string
	dryRun              bool
	moveCnt, skippedCnt int
}

type FileInfo struct {
	Name string
	Size int64
}

// TODO:
// find top 5 largest files or sth
// move to trash folder instead of deleting
// a very interactive mode like
// [?] Move file "weirdfile.xyz" (12 MB)?
// [y] docs   [i] images   [m] movies   [s] skip
// make size human readable
// make logging tree like
// cache checksum
// a flag to just nuke duplicates
// docs:
// - file 1
// - file 2
// images:
// - img 1
// - img 2
// sort by size by calculating average size
// sort by data
// more expressive summary
// scan directories recursively for duplicates
// concurrency in system calls
// regex support in config
// blacklist / whitelist (like gitignore)
// interactive mode: ask users what to do with unmatched files
// something called MIME type. use that instead of ext
// Add * to add rest of the files to others * matches all
// Undo all sort and bring to root

var state = &State{}

var cmd = &cobra.Command{
	Short: "CLI to sort files based on extension and keywords",
	Use:   "sorta []",
	Run: func(cmd *cobra.Command, args []string) {
		for _, val := range args {
			state.cliDir += val + " "
		}
	},
}

func main() {
	cmd.Flags().BoolVar(&state.dryRun, "dry", false, "Do a dry run")

	if err := cmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	path, mode := getDirAndMode()
	fmt.Println("Dir:", path)

	err := filterFiles(path, mode)
	if err != nil {
		fmt.Println("Error filtering files: ", err)
		os.Exit(1)
	}
	if err := topLargestFiles(path, 5); err != nil {
		fmt.Println("Error finding largest files:", err)
	}
}

func createFolder(dir, foldername string) error {
	return os.MkdirAll(filepath.Join(dir, foldername), 0700)
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

func sortByExtension() {

}

func deduplicate() {

}

func sortByConfig() {

}

func reportResults() {

}

func handleMove(dir, foldername, filename string, size int64) error {
	if state.dryRun {
		fmt.Printf("Would move %s (%d bytes) to %s/%s\n", filename, size, dir, foldername)
	} else {
		createFolder(dir, foldername)
		err := moveFile(dir, foldername, filename)
		if err != nil {
			return err
		}
		fmt.Printf("Moved %s (%d bytes) to %s/%s\n", filename, size, dir, foldername)
	}
	return nil
}

func filterFiles(dir string, sortMode int) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		log.Fatalln("Error joining path: ", err)
	}

	hashes := make(map[string]string, len(entries))

	for _, entry := range entries {
		filename := entry.Name()
		isHidden := []rune(filename)[0] == '.'
		isDir := entry.IsDir()
		if isDir {
			filename += " (dir)"
		}

		if isHidden || isDir {
			continue
		}

		fullpath := filepath.Join(dir, filename)
		stat, err := os.Stat(fullpath)
		if err != nil {
			return err
		}
		size := stat.Size()

		switch sortMode {
		case 0:
			switch strings.ToLower(filepath.Ext(filename)) {
			case ".pdf", ".docx", ".pages", ".md", ".txt":
				err := handleMove(dir, "docs", filename, size)
				if err != nil {
					return err
				}
				state.moveCnt++

			case ".png", ".jpg", ".jpeg", ".heic", ".heif", ".webp":
				err := handleMove(dir, "images", filename, size)
				if err != nil {
					return err
				}
				state.moveCnt++

			case ".mp4", ".mov":
				err := handleMove(dir, "movies", filename, size)
				if err != nil {
					return err
				}
				state.moveCnt++
			default:
				state.skippedCnt++
			}

		case 1:
			configData, err := parseConfig()
			if err != nil {
				return err
			}
			foldername := categorize(configData, filename)
			if foldername != "" {
				err = handleMove(dir, foldername, filename, size)
				if err != nil {
					return err
				}
				state.moveCnt++
			} else {
				state.skippedCnt++
			}

		case 2:
			data, _ := os.ReadFile(fullpath)
			checksum256 := sha256.Sum256(data)
			digest := fmt.Sprintf("%x", checksum256)
			fmt.Printf("Checksum: %s (%s, %d bytes)\n", digest, filename, size)

			if _, exists := hashes[digest]; exists {
				if state.dryRun {
					fmt.Printf("Would move duplicate %s (%d bytes) to %s/duplicates\n", filename, size, dir)
				} else {
					os.MkdirAll(filepath.Join(dir, "duplicates"), 0700)
					err := os.Rename(fullpath, filepath.Join(dir, "duplicates", filename))
					if err != nil {
						return err
					}
					fmt.Printf("Moved duplicate %s (%d bytes) to %s/duplicates\n", filename, size, dir)
				}
				state.moveCnt++
			} else {
				hashes[digest] = fullpath
			}
		}
	}

	if sortMode != 2 {
		if state.moveCnt == 0 {
			if state.dryRun {
				fmt.Println("Nothing to do (dry run)")
			} else {
				fmt.Println("Already sorted")
			}
		} else {
			if state.dryRun {
				fmt.Printf("Dry run: %d files would be sorted, %d skipped.\n", state.moveCnt, state.skippedCnt)
			} else {
				fmt.Printf("%d files sorted, %d skipped.\n", state.moveCnt, state.skippedCnt)
			}
		}
	}
	return nil
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

func topLargestFiles(dir string, n int) error {
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

func parseConfig() (ConfigData, error) {
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

func getDirAndMode() (string, int) {
	var mode int
	reader := bufio.NewReader(os.Stdin)
	var dir string
	var err error

	if strings.TrimSpace(state.cliDir) != "" {
		dir = state.cliDir
	} else {
		fmt.Println("Enter the directory (relative to home dir, no quotes):")
		fmt.Print("~/")
		dir, err = reader.ReadString('\n')
	}
	home, err := os.UserHomeDir()
	dir = filepath.Join(home, dir)

	if err != nil {
		log.Fatalln("Error reading directory path", err)
	}
	dir = strings.TrimSpace(dir)
	fmt.Println("Choose mode index:")
	fmt.Println("0: Sort based on file extension")
	fmt.Println("1: Sort based on keywords in config")
	fmt.Println("2: Filter duplicates")

	fmt.Fscanln(reader, &mode)
	if err != nil {
		log.Fatalln("Error joining path", err)
	}
	return dir, mode
}
