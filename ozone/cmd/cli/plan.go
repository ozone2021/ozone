package cli

import (
	ozoneConfig "github.com/ozone2021/ozone/ozone-lib/config"
	runspec2 "github.com/ozone2021/ozone/ozone-lib/config/runspec"
	"github.com/spf13/cobra"
	"log"
)

func init() {
	rootCmd.AddCommand(planCmd)
}

var planCmd = &cobra.Command{
	Use:  "plan",
	Long: `Shows a dry run of what is going to be ran.`,
	Run: func(cmd *cobra.Command, args []string) {
		runspec := runspec2.NewRunspec(ozoneContext, ozoneWorkingDir, config)

		var builds []*ozoneConfig.Runnable

		for _, arg := range args {
			if has, runnable := config.FetchRunnable(arg); has == true {
				builds = append(builds, runnable)
				continue
			} else {
				log.Fatalf("Config doesn't have runnable: %s \n", arg)
			}
		}

		runspec.AddCallstacks(builds, config, ozoneContext)

		runspec.PrintRunspec()
	},
}
