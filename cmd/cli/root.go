package cli

import (
	"fmt"
	"github.com/spf13/cobra"
	"log"
	"os"
	process_manager_client "ozone-daemon-lib/process-manager-client"
	ozoneConfig "ozone-lib/config"
)

var rootCmd = &cobra.Command{
	Use:   "ozone",
	Short: "Environment and run management, localhost orchestrator",
	Long: ``,
}

var ozoneWorkingDir = ""
var config *ozoneConfig.OzoneConfig
var context string

func init() {
	config = ozoneConfig.ReadConfig()

	var err error
	ozoneWorkingDir, err = os.Getwd()
	if err != nil {
		log.Fatalln(err)
	}

	context, err = process_manager_client.FetchContext(ozoneWorkingDir)
	if err != nil {
		log.Fatalln("FetchContext error:", err)
	}
	if context == "" {
		context = config.ContextInfo.Default
	}

	rootCmd.PersistentFlags().StringP("env", "e", "local", "verbose output")

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

