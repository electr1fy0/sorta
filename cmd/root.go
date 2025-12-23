package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/electr1fy0/sorta/internal"
	"github.com/spf13/cobra"
)

// flag logic (to be implemented)
// 8 bits
// 0th bit = dryRun
// 1st bit = interactive (disabled for now)
// 3rd bit = recurseLevel

var flags uint8 = 0

var (
	dryRun      bool
	interactive bool
	configPath  string
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

		if strings.HasPrefix(configPath, "~") {
			home, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("cannot determine home directory: %w", err)
			}
			configPath = filepath.Join(home, configPath[1:])
		}

		configSorter, err := internal.NewConfigSorter(dir, configPath)
		if err != nil {
			return fmt.Errorf("error creating config sorter: %w", err)
		}
		return runSort(dir, configSorter, configSorter.GetBlacklist())
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
	rootCmd.PersistentFlags().StringVar(&configPath, "config", "~/.sorta/config", "Path to config file")
	rootCmd.PersistentFlags().IntVar(&internal.RecurseLevel, "recurselevel", -1, "Level of recursion to perform in the directory")
	rootCmd.PersistentFlags().StringVar(&internal.Mode, "categorizeby", "contains", "Categorize by contains/regex")
}
