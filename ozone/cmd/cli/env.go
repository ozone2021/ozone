package cli

import (
	"fmt"
	process_manager_client "github.com/ozone2021/ozone/ozone-daemon-lib/process-manager-client"
	ozoneConfig "github.com/ozone2021/ozone/ozone-lib/config"
	"github.com/ozone2021/ozone/ozone-lib/config/config_variable"
	"github.com/spf13/cobra"
	"log"
	"os"
)

func init() {
	rootCmd.AddCommand(envCmd)
	config = ozoneConfig.ReadConfig()
	envCmd.PersistentFlags().StringP("context", "c", "", fmt.Sprintf("context (default is %s)", config.ContextInfo.Default))
	envCmd.PersistentFlags().BoolP("detached", "d", false, "detached is for running headless, without docker daemon (you will likely want detached for server based ci/cd. Use the daemon for local)")
	envCmd.PersistentFlags().StringP("format", "f", "", "pongo2 format for env output")
	envCmd.PersistentFlags().StringP("runnable", "r", "", "Runnable name")

}

var envCmd = &cobra.Command{
	Use:  "e",
	Long: `Env - get environment variables back`,
	Run: func(cmd *cobra.Command, args []string) {
		log.SetFlags(0)
		log.SetOutput(os.Stdout)

		headless, _ = cmd.Flags().GetBool("detached")

		contextFlag, _ := cmd.Flags().GetString("context")
		if contextFlag == "" {
			if headless == true {
				log.Fatalln("--context must be set if --headless mode used")
			} else {
				var err error
				context, err = process_manager_client.FetchContext(ozoneWorkingDir)
				if err != nil {
					log.Fatalln("FetchContext error:", err)
				}
			}
		} else if contextFlag != "" {
			if !config.HasContext(contextFlag) {
				log.Fatalf("Context %s doesn't exist in Ozonefile", contextFlag)
			}
			context = contextFlag
		}
		if context == "" {
			context = config.ContextInfo.Default
		}

		formatFlag, _ := cmd.Flags().GetString("format")
		if formatFlag == "" {
			log.Fatalln("--format must be set")
		}

		runnableFlag, _ := cmd.Flags().GetString("runnable")
		if runnableFlag == "" {
			log.Fatalln("--runnable must be set")
		}

		topLevelScope := config_variable.CopyOrCreateNew(config.BuildVars)
		topLevelScope.AddVariable(config_variable.NewStringVariable("CONTEXT", context), 0)
		topLevelScope.AddVariable(config_variable.NewStringVariable("OZONE_WORKING_DIR", ozoneWorkingDir), 0)

		ok, b := config.FetchRunnable(runnableFlag)
		if !ok {
			log.Fatalf("Runnable %s not found.", runnableFlag)
		}
		asOutput := make(map[string]string)
		_, env, err := runIndividual(b, 0, context, config, config_variable.CopyOrCreateNew(topLevelScope), asOutput, false)
		if err != nil {
			log.Fatalf("Error %s in runnable %s", err, b.Name)
		}
		rendered, err := env.RenderSentence(formatFlag)
		if err != nil {
			log.Fatalln(err)
		}
		log.Println(rendered)
	},
}
