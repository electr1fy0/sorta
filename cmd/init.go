package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:     "init <directory>",
	Short:   "Initialize directory with the default config and prompt",
	Aliases: []string{"setup", "create", "initialize"},
	Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, err := validateDir(args[0])
			if err != nil {
				return err
			}
			localPath := filepath.Join(dir, ".sorta")
			if err := os.Mkdir(localPath, 0755); err != nil {
				if os.IsExist(err) {
					return fmt.Errorf("directory already initialized: %s", localPath)
				}
				return err
			}
			home, err := os.UserHomeDir()
			defaultPath := filepath.Join(home, ".sorta")
	
			configData, err := os.ReadFile(filepath.Join(defaultPath, "config"))
			if err != nil {
				return fmt.Errorf("failed to read default config: %w", err)
			}
			promptData, err := os.ReadFile(filepath.Join(defaultPath, "prompt"))
			if err != nil {
				return fmt.Errorf("failed to read default prompt: %w", err)
			}
	
			if err := os.WriteFile(filepath.Join(localPath, "config"), configData, 0644); err != nil {
				return err
			}
			if err := os.WriteFile(filepath.Join(localPath, "prompt"), promptData, 0644); err != nil {
				return err
			}
	
			fmt.Printf("Initialized sorta in: %s\n", localPath)
			return nil
		},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
