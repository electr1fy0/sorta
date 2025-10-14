package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	dryRun      bool
	interactive bool
)

var rootCmd = &cobra.Command{
	Use:   "sorta",
	Short: "CLI to sort files based on extension and keywords",
	Long:  "A file organization tool that can sort by extension, config rules, or find duplicates.",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&dryRun, "dry", false, "Do a dry run without making changes")
	rootCmd.PersistentFlags().BoolVar(&interactive, "interactive", false, "Interactive mode")
}
