package cli

import (
	"fmt"
	"github.com/spf13/cobra"
	process_manager_client "github.com/JamesArthurHolland/ozone/ozone-daemon-lib/process-manager-client"
)

func init() {
	rootCmd.AddCommand(contextCmd)
}

var contextCmd = &cobra.Command{
	Use:   "c",
	Long:  `Show context or change context`,
	Run: func(cmd *cobra.Command, args []string) {

		givenContext := args[0]

		if givenContext == "" {
			fmt.Println(context)
		} else {
			// TODO check context is in
			if config.HasContext(givenContext) {
				process_manager_client.SetContext(ozoneWorkingDir, givenContext)
				fmt.Printf("Switch to context: '%s'", givenContext)
			} else {
				fmt.Printf("Context '%s' doesn't exist in Ozonefile.", givenContext)
			}
		}
	},
}