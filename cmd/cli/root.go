package cli

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

var rootCmd = &cobra.Command{
	Use:   "ozone",
	Short: "Environment and build management, localhost orchestrator",
	Long: ``,
}

func init() {
	rootCmd.PersistentFlags().StringP("env", "e", "local", "verbose output")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

