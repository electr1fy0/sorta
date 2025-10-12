package cmd

import (
	"github.com/electr1fy0/sorta/internal"
	"github.com/spf13/cobra"
)

var lrgCmd = &cobra.Command{
	Short: "List top 5 largest files",
	Use:   "lrg <directory>",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, err := validateDir(args[0])
		if err != nil {
			return err
		}
		return internal.TopLargestFiles(dir, 5)
	},
}

func init() {
	rootCmd.AddCommand(lrgCmd)
}
