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
	Long: ``,
}

var ozoneWorkingDir = ""
var config *ozoneConfig.OzoneConfig
var context string
var headless bool

func init() {
	config = ozoneConfig.ReadConfig()

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

