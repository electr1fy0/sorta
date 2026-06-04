package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/creativeprojects/go-selfupdate"
	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update sorta to the latest version",
	RunE: func(cmd *cobra.Command, args []string) error {
		repo := selfupdate.ParseSlug("electr1fy0/sorta")
		latest, err := selfupdate.UpdateSelf(context.Background(), Version, repo)

		if err != nil {
			return fmt.Errorf("update failed: %w", err)
		}

		if latest.Version() == Version {
			fmt.Printf("sorta is already up to date (version %s)\n", Version)
		} else {
			fmt.Printf("Successfully updated sorta to version %s\n", latest.Version())
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
}

func CheckForUpdates() {
	if !isatty.IsTerminal(os.Stdout.Fd()) && !isatty.IsCygwinTerminal(os.Stdout.Fd()) {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	repo := selfupdate.ParseSlug("electr1fy0/sorta")
	latest, found, err := selfupdate.DetectLatest(ctx, repo)
	if err != nil || !found {
		return
	}

	if latest.GreaterThan(Version) {
		fmt.Printf("\n--- A new version of sorta is available: %s ---\n", latest.Version())
		fmt.Printf("Run 'sorta update' to install it.\n\n")
	}
}
