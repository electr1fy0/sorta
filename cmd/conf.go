package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"text/tabwriter"

	"github.com/electr1fy0/sorta/internal"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:     "config",
	Short:   "Manage sorta configuration",
	Aliases: []string{"conf", "cfg", "settings"},
}

var configEditCmd = &cobra.Command{
	Use:     "edit",
	Short:   "Open config file in default editor",
	Aliases: []string{"e", "open"},
	RunE: func(cmd *cobra.Command, args []string) error {
		path, err := resolvePath(configPath)
		if err != nil {
			return err
		}

		editor := os.Getenv("EDITOR")
		if editor == "" {
			editor = os.Getenv("VISUAL")
		}
		if editor == "" {
			if runtime.GOOS == "windows" {
				editor = "notepad"
			} else {
				editor = "vim"
			}
		}

		c := exec.Command(editor, path)
		c.Stdin = os.Stdin
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr

		return c.Run()
	},
}

var configInitCmd = &cobra.Command{
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

var configListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List all configuration rules",
	Aliases: []string{"ls", "show"},
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error
		configPath, err = resolvePath(configPath)
		if err != nil {
			return err
		}

		cfg, err := internal.ParseConfig(configPath)
		if err != nil {
			return err
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 4, ' ', 0)
		fmt.Fprintln(w, "FOLDER\tMATCHERS")
		fmt.Fprintln(w, "------\t--------")

		for i, folder := range cfg.Foldernames {
			var matchers []string
			for _, m := range cfg.Matchers[i] {
				if m.Regex != nil {
					matchers = append(matchers, fmt.Sprintf("regex(%s)", m.Regex.String()))
				} else {
					matchers = append(matchers, m.Raw)
				}
			}
			fmt.Fprintf(w, "%s\t%s\n", folder, strings.Join(matchers, ", "))
		}

		if len(cfg.Blacklist) > 0 {
			fmt.Fprintln(w, "\nBLACKLISTED FOLDERS")
			fmt.Fprintln(w, "-------------------")
			for _, b := range cfg.Blacklist {
				fmt.Fprintln(w, b)
			}
		}

		return w.Flush()
	},
}

var configAddCmd = &cobra.Command{
	Use:     `add "<foldername> = <keyword1>, <keyword2>..."`,
	Short:   "Add new folder-to-keyword rule to the config file",
	Aliases: []string{"new", "a"},
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		parts := strings.Split(args[0], "=")

		foldername := strings.TrimSpace(parts[0])
		keywords := strings.Split(parts[1], ",")

		return manageConfig(foldername, "add", keywords)
	},
}

var configRemoveCmd = &cobra.Command{
	Use:     "remove <foldername>",
	Short:   "Remove a folder-to-keyword rule from the config file",
	Aliases: []string{"rm", "del", "delete"},
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		keywords := args[0:]
		return manageConfig("", "remove", keywords)
	},
}

func manageConfig(foldername, operation string, keywords []string) error {
	var err error
	configPath, err = resolvePath(configPath)
	if err != nil {
		return err
	}

	switch operation {
	case "add":
		f, err := os.OpenFile(configPath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
		if err != nil {
			return fmt.Errorf("error opening config file: %w", err)
		}
		defer f.Close()
		keyLine := strings.Join(keywords, ", ")
		line := fmt.Sprintf("%s = %s\n", foldername, keyLine)
		if _, err := f.WriteString(line); err != nil {
			return fmt.Errorf("error writing to config file: %w", err)
		}
		fmt.Printf("Added rule: %s=%s to %s\n", foldername, keyLine, configPath)
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
			if strings.HasPrefix(line, keywords[0]+" =") || strings.HasPrefix(line, "!"+keywords[0]) {
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
		fmt.Printf("Removed rule for foldername: %s from %s\n", keywords[0], configPath)
		return nil
	default:
		return fmt.Errorf("unknown operation: %s", operation)
	}
}

var configPathCmd = &cobra.Command{
	Use:     "path",
	Short:   "Show the path of the configuration being used globally",
	Aliases: []string{"p", "location"},
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error
		configPath, err = resolvePath(configPath)
		if err != nil {
			return err
		}
		fmt.Println(configPath)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(configCmd)

	configCmd.AddCommand(configAddCmd)
	configCmd.AddCommand(configRemoveCmd)
	configCmd.AddCommand(configListCmd)
	configCmd.AddCommand(configPathCmd)
	configCmd.AddCommand(configEditCmd)
	configCmd.AddCommand(configInitCmd)
}
