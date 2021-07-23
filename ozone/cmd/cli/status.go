package cli

import (
	"fmt"
	process_manager_client "github.com/ozone2021/ozone/ozone-daemon-lib/process-manager-client"
	"github.com/spf13/cobra"
	"log"
)

func init() {
	rootCmd.AddCommand(statusCmd)
}

var statusCmd = &cobra.Command{
	Use:   "s",
	Long:  `Status of all services in current directory`,
	Run: func(cmd *cobra.Command, args []string) {
		err, status := process_manager_client.Status(ozoneWorkingDir)
		if err != nil {
			log.Fatalln(err)
			return
		}
		fmt.Println(status)
	},
}