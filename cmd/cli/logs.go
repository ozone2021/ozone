package cli

import (
	"fmt"
	"github.com/spf13/cobra"
	"log"
	"net/rpc"
	"os"
	"os/exec"
	process_manager "ozone-daemon-lib/process-manager"
	ozoneConfig "ozone-lib/config"
)

func init() {
	rootCmd.AddCommand(logsCmd)
}

func fetchTempDir() string {
	ozoneWorkingDir, err := os.Getwd()
	if err != nil {
		log.Println(err)
	}

	request := process_manager.TempDirRequest{
		ozoneWorkingDir,
	}

	var response process_manager.TempDirResponse

	client, err := rpc.DialHTTP("tcp", ":8000")
	if err != nil {
		log.Fatal("dialing:", err)
	}
	defer client.Close()
	err = client.Call("ProcessManager.TempDirRequest", request, &response)

	return response.TempDir
}

func logs(service string) {
	tempDir := fetchTempDir()

	logsPath := fmt.Sprintf("%s/%s-logs", tempDir, service)
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