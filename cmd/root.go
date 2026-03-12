package cmd

import (
	"fmt"
	"os"

	"github.com/electr1fy0/sorta/internal/ops"
	"github.com/spf13/cobra"
)

var (
	dryRun     bool
	configPath string
)

var rootCmd = &cobra.Command{
	Use:   "sorta",
	Short: "CLI to sort files based on keywords and extensions",
	Long:  "A file organization tool that can sort by extension, config rules, or find duplicates.",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "Do a dry run without making changes")
	rootCmd.PersistentFlags().StringVar(&configPath, "config-path", "", "Path to config file (default: ~/.sorta/config)")
	rootCmd.PersistentFlags().IntVar(&ops.RecurseLevel, "recurse-level", 1<<10, "Level of recursion to perform in the directory")
}
