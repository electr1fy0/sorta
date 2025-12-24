package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/electr1fy0/sorta/internal"
	"github.com/spf13/cobra"
)

var historyCmd = &cobra.Command{
	Use:     "history",
	Short:   "View operation history",
	Aliases: []string{"log", "ls"},
	Long:    "Displays a list of past sort operations recorded by sorta.",
	RunE: func(cmd *cobra.Command, args []string) error {
		transactions, err := internal.GetHistory()
		if err != nil {
			return fmt.Errorf("failed to retrieve history: %w", err)
		}

		if len(transactions) == 0 {
			fmt.Println("No history found.")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tType\tFiles Affected\tRoot Directory")
		for _, t := range transactions {
			typeStr := "Action"
			if t.TType == internal.TUndo {
				typeStr = "Undo"
			}
			rootDir := ""
			if len(t.Operations) > 0 {
				rootDir = t.Operations[0].File.RootDir
			}
			fmt.Fprintf(w, "%s\t%s\t%d\t%s\n", t.ID, typeStr, len(t.Operations), rootDir)
		}
		w.Flush()

		return nil
	},
}

func init() {
	rootCmd.AddCommand(historyCmd)
}
