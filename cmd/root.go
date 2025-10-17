package cmd

import (
	"fmt"
	"os"

	"github.com/electr1fy0/sorta/internal"
	"github.com/spf13/cobra"
)

var (
	dryRun      bool
	interactive bool
)

var rootCmd = &cobra.Command{
	Use:   "sorta",
	Short: "CLI to sort files based on keywords and extensions",
	Long:  "A file organization tool that can sort by extension, config rules, or find duplicates.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, err := validateDir(args[0])
		if err != nil {
			return err
		}
		configSorter, err := internal.NewConfigSorter()
		if err != nil {
			fmt.Println("error creating config sorter:", err)
			return err
		}
		return runSort(dir, configSorter)
	},
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
