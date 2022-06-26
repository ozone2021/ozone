package config_utils

import (
	. "github.com/ozone2021/ozone/ozone-lib/config/config_variable"
	"log"
	"os"
	"strings"
)

func ContextInPattern(context, pattern string, scope VariableMap) (bool, error) {
	pattern, err := PongoRender(pattern, scope)
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

func OSEnvToVarsMap(ordinal int) VariableMap {
	newMap := make(VariableMap)
	for _, kvString := range os.Environ() {
		parts := strings.Split(kvString, "=")
		key, value := parts[0], parts[1]
		newMap[key] = NewStringVariable(value, ordinal)
	}
	return newMap
}

func RenderNoMerge(ordinal int, base VariableMap, scope VariableMap) VariableMap {
	combinedScope := CopyVariableMap(scope)
	osEnv := OSEnvToVarsMap(ordinal)
	combinedScope, err := MergeVariableMaps(combinedScope, osEnv)
	if err != nil {
		log.Fatalln(err)
	}
	newMap := CopyVariableMap(base)
	for _, variable := range newMap {
		err := variable.Render(combinedScope)
		if err != nil {
			log.Fatalln(err)
		}
	}
	return newMap
}

func CopyMap(toCopy map[string]string) map[string]string {
	newMap := make(map[string]string)
	for k, v := range toCopy {
		newMap[k] = v
	}
	return newMap
}

func CopyVariableMap(toCopy VariableMap) VariableMap {
	newMap := make(VariableMap)
	for k, v := range toCopy {
		newMap[k] = v
	}
	return newMap
}

func MergeMapsSelfRender(ordinal int, base VariableMap, overwrite VariableMap) VariableMap {
	if base == nil {
		return CopyVariableMap(overwrite)
	}
	newMap := CopyVariableMap(base)
	for k, v := range overwrite {
		newMap[k] = v
	}
	newMap = RenderNoMerge(ordinal, newMap, newMap)
	return newMap
}

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

func MergeMaps(base map[string]string, overwrite map[string]string) map[string]string {
	if base == nil {
		return CopyMap(overwrite)
	}
	newMap := CopyMap(base)
	for k, v := range overwrite {
		newMap[k] = v
	}
	return newMap
}

// TODO make a method of vm* potentially
func MergeVariableMaps(base VariableMap, overwrite VariableMap) (VariableMap, error) {
	if base == nil {
		return CopyVariableMap(overwrite), nil
	}
	newMap := CopyVariableMap(base)
	for key, overwriteVariable := range overwrite {
		_, exists := newMap[key]
		if exists {
			overwriteOrdinal := overwriteVariable.GetOrdinal()

			baseOrdinal := newMap[key].GetOrdinal()

			if overwriteOrdinal < baseOrdinal {
				newMap[key] = overwriteVariable
			}
		} else {
			newMap[key] = overwriteVariable
		}
	}
	return newMap, nil
}
