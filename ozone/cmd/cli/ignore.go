package cli

import (
	process_manager_client "github.com/ozone2021/ozone/ozone-daemon-lib/process-manager-client"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(ignoreCmd)
}

var ignoreCmd = &cobra.Command{
	Use:   "i",
	Long:  `Logs for given services`,
	Run: func(cmd *cobra.Command, args []string) {

		service := args[0]

		process_manager_client.Ignore(ozoneWorkingDir, service)
		//_go.Build("microA", "micro-a", "main.go")
		//executable.Build("microA")
	},
}