package cli

//import (
//	"github.com/ozone2021/ozone/ozone-lib/config/config_variable"
//	"github.com/spf13/cobra"
//	"log"
//	"os"
//)
//
//func init() {
//	rootCmd.AddCommand(envCmd)
//
//	envCmd.PersistentFlags().StringP("format", "f", "", "pongo2 format for env output")
//	envCmd.PersistentFlags().StringP("runnable", "r", "", "Runnable name")
//
//}
//
//var envCmd = &cobra.Command{
//	Use:  "e",
//	Long: `BuildScope - get environment variables back`,
//	Run: func(cmd *cobra.Command, args []string) {
//		log.SetFlags(0)
//		log.SetOutput(os.Stdout)
//
//		formatFlag, _ := cmd.Flags().GetString("format")
//		if formatFlag == "" {
//			log.Fatalln("--format must be set")
//		}
//
//		runnableFlag, _ := cmd.Flags().GetString("runnable")
//		if runnableFlag == "" {
//			log.Fatalln("--runnable must be set")
//		}
//
//		topLevelScope := config_variable.CopyOrCreateNew(config.BuildVars)
//		topLevelScope.AddVariable(config_variable.NewStringVariable("CONTEXT", context), 0)
//		topLevelScope.AddVariable(config_variable.NewStringVariable("OZONE_WORKING_DIR", ozoneWorkingDir), 0)
//
//		ok, b := config.FetchRunnable(runnableFlag)
//		if !ok {
//			log.Fatalf("Runnable %s not found.", runnableFlag)
//		}
//		asOutput := make(map[string]string)
//		_, env, err := runIndividual(b, 0, context, config, config_variable.CopyOrCreateNew(topLevelScope), asOutput, false)
//		if err != nil {
//			log.Fatalf("Error %s in runnable %s", err, b.Name)
//		}
//		rendered, err := env.RenderSentence(formatFlag)
//		if err != nil {
//			log.Fatalln(err)
//		}
//		log.Println(rendered)
//	},
//}
