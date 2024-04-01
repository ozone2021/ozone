package cli

import (
	"fmt"
	process_manager_client "github.com/ozone2021/ozone/ozone-daemon-lib/process-manager-client"
	"github.com/ozone2021/ozone/ozone-lib/logs/logapp_controller"
	"github.com/spf13/cobra"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
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
	Use:  "logs",
	Long: `Logs for given services`,
	Run: func(cmd *cobra.Command, args []string) {
		controller := logapp_controller.NewLogAppController(ozoneWorkingDir)

		go controller.Start()

		sig := make(chan os.Signal, 2)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
		<-sig
	},
}
