package cli

import (
	"fmt"
	ozoneConfig "github.com/ozone2021/ozone/ozone-lib/config"
	"github.com/ozone2021/ozone/ozone-lib/run/runapp_controller"
	"github.com/spf13/cobra"
	"log"
	"os"
	"os/signal"
	"syscall"
)

var runCmd = &cobra.Command{
	Use:  "run",
	Long: `Shows a dry run of what is going to be ran.`,
	Run: func(cmd *cobra.Command, args []string) {

		var runnables []*ozoneConfig.Runnable

		combinedArgs := ""
		for _, arg := range args {
			combinedArgs += fmt.Sprintf("%s ", arg)
			if has, runnable := config.FetchRunnable(arg); has == true {
				runnables = append(runnables, runnable)
				continue
			} else {
				log.Fatalf("Config doesn't have runnable: %s \n", arg)
			}
		}

		controller := runapp_controller.NewRunController(ozoneContext, ozoneWorkingDir, combinedArgs, config)

		go controller.Start()

		controller.Run(runnables)

		sig := make(chan os.Signal, 2)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
		<-sig
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}
