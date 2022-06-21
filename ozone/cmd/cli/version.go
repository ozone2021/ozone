package cli

import (
	"github.com/spf13/cobra"
	"log"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:  "version",
	Long: `Status of all services in current directory`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Println("1.2-pre")
	},
}
