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
	Use:   `add <foldername> "<keyword1>, <keyword2>..."`,
	Short: "Add new folder-to-keyword rule to the config file",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		foldername := args[0]
		keywordsStr := args[1]
		keywords := strings.Split(keywordsStr, ",")
		fmt.Println(args, len(args))
		return manageConfig(foldername, "add", keywords)
	},
}

var configRemoveCmd = &cobra.Command{
	Use:   "remove <foldername>",
	Short: "Remove a folder-to-keyword rule from the config file",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		keywords := args[0:]
		return manageConfig("", "remove", keywords)
	},
}

func manageConfig(foldername, operation string, keywords []string) error {
	if strings.HasPrefix(configPath, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("cannot determine home directory: %w", err)
		}
		configPath = filepath.Join(home, configPath[1:])
	}

	switch operation {
	case "add":
		f, err := os.OpenFile(configPath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
		if err != nil {
			return fmt.Errorf("error opening config file: %w", err)
		}
		defer f.Close()
		keyLine := strings.Join(keywords, ", ")
		line := fmt.Sprintf("%s = %s\n", foldername, keyLine)
		if _, err := f.WriteString(line); err != nil {
			return fmt.Errorf("error writing to config file: %w", err)
		}
		fmt.Printf("Added rule: %s=%s\n", foldername, keyLine)
		return nil
	case "remove":
		data, err := os.ReadFile(configPath)
		if err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("config file not found, nothing to remove")
			}
			return fmt.Errorf("error reading config file: %w", err)
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
			return fmt.Errorf("error writing updated config file: %w", err)
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
