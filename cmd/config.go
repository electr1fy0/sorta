package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/electr1fy0/sorta/internal/config"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type yamlConfigDump struct {
	Rules  []yamlRuleDump `yaml:"rules"`
	Ignore []string       `yaml:"ignore,omitempty"`
}

type yamlRuleDump struct {
	Folder    string   `yaml:"folder"`
	Keywords  []string `yaml:"keywords,omitempty"`
	Regex     []string `yaml:"regex,omitempty"`
	Priority  int      `yaml:"priority,omitempty"`
	Match     string   `yaml:"match,omitempty"`
	CatchAll  bool     `yaml:"catch_all,omitempty"`
}

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
		if configPath != "" {
			var err error
			configPath, err = resolvePath(configPath)
			if err != nil {
				return err
			}
		}

		path, err := config.ResolveConfigPath(configPath, ".")
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

		globalPath, err := config.ResolveConfigPath("", "")
		if err != nil {
			return err
		}

		configData, err := os.ReadFile(globalPath)
		if err != nil {
			return fmt.Errorf("failed to read global config: %w", err)
		}

		if err := os.WriteFile(filepath.Join(localPath, "config"), configData, 0644); err != nil {
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
		if configPath != "" {
			var err error
			configPath, err = resolvePath(configPath)
			if err != nil {
				return err
			}
		}

		cfg, _, err := config.LoadConfig(configPath, ".")
		if err != nil {
			return err
		}

		if len(cfg.Rules) > 0 {
			fmt.Println("RULES:")
			for i, rule := range cfg.Rules {
				parts := make([]string, 0, len(rule.Matchers))
				for _, m := range rule.Matchers {
					s := matcherString(m)
					parts = append(parts, s)
				}
				flags := ""
				if rule.MatchAll {
					flags += " [match: all]"
				}
				if rule.Priority != 0 {
					flags += fmt.Sprintf(" [priority: %d]", rule.Priority)
				}
				fmt.Printf("  %d. %s = %s%s\n", i+1, rule.Folder, strings.Join(parts, ", "), flags)
			}
		}

		if len(cfg.Blacklist) > 0 {
			fmt.Println("\nIGNORE PATTERNS:")
			for _, b := range cfg.Blacklist {
				fmt.Printf("  !%s\n", b)
			}
		}

		if len(cfg.Warnings) > 0 {
			fmt.Println("\nWARNINGS:")
			for _, w := range cfg.Warnings {
				fmt.Printf("  %s\n", w)
			}
		}

		return nil
	},
}

func matcherString(m config.TypedMatcher) string {
	prefix := ""
	if m.Negate {
		prefix = "!"
	}
	switch m.Type {
	case config.MatcherRegex:
		return fmt.Sprintf("%sregex(%s)", prefix, m.Raw)
	case config.MatcherCatchAll:
		return "*"
	case config.MatcherExtension:
		return prefix + m.Raw
	default:
		return prefix + m.Raw
	}
}

var configAddCmd = &cobra.Command{
	Use:     `add "<foldername> = <keyword1>, <keyword2>..."`,
	Short:   "Add new rule to the config file",
	Aliases: []string{"new", "a"},
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		parts := strings.Split(args[0], "=")
		if len(parts) != 2 {
			return fmt.Errorf("usage: add \"<foldername> = <keyword1>, <keyword2>...\"")
		}
		foldername := strings.TrimSpace(parts[0])
		rawKeywords := strings.Split(parts[1], ",")
		keywords := make([]string, 0, len(rawKeywords))
		for _, k := range rawKeywords {
			k = strings.TrimSpace(k)
			if k != "" {
				keywords = append(keywords, k)
			}
		}

		return manageConfigAdd(foldername, keywords)
	},
}

var configRemoveCmd = &cobra.Command{
	Use:     "remove <foldername>",
	Short:   "Remove a rule by folder name from the config file",
	Aliases: []string{"rm", "del", "delete"},
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return manageConfigRemove(args[0])
	},
}

func getConfigFilePath() (string, error) {
	if configPath != "" {
		var err error
		configPath, err = resolvePath(configPath)
		if err != nil {
			return "", err
		}
	}
	return config.ResolveConfigPath(configPath, ".")
}

func manageConfigAdd(foldername string, keywords []string) error {
	path, err := getConfigFilePath()
	if err != nil {
		return err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("error reading config file: %w", err)
	}

	var ys yamlConfigDump
	if err := yaml.Unmarshal(data, &ys); err != nil {
		return fmt.Errorf("error parsing config: %w", err)
	}

	ys.Rules = append(ys.Rules, yamlRuleDump{
		Folder:   foldername,
		Keywords: keywords,
	})

	out, err := yaml.Marshal(&ys)
	if err != nil {
		return fmt.Errorf("error marshaling config: %w", err)
	}

	if err := os.WriteFile(path, out, 0644); err != nil {
		return fmt.Errorf("error writing config file: %w", err)
	}

	fmt.Printf("Added rule %q with keywords %v to %s\n", foldername, keywords, path)
	return nil
}

func manageConfigRemove(foldername string) error {
	path, err := getConfigFilePath()
	if err != nil {
		return err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("error reading config file: %w", err)
	}

	var ys yamlConfigDump
	if err := yaml.Unmarshal(data, &ys); err != nil {
		return fmt.Errorf("error parsing config: %w", err)
	}

	filtered := make([]yamlRuleDump, 0, len(ys.Rules))
	found := false
	for _, r := range ys.Rules {
		if r.Folder == foldername {
			found = true
			continue
		}
		filtered = append(filtered, r)
	}

	if !found {
		return fmt.Errorf("no rule found for folder: %s", foldername)
	}

	ys.Rules = filtered
	out, err := yaml.Marshal(&ys)
	if err != nil {
		return fmt.Errorf("error marshaling config: %w", err)
	}

	if err := os.WriteFile(path, out, 0644); err != nil {
		return fmt.Errorf("error writing config file: %w", err)
	}

	fmt.Printf("Removed rule for folder: %s from %s\n", foldername, path)
	return nil
}

var configPathCmd = &cobra.Command{
	Use:     "path",
	Short:   "Show the path of the configuration being used globally",
	Aliases: []string{"p", "location"},
	RunE: func(cmd *cobra.Command, args []string) error {
		if configPath != "" {
			var err error
			configPath, err = resolvePath(configPath)
			if err != nil {
				return err
			}
		}

		path, err := config.ResolveConfigPath(configPath, ".")
		if err != nil {
			return err
		}
		fmt.Println(path)
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
