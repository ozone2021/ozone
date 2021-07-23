package cli

import (
	process_manager_client "github.com/ozone2021/ozone/ozone-daemon-lib/process-manager-client"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(haltCmd)
}

var haltCmd = &cobra.Command{
	Use:   "h",
	Long:  `Halts all services in current directory`,
	Run: func(cmd *cobra.Command, args []string) {
		service := ""
		if len(args) == 1 {
			service = args[0]
		}
		process_manager_client.Halt(service)
	},
}