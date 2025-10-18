package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage sorta configuration",
}

var configAddCmd = &cobra.Command{
	Use:   "add <foldername> <keyword1> <keyword2>...",
	Short: "Add new folder-to-keyword rule to .sorta-config",
	Args:  cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		foldername := args[0]
		keywords := args[1:]

		return manageConfig(foldername, "add", keywords)
	},
}

var configRemoveCmd = &cobra.Command{
	Use:   "remove <foldername>",
	Short: "Remove a folder-to-keyword rule from .sorta-config",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		keywords := args[0:]
		return manageConfig("", "remove", keywords)
	},
}

func manageConfig(foldername, operation string, keywords []string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("error getting home directory: %w", err)
	}
	configPath := filepath.Join(homeDir, ".sorta-config")

	switch operation {
	case "add":
		f, err := os.OpenFile(configPath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
		if err != nil {
			return fmt.Errorf("error opening .sorta-config: %w", err)
		}
		defer f.Close()
		keyLine := strings.Join(keywords, ", ")
		line := fmt.Sprintf("%s = %s\n", foldername, keyLine)
		if _, err := f.WriteString(line); err != nil {
			return fmt.Errorf("error writing to .sorta-config: %w", err)
		}
		fmt.Printf("Added rule: %s=%s\n", foldername, keyLine)
		return nil
	case "remove":
		data, err := os.ReadFile(configPath)
		if err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf(".sorta-config not found, nothing to remove")
			}
			return fmt.Errorf("error reading .sorta-config: %w", err)
		}

		lines := strings.Split(string(data), "\n")
		var sb strings.Builder
		found := false
		for _, line := range lines {
			if strings.HasPrefix(line, keywords[0]+" =") {
				found = true
				continue
			}
			if line != "" {
				sb.WriteString(line + "\n")
			}
		}

		if !found {
			return fmt.Errorf("no rule found for folder: %s", keywords[0])
		}

		if err := os.WriteFile(configPath, []byte(sb.String()), 0600); err != nil {
			return fmt.Errorf("error writing updated .sorta-config: %w", err)
		}
		fmt.Printf("Removed rule for foldername: %s\n", keywords[0])
		return nil
	default:
		return fmt.Errorf("unknown operation: %s", operation)
	}
}

func init() {
	rootCmd.AddCommand(configCmd)

	configCmd.AddCommand(configAddCmd)
	configCmd.AddCommand(configRemoveCmd)
}
