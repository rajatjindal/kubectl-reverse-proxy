package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Version is set during build time
var Version = "unknown"

// actionCmd is the github action command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Display version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(Version)
	},
}
