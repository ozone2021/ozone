package cli

import (
	"fmt"
	"github.com/spf13/cobra"
	"log"
	"os"
	"os/exec"
	process_manager "ozone-daemon-lib/process-manager"
	process_manager_client "ozone-daemon-lib/process-manager-client"
	ozoneConfig "ozone-lib/config"
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
	process_manager.CreateLogFileIfNotExists(logsPath)


	log.Println("---")
	log.Println(logsPath)
	log.Println("---")

	cmd := exec.Command("tail", "-f", logsPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stdout
	cmd.Run()
}

var logsCmd = &cobra.Command{
	Use:   "l",
	Long:  `Logs for given services`,
	Run: func(cmd *cobra.Command, args []string) {

		config := ozoneConfig.ReadConfig()
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