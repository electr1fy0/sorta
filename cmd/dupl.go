package cmd

import (
	"github.com/electr1fy0/sorta/internal"
	"github.com/spf13/cobra"
)

var duplCmd = &cobra.Command{
	Use:     "duplicates <directory>",
	Short:   "Filter out duplicate files",
	Args:    cobra.MaximumNArgs(1),
	Aliases: []string{"dupl", "dedupe", "dd"},

	RunE: func(cmd *cobra.Command, args []string) error {
		dir, err := getDir(args)
		if err != nil {
			return err
		}

		return runSort(dir, internal.NewDuplicateFinder(), nil)
	},
}

func init() {
	duplCmd.PersistentFlags().BoolVar(&internal.DuplNuke, "nuke", false, "Delete duplicates permanently")
	rootCmd.AddCommand(duplCmd)
}
