package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/electr1fy0/sorta/internal"
	"github.com/spf13/cobra"
)

var sortCmd = &cobra.Command{
	Use:     "sort <directory>",
	Short:   "Sort files based on keywords and extensions",
	Aliases: []string{"s", "organize"},
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, err := validateDir(args[0])
		if err != nil {
			return err
		}

		if !cmd.Flags().Changed("configPath") {
			localConfig := filepath.Join(dir, ".sorta", "config")
			if _, err := os.Stat(localConfig); err == nil {
				configPath = localConfig
			}
		}

		configPath, err = resolvePath(configPath)
		if err != nil {
			return err
		}

		configSorter, err := internal.NewConfigSorter(dir, configPath, inline)
		if err != nil {
			return fmt.Errorf("error creating config sorter: %w", err)
		}

		return runSort(dir, configSorter, configSorter.GetBlacklist())
	},
}

var inline string

func init() {
	rootCmd.AddCommand(sortCmd)
	sortCmd.PersistentFlags().StringVar(&inline, "inline", "", "Skip the config and read a single line from the flag's value")
}
