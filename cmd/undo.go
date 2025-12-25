package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/electr1fy0/sorta/internal"
	"github.com/spf13/cobra"
)

var undoCmd = &cobra.Command{
	Use:     "undo <directory>",
	Short:   "Undo the last operation on a directory",
	Aliases: []string{"u", "revert"},
	Args:    cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, err := getDir(args)
		if err != nil {
			return err
		}
		fmt.Printf("Are you sure you want to undo the last operation in %s? [y/N]: ", dir)
		reader := bufio.NewReader(os.Stdin)
		ans, _ := reader.ReadString('\n')
		if strings.ToLower(strings.TrimSpace(ans)) != "y" {
			fmt.Println("Undo cancelled.")
			return nil
		}

		if err := internal.Undo(dir); err != nil {
			if errors.Is(err, internal.ErrAlreadyUndone) {
				fmt.Printf("Last operation in %s already undone\n", dir)
				return nil
			}
			return err
		}

		fmt.Printf("Undid last operation in: %s\n", dir)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(undoCmd)
}
