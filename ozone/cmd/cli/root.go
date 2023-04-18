package cli

import (
	"fmt"
	ozoneConfig "github.com/ozone2021/ozone/ozone-lib/config"
	"github.com/spf13/cobra"
	"log"
	"os"
)

var rootCmd = &cobra.Command{
	Use:   "ozone",
	Short: "Environment and run management, localhost orchestrator",
	Long:  ``,
}

var ozoneWorkingDir = ""
var config *ozoneConfig.OzoneConfig
var context string
var headless bool

func init() {
	rootCmd.PersistentFlags().StringP("context", "c", "", fmt.Sprintf("context (default is %s)", "TODO")) // TODO
	rootCmd.PersistentFlags().BoolP("detached", "d", false, "detached is for running headless, without docker daemon (you will likely want detached for server based ci/cd. Use the daemon for local)")

	var err error
	ozoneWorkingDir, err = os.Getwd()
	if err != nil {
		log.Fatalln(err)
	}

	dir, err := os.Getwd()
	if err != nil {
		log.Println(err)
	}
	ozoneWorkingDir = dir
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
