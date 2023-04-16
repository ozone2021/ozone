package config_utils

import (
	process_manager_client "github.com/ozone2021/ozone/ozone-daemon-lib/process-manager-client"
	"github.com/ozone2021/ozone/ozone-lib/config"
	. "github.com/ozone2021/ozone/ozone-lib/config/config_variable"
	"github.com/spf13/cobra"
	"log"
	"strings"
)

func ContextInPattern(context, pattern string, scope *VariableMap) (bool, error) {
	pattern, err := PongoRender(pattern, scope.ConvertMapPongo())
	if err != nil {
		return false, err
	}

	patternArray := strings.Split(pattern, "|")
	for _, v := range patternArray {
		if context == v {
			return true, nil
		}
	}
	return false, nil
}

func FetchContext(cmd *cobra.Command, ozoneWorkingDir string, config *config.OzoneConfig) string {
	context := ""
	headless, _ := cmd.Flags().GetBool("detached")
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

	return context
}

//func RenderNoMerge(ordinal int, base VariableMap, scope VariableMap) VariableMap {
//	combinedScope := scope.Copy()
//	osEnv := OSEnvToVarsMap(ordinal)
//	err := combinedScope.MergeVariableMaps(osEnv)
//	if err != nil {
//		log.Fatalln(err)
//	}
//	newMap := base.Copy()
//	for _, variable := range newMap {
//		rendered, err := combinedScope.Render(variable)
//		if err != nil {
//			log.Fatalln(err)
//		}
//		variable = rendered
//	}
//	return newMap
//}

//func MergeMapsSelfRender(ordinal int, base *VariableMap, overwrite *VariableMap) *VariableMap {
//	if base.IsEmpty() {
//		return overwrite.Copy()
//	}
//	newMap := base.Copy()
//	for k, v := range overwrite {
//		newMap[k] = v
//	}
//	newMap = RenderNoMerge(ordinal, newMap, newMap)
//	return newMap
//}

//func RenderFilters(base *VariableMap) error {
//	mapWithFiltersApplied := CopyVariableMap(*base)
//
//	for _, variable := range mapWithFiltersApplied {
//		emptyVarsMap := make(VariableMap)
//		variable.Render(emptyVarsMap)
//	}
//
//	return nil
//}

// TODO make a method of vm* potentially
