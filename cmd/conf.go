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
	Use:   "add <keyword> <foldername>",
	Short: "Add new keyword-to-folder rule to .sorta-config",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		return manageConfig(args[0], args[1], "add")
	},
}

var configRemoveCmd = &cobra.Command{
	Use:   "remove <keyword>",
	Short: "Remove a keyword-to-folder rule from .sorta-config",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return manageConfig(args[0], "", "remove")
	},
}

func manageConfig(keyword, foldername, operation string) error {
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

		line := fmt.Sprintf("%s=%s\n", keyword, foldername)
		if _, err := f.WriteString(line); err != nil {
			return fmt.Errorf("error writing to .sorta-config: %w", err)
		}
		fmt.Printf("Added rule: %s=%s\n", keyword, foldername)
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
			if strings.HasPrefix(line, keyword+"=") {
				found = true
				continue
			}
			if line != "" {
				sb.WriteString(line + "\n")
			}
		}

		if !found {
			return fmt.Errorf("no rule found for keyword: %s", keyword)
		}

		if err := os.WriteFile(configPath, []byte(sb.String()), 0600); err != nil {
			return fmt.Errorf("error writing updated .sorta-config: %w", err)
		}
		fmt.Printf("Removed rule for keyword: %s\n", keyword)
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
