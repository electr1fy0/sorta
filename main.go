package main

import (
	"bufio"
	"crypto/sha256"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

type ConfigData map[string][]string

// TODO:
// scan directories recursively for duplicates
// concurrency in system calls
// config:
// - handle multi word keywords
// - add comment support with template explanation of config format
// sort by size: small, medium and large
// regex support in config
// blacklist / whitelist (like gitignore)
// cobra support
// dry run flag to see changes before they actually happen
// summary of moves / skips
// interactive mode: ask users what to do with unmatched files
// something called MIME type. use that instead of ext

var cliDir string

var cmd = cobra.Command{ //todo
	Short: "CLI to sort files based on extension and keywords",
	Use:   "sorta []",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		for _, val := range args {
			cliDir += val + " "
		}
	},
}

func getPath(path string) string {
	return path
}

func main() {
	cmd.Execute()
	path, mode := getPathAndMode()
	fmt.Println("Dir:", path)

	err := filterFiles(path, mode)
	if err != nil {
		fmt.Println("Error filtering files: ", err)
		os.Exit(1)
	}
}

func createFolder(path, foldername string) error {
	return os.MkdirAll(filepath.Join(path, foldername), 0700)
}

func moveFile(folder, subfolder, filename string) error {
	err := os.Rename(filepath.Join(folder, filename), filepath.Join(folder, subfolder, filename))
	return err
}

func categorize(configData ConfigData, filename string) string {
	for foldername, keywords := range configData {
		for _, keyword := range keywords {
			if strings.Contains(filename, keyword) {
				return foldername
			}
		}
	}
	return ""
}

func filterFiles(path string, sortMode int) error {
	entries, err := os.ReadDir(path)
	if err != nil {
		log.Fatalln("Error joining path: ", err)
	}

	moveCnt := 0
	var hashes map[string]string = make(map[string]string, len(entries))
	for _, entry := range entries {
		filename := entry.Name()

		var isHidden bool = []rune(filename)[0] == '.'
		var isDir = entry.IsDir()
		if isDir {
			filename += " (dir)"
		}

		if isHidden || isDir {
			continue
		}

		switch sortMode {
		case 0:
			createFolder(path, "docs")
			createFolder(path, "images")
			createFolder(path, "movies")
			switch strings.ToLower(filepath.Ext(filename)) {
			case ".pdf", ".docx", ".pages", ".md", ".txt":
				err := moveFile(path, "docs", filename)
				moveCnt++
				log.Println("moving1")
				if err != nil {
					return err
				}
			case ".png", ".jpg", ".jpeg", ".heic", ".heif", ".webp":
				err := moveFile(path, "images", filename)
				log.Println("moving1")
				moveCnt++
				if err != nil {
					return err
				}
			case ".mp4", ".mov":
				log.Println("moving1")
				moveCnt++
				err := moveFile(path, "movies", filename)
				if err != nil {
					return err
				}
			}
		case 1:
			configData, err := parseConfig()
			if err != nil {
				return err
			}
			foldername := categorize(configData, filename)
			if foldername != "" {
				err := createFolder(path, foldername)
				if err != nil {
					return err
				}
				moveCnt++
				err = moveFile(path, foldername, filename)
				if err != nil {
					return err
				}
			}

		case 2:
			fullpath := filepath.Join(path, filename)
			data, _ := os.ReadFile(fullpath)
			checksum256 := sha256.Sum256(data)
			digest := fmt.Sprintf("%x", checksum256)
			fmt.Println(digest)
			if _, exists := hashes[digest]; exists {
				os.MkdirAll(filepath.Join(path, "duplicates"), 0700)
				moveCnt++
				err := os.Rename(fullpath, filepath.Join(path, "duplicates", filename))
				if err != nil {
					return err
				}
			} else {
				hashes[digest] = fullpath
			}
		}
	}

	if moveCnt == 0 {
		fmt.Println("Already sorted")
	} else {
		fmt.Println(moveCnt, "files sorted successfully.")
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
//
// Example:
// invoice,bill,txt=Finance
// track,song=Music
// notes,book=Study`)

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

func getPathAndMode() (string, int) {
	fmt.Println("Enter the directory (relative to home dir, no quotes):")

	var mode int
	var path string
	reader := bufio.NewReader(os.Stdin)
	var dir string
	var err error
	if strings.TrimSpace(cliDir) != "" {
		dir = cliDir
	} else {
		fmt.Print("~/")
		dir, err = reader.ReadString('\n')
	}
	if err != nil {
		log.Fatalln("Error reading directory path", err)
	}
	dir = strings.TrimSpace(dir)
	fmt.Println("Choose mode index:")
	fmt.Println("0: Sort based on file extension")
	fmt.Println("1: Sort based on keywords in config")
	fmt.Println("2: Filter duplicates")

	fmt.Fscanln(reader, &mode)
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatalln("Error joining path", err)
	}
	path = filepath.Join(home, dir)
	return path, mode
}
