package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var cliVersion = "dev"

// SetVersion sets the CLI version (called from main with ldflags value).
func SetVersion(v string) {
	cliVersion = v
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the CLI version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("adapto " + cliVersion)
	},
}
