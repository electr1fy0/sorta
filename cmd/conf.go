package cmd

import (
	"fmt"

	"github.com/electr1fy0/sorta/internal"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "conf <directory>",
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

		return runSort(dir, s)

	},
}

func init() {
	rootCmd.AddCommand(configCmd)
}
