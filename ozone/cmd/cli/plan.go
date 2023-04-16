package cli

import (
	"fmt"
	"github.com/ozone2021/ozone/ozone-lib/config/config_utils"
	worktree2 "github.com/ozone2021/ozone/ozone-lib/worktree"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(planCmd)
	planCmd.PersistentFlags().StringP("context", "c", "", fmt.Sprintf("context (default is %s)", config.ContextInfo.Default))
	planCmd.PersistentFlags().BoolP("detached", "d", false, "detached is for running headless, without docker daemon (you will likely want detached for server based ci/cd. Use the daemon for local)")
}

var planCmd = &cobra.Command{
	Use:  "plan",
	Long: `Shows a dry run of what is going to be ran.`,
	Run: func(cmd *cobra.Command, args []string) {
		context := config_utils.FetchContext(cmd, ozoneWorkingDir, config)

		worktree := worktree2.NewWorktree(context, config)

		worktree.PrintWorktree()
	},
}
