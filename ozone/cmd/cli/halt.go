package cli

import (
	process_manager_client "github.com/JamesArthurHolland/ozone/ozone-daemon-lib/process-manager-client"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(haltCmd)
}

var haltCmd = &cobra.Command{
	Use:   "h",
	Long:  `Halts all services in current directory`,
	Run: func(cmd *cobra.Command, args []string) {
		service := args[0]
		process_manager_client.Halt(service)
	},
}