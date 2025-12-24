package cmd

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:     "init <directory>",
	Short:   "Initialize directory with the default config and prompt",
	Aliases: []string{"setup", "create"},
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, err := validateDir(args[0])
		if err != nil {
			return err
		}
		localPath := filepath.Join(dir, ".sorta")
		err = os.Mkdir(localPath, 0755)
		home, err := os.UserHomeDir()
		defaultPath := filepath.Join(home, ".sorta")

		configData, err := os.ReadFile(filepath.Join(defaultPath, "config"))
		promptData, err := os.ReadFile(filepath.Join(defaultPath, "prompt"))
		err = os.WriteFile(filepath.Join(localPath, "config"), configData, 0644)
		err = os.WriteFile(filepath.Join(localPath, "prompt"), promptData, 0644)
		return err
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
