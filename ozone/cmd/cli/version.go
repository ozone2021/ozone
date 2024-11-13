package cli

import (
	"fmt"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:  "version",
	Long: `Status of all services in current directory`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("1.9")
	},
}
