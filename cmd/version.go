package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

var (
	Version = "0.6.3"
	Date    = "25-dec"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of sorta",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("sorta version %s\n", Version)
		fmt.Printf("built on: %s\n", Date)
		fmt.Printf("go version: %s\n", runtime.Version())
		fmt.Printf("os/arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
