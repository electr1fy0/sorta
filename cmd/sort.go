package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/electr1fy0/sorta/internal"
	"github.com/spf13/cobra"
)

var sortCmd = &cobra.Command{
	Use:   "sort <directory>",
	Short: "Sort files based on keywords and extensions",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, err := validateDir(args[0])
		if err != nil {
			return err
		}

		if !cmd.Flags().Changed("config") {
			cwd, err := os.Getwd()
			if err == nil {
				localConfig := filepath.Join(cwd, ".sorta", "config")
				if _, err := os.Stat(localConfig); err == nil {
					configPath = localConfig
				}
			}
		}

		configPath, err = internal.ExpandPath(configPath)
		if err != nil {
			return err
		}

		configSorter, err := internal.NewConfigSorter(dir, configPath)
		if err != nil {
			return fmt.Errorf("error creating config sorter: %w", err)
		}
		return runSort(dir, configSorter, configSorter.GetBlacklist())
	},
}

func init() {
	rootCmd.AddCommand(sortCmd)
}
