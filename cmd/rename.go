package cmd

import (
	"github.com/electr1fy0/sorta/internal"
	"github.com/spf13/cobra"
)

var renameCmd = &cobra.Command{
	Short: "Let Gemini rename your files",
	Use:   "rename <directory>",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, err := validateDir(args[0])
		if err != nil {
			return err
		}
		return runSort(dir, internal.NewRenamer())
	},
}

func init() {
	rootCmd.AddCommand(renameCmd)
}
