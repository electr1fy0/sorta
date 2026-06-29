package cmd

import (
	"github.com/electr1fy0/sorta/internal/llm"
	"github.com/electr1fy0/sorta/internal/rename"
	"github.com/spf13/cobra"
)

var (
	renameModel     string
	renameNames     []string
	renameNamesFile string
	renameCase      string
)

var renameCmd = &cobra.Command{
	Short:   "Let the planner rename your files",
	Use:     "rename <directory>",
	Aliases: []string{"rn"},
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, err := validateDir(args[0])
		if err != nil {
			return err
		}

		caseType, err := rename.ParseCaseType(renameCase)
		if err != nil {
			return err
		}

		names, err := loadAgentNames(renameNames, renameNamesFile)
		if err != nil {
			return err
		}

		if len(names) == 0 && caseType != rename.CaseNone {
			return runSort(dir, rename.NewCaseRenamer(caseType), nil)
		}

		client, err := llm.NewClient(renameModel)
		if err != nil {
			return err
		}
		r := rename.NewRenamer(client, renameModel, names)
		r.SetCaseType(caseType)
		return runSort(dir, r, nil)
	},
}

func init() {
	renameCmd.Flags().StringVar(&renameModel, "model", llm.DefaultModel, "Rename model to use")
	renameCmd.Flags().StringSliceVar(&renameNames, "names", nil, "Extra labels or examples to guide rename")
	renameCmd.Flags().StringVar(&renameNamesFile, "names-file", "", "File containing one extra rename hint per line")
	renameCmd.Flags().StringVar(&renameCase, "case", "", "Case transformation: snake, kebab, camel, pascal, upper, lower, title")
	rootCmd.AddCommand(renameCmd)
}
