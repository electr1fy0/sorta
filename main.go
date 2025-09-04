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
// count skips in summary, provide overall more detailed summary
// sort by size: small, medium and large
// scan directories recursively for duplicates
// concurrency in system calls
// config:
// - handle multi word keywords
// regex support in config
// blacklist / whitelist (like gitignore)
// dry run flag to see changes before they actually happen
// interactive mode: ask users what to do with unmatched files
// something called MIME type. use that instead of ext

var cliDir string
var dryRun bool

var cmd = &cobra.Command{
	Short: "CLI to sort files based on extension and keywords",
	Use:   "sorta []",
	Run: func(cmd *cobra.Command, args []string) {
		for _, val := range args {
			cliDir += val + " "
		}
	},
}

func main() {
	cmd.Flags().BoolVar(&dryRun, "dry", false, "Do a dry run")

	if err := cmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
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

	moveCnt, skippedCnt := 0, 0
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
			if dryRun {
				println("Would create folder:", filepath.Join(path, "docs"))
				println("Would create folder:", filepath.Join(path, "images"))
				println("Would create folder:", filepath.Join(path, "movies"))
			} else {
				createFolder(path, "docs")
				createFolder(path, "images")
				createFolder(path, "movies")
			}

			switch strings.ToLower(filepath.Ext(filename)) {
			case ".pdf", ".docx", ".pages", ".md", ".txt":
				if dryRun {
					fmt.Printf("Would move %s to %s/docs\n", filename, path)
				} else {
					err := moveFile(path, "docs", filename)
					if err != nil {
						return err
					}
				}
				moveCnt++

			case ".png", ".jpg", ".jpeg", ".heic", ".heif", ".webp":
				if dryRun {
					fmt.Printf("Would move %s to %s/images\n", filename, path)
				} else {
					err := moveFile(path, "images", filename)
					if err != nil {
						return err
					}
				}
				moveCnt++

			case ".mp4", ".mov":
				if dryRun {
					fmt.Printf("Would move %s to %s/movies\n", filename, path)
				} else {
					err := moveFile(path, "movies", filename)
					if err != nil {
						return err
					}
				}
				moveCnt++

			default:
				skippedCnt++
			}

		case 1:
			configData, err := parseConfig()
			if err != nil {
				return err
			}
			foldername := categorize(configData, filename)
			if foldername != "" {
				if dryRun {
					fmt.Printf("Would create folder %s and move %s there\n", foldername, filename)
				} else {
					err := createFolder(path, foldername)
					if err != nil {
						return err
					}
					err = moveFile(path, foldername, filename)
					if err != nil {
						return err
					}
				}
				moveCnt++
			} else {
				skippedCnt++
			}

		case 2:
			fullpath := filepath.Join(path, filename)
			data, _ := os.ReadFile(fullpath)
			checksum256 := sha256.Sum256(data)
			digest := fmt.Sprintf("%x", checksum256)
			fmt.Println("Checksum:", digest)

			if _, exists := hashes[digest]; exists {
				if dryRun {
					fmt.Printf("Would move duplicate %s to %s/duplicates\n", filename, path)
				} else {
					os.MkdirAll(filepath.Join(path, "duplicates"), 0700)
					err := os.Rename(fullpath, filepath.Join(path, "duplicates", filename))
					if err != nil {
						return err
					}
				}
				moveCnt++
			} else {
				hashes[digest] = fullpath
			}
		}
	}
	if sortMode == 2 {
		return nil
	}
	if sortMode == 2 {
		return nil
	}

	if moveCnt == 0 {
		if dryRun {
			fmt.Println("Nothing to do (dry run)")
		} else {
			fmt.Println("Already sorted")
		}
	} else {
		if dryRun {
			fmt.Printf("Dry run: %d files would be sorted, %d skipped.\n", moveCnt, skippedCnt)
		} else {
			fmt.Printf("%d files sorted, %d skipped.\n", moveCnt, skippedCnt)
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
	var mode int
	var path string
	reader := bufio.NewReader(os.Stdin)
	var dir string
	var err error
	if strings.TrimSpace(cliDir) != "" {
		dir = cliDir
	} else {
		fmt.Println("Enter the directory (relative to home dir, no quotes):")
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
