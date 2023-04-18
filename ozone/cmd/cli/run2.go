package cli

import (
	ozoneConfig "github.com/ozone2021/ozone/ozone-lib/config"
	"github.com/ozone2021/ozone/ozone-lib/config/config_utils"
	runspec2 "github.com/ozone2021/ozone/ozone-lib/runspec"
	"github.com/spf13/cobra"
	"log"
)

func init() {
	rootCmd.AddCommand(run2Cmd)
	run2Cmd.PersistentFlags().BoolP("detached", "d", false, "detached is for running headless, without docker daemon (you will likely want detached for server based ci/cd. Use the daemon for local)")
}

var run2Cmd = &cobra.Command{
	Use:  "run2",
	Long: `Shows a dry run of what is going to be ran.`,
	Run: func(cmd *cobra.Command, args []string) {

		context := config_utils.FetchContext(cmd, ozoneWorkingDir, config)

		runspec := runspec2.NewRunspec(context, ozoneWorkingDir, config)

		var builds []*ozoneConfig.Runnable

		for _, arg := range args {
			if has, runnable := config.FetchRunnable(arg); has == true {
				builds = append(builds, runnable)
				continue
			} else {
				log.Fatalf("Config doesn't have runnable: %s \n", arg)
			}
		}

		runspec.AddCallstacks(builds, config, context)
		runspec.ExecuteCallstacks()

	},
}
