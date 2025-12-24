package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/electr1fy0/sorta/internal"
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
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		globalConfig := filepath.Join(home, ".sorta", "config")
		if _, err := os.Stat(globalConfig); os.IsNotExist(err) {
			if err := os.MkdirAll(filepath.Dir(globalConfig), 0755); err != nil {
				return err
			}
			if err := internal.CreateConfig(globalConfig); err != nil {
				return err
			}
		}
		return nil
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
	rootCmd.PersistentFlags().StringVar(&configPath, "config", "~/.sorta/config", "Path to config file")
	rootCmd.PersistentFlags().IntVar(&internal.RecurseLevel, "recurselevel", -1, "Level of recursion to perform in the directory")
}
