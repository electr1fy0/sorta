package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

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

var state = &internal.State{}
var sorter internal.Sorter

var rootCmd = &cobra.Command{
	Use:   "sorta",
	Short: "CLI to sort files based on extension and keywords",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println(ansiYellow + "Warning: No directory provided. Use a subcommand like 'ext', 'config', or 'dupl'." + ansiReset)
			return
		}
		for _, val := range args {
			if val != "" {
				state.CliDir += val + " "
			}
		}
	},
}

var extCmd = &cobra.Command{
	Use:   "ext <dir>",
	Short: "Sort files based on common file extensions",
	Run: func(cmd *cobra.Command, args []string) {
		dir := GetDir(args)
		if dir == "" {
			fmt.Println(ansiRed + "Error: No directory specified for 'ext'." + ansiReset)
			os.Exit(1)
		}
		state.CliDir = dir
		sorter = internal.NewExtensionSorter()
	},
}

var configCmd = &cobra.Command{
	Use:   "config <dir>",
	Short: "Sort files based on .sorta-config",
	Run: func(cmd *cobra.Command, args []string) {
		dir := GetDir(args)
		if dir == "" {
			fmt.Println(ansiRed + "Error: No directory specified for 'config'." + ansiReset)
			os.Exit(1)
		}
		state.CliDir = dir
		s, err := internal.NewConfigSorter()
		if err != nil {
			log.Fatalln(ansiRed+"Error creating config sorter:"+ansiReset, err)
		}
		sorter = s
	},
}

var duplCmd = &cobra.Command{
	Use:   "dupl <dir>",
	Short: "Filter out duplicate files",
	Run: func(cmd *cobra.Command, args []string) {
		dir := GetDir(args)
		if dir == "" {
			fmt.Println(ansiRed + "Error: No directory specified for 'dupl'." + ansiReset)
			os.Exit(1)
		}
		state.CliDir = dir
		sorter = internal.NewDuplicateFinder()
	},
}

func GetDir(args []string) string {
	if len(args) == 0 {
		return ""
	}

	var dirBuilder strings.Builder
	for _, val := range args {
		val = strings.TrimSpace(val)
		if val == "" {
			continue
		}
		dirBuilder.WriteString(val)
	}

	dir := dirBuilder.String()
	if dir == "" {
		return ""
	}

	if filepath.IsAbs(dir) {
		return filepath.Clean(dir)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatalln(ansiRed+"Error reading home directory:"+ansiReset, err)
	}

	full := filepath.Join(home, dir)
	info, err := os.Stat(full)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println(ansiRed+"Error:"+ansiReset, "Directory does not exist:", full)
		} else {
			fmt.Println(ansiRed+"Error:"+ansiReset, err)
		}
		os.Exit(1)
	}

	if !info.IsDir() {
		fmt.Println(ansiRed+"Error:"+ansiReset, full, "is not a directory.")
		os.Exit(1)
	}

	return full
}

func init() {
	rootCmd.Flags().BoolVar(&state.DryRun, "dry", false, "Do a dry run")
	rootCmd.AddCommand(extCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(duplCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(ansiRed+"Command error:"+ansiReset, err)
		os.Exit(1)
	}

	path := strings.TrimSpace(state.CliDir)
	if path == "" {
		fmt.Println(ansiRed + "Error: directory not provided." + ansiReset)
		os.Exit(1)
	}

	info, err := os.Stat(path)
	if err != nil {
		fmt.Println(ansiRed+"Error accessing directory:"+ansiReset, err)
		os.Exit(1)
	}
	if !info.IsDir() {
		fmt.Println(ansiRed+"Error:"+ansiReset, path, "is not a directory.")
		os.Exit(1)
	}

	if sorter == nil {
		fmt.Println(ansiRed + "Error: No sorting mode initialized. Use a subcommand like 'ext', 'config', or 'dupl'." + ansiReset)
		os.Exit(1)
	}

	fmt.Println(ansiCyan+"Dir:"+ansiReset, path)

	executor := &internal.Executor{DryRun: state.DryRun}
	reporter := &internal.Reporter{DryRun: state.DryRun}

	res, err := internal.FilterFiles(path, sorter, executor, reporter)
	if err != nil {
		fmt.Println(ansiRed+"Error filtering files:"+ansiReset, err)
		os.Exit(1)
	}

	if err := internal.TopLargestFiles(path, 5); err != nil {
		fmt.Println(ansiYellow+"Warning finding largest files:"+ansiReset, err)
	}

	res.Print()
}
