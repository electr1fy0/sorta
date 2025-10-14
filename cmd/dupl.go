package cmd

import (
	"github.com/electr1fy0/sorta/internal"
	"github.com/spf13/cobra"
)

var duplCmd = &cobra.Command{
	Use:   "dupl <directory>",
	Short: "Filter out duplicate files",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, err := validateDir(args[0])
		if err != nil {
			return err
		}

		return runSort(dir, internal.NewDuplicateFinder())
	},
}

func init() {
	duplCmd.PersistentFlags().BoolVar(&internal.DuplNuke, "nuke", false, "Delete duplicates permanently")
	rootCmd.AddCommand(duplCmd)
}
