package cmd

import (
	"io"
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

		defaultConf, err := os.Open(filepath.Join(defaultPath, "config"))
		if err != nil {
			return err
		}
		defaultPrompt, err := os.Open(filepath.Join(defaultPath, "prompt"))
		if err != nil {
			return err
		}
		localConf, err := os.Open(filepath.Join(localPath, "config"))
		if err != nil {
			return err
		}
		localPrompt, err := os.Open(filepath.Join(localPath, "prompt"))
		if err != nil {
			return err
		}

		_, err = io.Copy(localConf, defaultConf)
		if err != nil {
			return err
		}
		_, err = io.Copy(localPrompt, defaultPrompt)
		if err != nil {
			return err
		}
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
