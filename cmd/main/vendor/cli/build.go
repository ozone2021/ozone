package cli

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	ozoneConfig "ozone-lib/config"
)

func init() {
	rootCmd.AddCommand(buildCmd)
}

var buildCmd = &cobra.Command{
	Use:   "b",
	Long:  `List running processes`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Building...")

		config := ozoneConfig.ReadConfig()

		s, _ := json.MarshalIndent(config, "", "\t");
		fmt.Println(s)

		for _, v := range config.Builds {
			fmt.Println("-")
			fmt.Println(v.Name)
		}
		//_go.Build("microA", "micro-a", "main.go")
		//executable.Build("microA")
	},
}