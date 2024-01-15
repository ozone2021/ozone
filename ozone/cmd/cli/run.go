package cli

import (
	"fmt"
	ozoneConfig "github.com/ozone2021/ozone/ozone-lib/config"
	"github.com/ozone2021/ozone/ozone-lib/run/runapp_controller"
	"github.com/spf13/cobra"
	"log"
	"sync"
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

		//runResult := spec.RunSpecRootNodeToRunResult(spec.CallStacks[ozoneConfig.BuildType][0])

		controller := runapp_controller.NewRunController(ozoneContext, ozoneWorkingDir, combinedArgs, config)

		var wg sync.WaitGroup
		controller.Start(&wg)

		//runResult.PrintErrorLog() TODO if headless, print error log
		//
		//fmt.Println("=================================================")
		//fmt.Println("====================  Run result  ===============")
		//fmt.Println("=================================================")

		controller.Run(runnables)

		wg.Wait()

		//runResult.PrintRunResult(true)
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}
