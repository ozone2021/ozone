package cli

import (
	"github.com/spf13/cobra"
	process_manager_client "ozone-daemon-lib/process-manager-client"
)

func init() {
	rootCmd.AddCommand(haltCmd)
}

var haltCmd = &cobra.Command{
	Use:   "h",
	Long:  `Halts all services in current directory`,
	Run: func(cmd *cobra.Command, args []string) {
		process_manager_client.Halt()
	},
}