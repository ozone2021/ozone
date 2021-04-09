package cli

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	ozoneConfig "ozone-lib/config"
	_go "ozone-lib/go"
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

		env := rootCmd.PersistentFlags().Lookup("author").Value.String()

		s, _ := json.MarshalIndent(config, "", "\t");
		fmt.Println(s)

		for _, b := range config.Builds {
			fmt.Println(b.Name)
			fmt.Println("-")
			for _, es := range b.EnvSteps {
				if es.Env == env {
					for _, step := range es.Steps {
						if step.Type == "builtin" {
							switch step.Value {
							case "go":
								_go.Build("microA", "micro-a", "main.go")

							}
						}
					}
				}
			}
		}
		//_go.Build("microA", "micro-a", "main.go")
		//executable.Build("microA")
	},
}