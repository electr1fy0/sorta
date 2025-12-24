package cmd

import (
	"fmt"

	"github.com/electr1fy0/sorta/internal"
	"github.com/spf13/cobra"
)

var undoCmd = &cobra.Command{
	Use:     "undo <directory>",
	Short:   "Undo the last operation on a directory",
	Aliases: []string{"u", "revert"},
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dir := args[0]
		fmt.Println(dir)
		return internal.Undo(dir)
	},
}

func init() {
	rootCmd.AddCommand(undoCmd)
}
