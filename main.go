package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/electr1fy0/sorta/internal"
	"github.com/spf13/cobra"
)

const (
	ansiReset  = "\033[0m"
	ansiRed    = "\033[31m"
	ansiGreen  = "\033[32m"
	ansiYellow = "\033[33m"
	ansiCyan   = "\033[36m"
)

var (
	dryRun bool
	sorter internal.Sorter
)

var rootCmd = &cobra.Command{
	Use:   "sorta",
	Short: "CLI to sort files based on extension and keywords",
	Long:  "A file organization tool that can sort by extension, config rules, or find duplicates.",
}

var extCmd = &cobra.Command{
	Use:   "ext <directory>",
	Short: "Sort files based on common file extensions",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, err := validateDir(args[0])
		if err != nil {
			return err
		}

		sorter = internal.NewExtensionSorter()
		return runSort(dir)
	},
}

var configCmd = &cobra.Command{
	Use:   "config <directory>",
	Short: "Sort files based on .sorta-config",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, err := validateDir(args[0])
		if err != nil {
			return err
		}

		s, err := internal.NewConfigSorter()
		if err != nil {
			return fmt.Errorf("failed to create config sorter: %w", err)
		}
		sorter = s

		return runSort(dir)
	},
}

var duplCmd = &cobra.Command{
	Use:   "dupl <directory>",
	Short: "Filter out duplicate files",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, err := validateDir(args[0])
		if err != nil {
			return err
		}

		sorter = internal.NewDuplicateFinder()
		return runSort(dir)
	},
}

func validateDir(path string) (string, error) {
	// Handle absolute paths
	if filepath.IsAbs(path) {
		path = filepath.Clean(path)
	} else {
		// Resolve relative to home directory
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("cannot determine home directory: %w", err)
		}
		path = filepath.Join(home, path)
	}

	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("directory does not exist: %s", path)
		}
		return "", fmt.Errorf("cannot access path: %w", err)
	}

	if !info.IsDir() {
		return "", fmt.Errorf("not a directory: %s", path)
	}

	return path, nil
}

func runSort(dir string) error {
	fmt.Printf("%sDir:%s %s\n", ansiCyan, ansiReset, dir)

	executor := &internal.Executor{DryRun: dryRun}
	reporter := &internal.Reporter{DryRun: dryRun}

	res, err := internal.FilterFiles(dir, sorter, executor, reporter)
	if err != nil {
		return fmt.Errorf("failed to filter files: %w", err)
	}

	if err := internal.TopLargestFiles(dir, 5); err != nil {
		fmt.Printf("%sWarning:%s could not find largest files: %v\n", ansiYellow, ansiReset, err)
	}

	res.Print()
	return nil
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&dryRun, "dry", false, "Do a dry run without making changes")
	rootCmd.AddCommand(extCmd, configCmd, duplCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%sError:%s %v\n", ansiRed, ansiReset, err)
		os.Exit(1)
	}
}
