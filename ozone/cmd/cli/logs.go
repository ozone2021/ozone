package cli

import (
	"fmt"
	process_manager_client "github.com/ozone2021/ozone/ozone-daemon-lib/process-manager-client"
	"github.com/spf13/cobra"
	"log"
	"os"
	"os/exec"
)

func init() {
	rootCmd.AddCommand(logsCmd)
}

func logs(service string) {
	err, tempDir := process_manager_client.FetchTempDir()

	if err != nil {
		fmt.Println(err)
		return
	}

	logsPath := fmt.Sprintf("%s/%s-logs", tempDir, service)

	cmd := exec.Command("tail", "-f", logsPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stdout
	cmd.Run()
}

var logsCmd = &cobra.Command{
	Use:   "l",
	Long:  `Logs for given services`,
	Run: func(cmd *cobra.Command, args []string) {

		service := args[0]

		if config.DeploysHasService(service) {
			fmt.Printf("Logs for %s ...\n", service)
			logs(service)
		} else {
			log.Fatalf("No deploy with service name: %s \n", service)
		}
		//_go.Build("microA", "micro-a", "main.go")
		//executable.Build("microA")
	},
}