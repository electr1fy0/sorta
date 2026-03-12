package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/electr1fy0/sorta/internal/config"
	"github.com/electr1fy0/sorta/internal/ignore"
	"github.com/spf13/cobra"
)

var checkIgnoreCmd = &cobra.Command{
	Use:   "check-ignore <path> | check-ignore <directory> <path>",
	Short: "Explain whether a path is ignored and by which rule",
	Args:  cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		rootDir := "."
		targetPath := args[0]

		if len(args) == 2 {
			rootDir = args[0]
			targetPath = args[1]
		}

		root, err := validateDir(rootDir)
		if err != nil {
			return err
		}
		if configPath != "" {
			configPath, err = resolvePath(configPath)
			if err != nil {
				return err
			}
		}

		cfg, _, err := config.LoadConfig(configPath, root)
		if err != nil {
			return err
		}

		matcher, err := ignore.LoadIgnoreMatcher(root, cfg.Blacklist)
		if err != nil {
			return err
		}

		pathToCheck := targetPath
		if !filepath.IsAbs(pathToCheck) {
			pathToCheck = filepath.Join(root, pathToCheck)
		}
		pathToCheck = filepath.Clean(pathToCheck)

		info, err := os.Stat(pathToCheck)
		isDir := false
		if err == nil {
			isDir = info.IsDir()
		}

		rule, matched := matcher.Explain(root, pathToCheck, isDir)
		if !matched {
			fmt.Printf("%s is not ignored\n", pathToCheck)
			return nil
		}

		fmt.Printf("%s is ignored\n", pathToCheck)
		fmt.Printf("pattern: %s\n", rule.Pattern)
		fmt.Printf("source: %s\n", rule.Source)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(checkIgnoreCmd)
}
