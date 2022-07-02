package config_utils

import (
	. "github.com/ozone2021/ozone/ozone-lib/config/config_variable"
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