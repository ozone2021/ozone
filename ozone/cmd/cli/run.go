package cli

import (
	"fmt"
	ozoneConfig "github.com/ozone2021/ozone/ozone-lib/config"
	"github.com/ozone2021/ozone/ozone-lib/runspec"
	"github.com/spf13/cobra"
	"log"
)

var runCmd = &cobra.Command{
	Use:  "run",
	Long: `Shows a dry run of what is going to be ran.`,
	Run: func(cmd *cobra.Command, args []string) {

		spec := runspec.NewRunspec(context, ozoneWorkingDir, config)

		var runnables []*ozoneConfig.Runnable

		for _, arg := range args {
			if has, runnable := config.FetchRunnable(arg); has == true {
				runnables = append(runnables, runnable)
				continue
			} else {
				log.Fatalf("Config doesn't have runnable: %s \n", arg)
			}
		}

		spec.AddCallstacks(runnables, config, context)
		//runResult := spec.RunSpecRootNodeToRunResult(spec.CallStacks[ozoneConfig.BuildType][0])
		runResult := spec.ExecuteCallstacks()

		runResult.PrintErrorLog()

		fmt.Println("=================================================")
		fmt.Println("====================  Run result  ===============")
		fmt.Println("=================================================")

		runResult.PrintRunResult()
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}
