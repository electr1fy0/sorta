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
	Use:   "add",
	Short: "Add new keyword-to-folder rule to .sorta-config",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 2 {
			fmt.Println("usage: sorta config add <keyword> <foldername>")
			return
		}

		homeDir, err := os.UserHomeDir()
		if err != nil {
			fmt.Println("error getting home directory:", err)
			return
		}

		configPath := filepath.Join(homeDir, ".sorta-config")
		f, err := os.OpenFile(configPath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
		if err != nil {
			fmt.Println("error opening .sorta-config:", err)
			return
		}
		defer func() {
			if cerr := f.Close(); cerr != nil {
				fmt.Println("error closing .sorta-config:", cerr)
			}
		}()

		keyword := args[0]
		foldername := args[1]
		line := fmt.Sprintf("%s=%s\n", keyword, foldername)

		if _, err := f.WriteString(line); err != nil {
			fmt.Println("error writing to .sorta-config:", err)
			return
		}

		fmt.Printf("Added rule: %s=%s\n", keyword, foldername)
	},
}

var configRemoveCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove a keyword-to-folder rule from .sorta-config",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			fmt.Println("usage: sorta config remove <keyword>")
			return
		}

		homeDir, err := os.UserHomeDir()
		if err != nil {
			fmt.Println("error getting home directory:", err)
			return
		}

		configPath := filepath.Join(homeDir, ".sorta-config")
		data, err := os.ReadFile(configPath)
		if err != nil {
			if os.IsNotExist(err) {
				fmt.Println(".sorta-config not found, nothing to remove")
			} else {
				fmt.Println("error reading .sorta-config:", err)
			}
			return
		}

		lines := strings.Split(string(data), "\n")
		var sb strings.Builder
		found := false
		for _, line := range lines {
			if strings.HasPrefix(line, args[0]+"=") {
				found = true
				continue
			}
			if line != "" {
				sb.WriteString(line + "\n")
			}
		}

		if !found {
			fmt.Printf("no rule found for keyword: %s\n", args[0])
			return
		}

		if err := os.WriteFile(configPath, []byte(sb.String()), 0600); err != nil {
			fmt.Println("error writing updated .sorta-config:", err)
			return
		}

		fmt.Printf("removed rule for keyword: %s\n", args[0])
	},
}

func init() {
	rootCmd.AddCommand(configCmd)

	configCmd.AddCommand(configAddCmd)
	configCmd.AddCommand(configRemoveCmd)
}
