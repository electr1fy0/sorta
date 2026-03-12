package cmd

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/electr1fy0/sorta/internal/core"
	"github.com/electr1fy0/sorta/internal/ops"
	"github.com/spf13/cobra"
)

var historyCmd = &cobra.Command{
	Use:     "history",
	Short:   "View operation history",
	Aliases: []string{"log", "ls"},
	Long:    "Displays a list of past sort operations recorded by sorta.",
	RunE: func(cmd *cobra.Command, args []string) error {
		transactions, err := ops.GetHistory()
		if err != nil {
			return fmt.Errorf("failed to retrieve history: %w", err)
		}

		if len(transactions) == 0 {
			fmt.Println("No history found.")
			return nil
		}

		oneline, err := cmd.Flags().GetBool("oneline")
		if err != nil {
			return err
		}
		if oneline {
			for _, t := range transactions {
				typeStr := "action"
				if t.TType == core.TUndo {
					typeStr = "undo"
				}
				rootDir := ""
				if len(t.Operations) > 0 {
					rootDir = t.Operations[0].File.RootDir
				}
				id := strings.ReplaceAll(t.ID, " ", "_")
				fmt.Printf("%s %s %d %s\n", id, typeStr, len(t.Operations), rootDir)
			}
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tType\tFiles Affected\tRoot Directory")
		for _, t := range transactions {
			typeStr := "Action"
			if t.TType == core.TUndo {
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
	historyCmd.Flags().Bool("oneline", false, "Show compact history output")
	rootCmd.AddCommand(historyCmd)
}
