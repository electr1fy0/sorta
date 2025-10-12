package cmd

import (
	"github.com/electr1fy0/sorta/internal"
	"github.com/spf13/cobra"
)

var extCmd = &cobra.Command{
	Use:   "ext <directory>",
	Short: "Sort files based on common file extensions",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, err := validateDir(args[0])
		if err != nil {
			return err
		}

		return runSort(dir, internal.NewExtensionSorter())
	},
}

func init() {
	rootCmd.AddCommand(extCmd)
}
