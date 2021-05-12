package cli

import (
	"fmt"
	"github.com/spf13/cobra"
	"log"
	process_manager_client "ozone-daemon-lib/process-manager-client"
)

func init() {
	rootCmd.AddCommand(statusCmd)
}

var statusCmd = &cobra.Command{
	Use:   "s",
	Long:  `Status of all services in current directory`,
	Run: func(cmd *cobra.Command, args []string) {
		err, status := process_manager_client.Status()
		if err != nil {
			log.Fatalln(err)
			return
		}
		fmt.Println(status)
	},
}