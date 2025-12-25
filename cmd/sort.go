package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/electr1fy0/sorta/internal"
	"github.com/spf13/cobra"
)

func getDir(args []string) (string, error) {
	var dirLine string
	if len(args) < 1 {
		fmt.Printf("Enter directory for sorta to run on: ")
		reader := bufio.NewReader(os.Stdin)
		input, err := reader.ReadString('\n')
		if err != nil {
			return "", err
		}
		dirLine = input
	} else {
		dirLine = args[0]
	}
	return validateDir(strings.TrimSpace(dirLine))
}

var sortCmd = &cobra.Command{
	Use:     "sort <directory>",
	Short:   "Sort files based on keywords and extensions",
	Aliases: []string{"s", "organize"},
	Args:    cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, err := getDir(args)
		if err != nil {
			return err
		}

		if !cmd.Flags().Changed("config-path") {
			localConfig := filepath.Join(dir, ".sorta", "config")
			if _, err := os.Stat(localConfig); err == nil {
				configPath = localConfig
			}
		}

		configPath, err = resolvePath(configPath)
		if err != nil {
			return err
		}

		configSorter, err := internal.NewConfigSorter(dir, configPath, inline)
		if err != nil {
			return fmt.Errorf("error creating config sorter: %w", err)
		}

		return runSort(dir, configSorter, configSorter.GetBlacklist())
	},
}

var inline string

func init() {
	rootCmd.AddCommand(sortCmd)
	sortCmd.PersistentFlags().StringVar(&inline, "inline", "", "Skip the config and read a single line from the flag's value")
}
