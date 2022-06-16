package config_utils

import (
	"github.com/flosch/pongo2/v4"
	"github.com/ozone2021/ozone/ozone-lib/config/config_variable"
	"os"
	"strings"
)

func convertMap(originalMap interface{}) pongo2.Context {
	convertedMap := make(map[string]interface{})
	for key, variable := range originalMap.(map[string]config_variable.Variable) {
		convertedMap[key] = variable.GetValue() // TODO does this break for arrays?
	}

	return convertedMap
}

func ContextInPattern(context, pattern string, scope map[string]config_variable.Variable) (bool, error) {
	pattern, err := config_variable.RenderSingleString(pattern, scope)
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

//func renderVars(input string, varsMap map[string]variable.Variable) string {
//	//tpl, err := pongo2.FromString("Hello {{ name|capfirst }}!")
//	tpl, err := pongo2.FromString(input)
//	if err != nil {
//		log.Fatalln(err)
//	}
//	context := convertMap(varsMap)
//	out, err := tpl.Execute(context)
//	if err != nil {
//		log.Fatalln(err)
//	}
//	if out == "" {
//		return input
//	}
//	return out
//}

//func OSEnvToVarsMap() map[string]string {
//	newMap := make(map[string]string)
//	for _, kvString := range os.Environ() {
//		parts := strings.Split(kvString, "=")
//		key, value := parts[0], parts[1]
//		newMap[key] = value
//	}
//	return newMap
//}

func OSEnvToVarsMap(ordinal int) map[string]config_variable.Variable {
	newMap := make(map[string]config_variable.Variable)
	for _, kvString := range os.Environ() {
		parts := strings.Split(kvString, "=")
		key, value := parts[0], parts[1]
		newMap[key] = config_variable.NewSingleVariable(value, ordinal)
	}
	return newMap
}

func RenderNoMerge(ordinal int, base map[string]config_variable.Variable, scope map[string]config_variable.Variable) map[string]config_variable.Variable {
	combinedScope := CopyVariableMap(scope)
	osEnv := OSEnvToVarsMap(ordinal)
	combinedScope = MergeVariableMaps(combinedScope, osEnv)
	newMap := CopyVariableMap(base)
	for k, variable := range newMap {
		config_variable.RenderVariable(variable, combinedScope)
		newMap[k] = variable
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

func CopyVariableMap(toCopy map[string]config_variable.Variable) map[string]config_variable.Variable {
	newMap := make(map[string]config_variable.Variable)
	for k, v := range toCopy {
		if variable, ok := v.(config_variable.Variable); ok {
			newMap[k] = variable.Copy()
		}
	}
	return newMap
}

func MergeMapsSelfRender(ordinal int, base map[string]config_variable.Variable, overwrite map[string]config_variable.Variable) map[string]config_variable.Variable {
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

func MergeVariableMaps(base map[string]config_variable.Variable, overwrite map[string]config_variable.Variable) map[string]config_variable.Variable {
	if base == nil {
		return CopyVariableMap(overwrite)
	}
	newMap := CopyVariableMap(base)
	for k, v := range overwrite {
		_, exists := newMap[k]
		if exists {
			if v.GetOrdinal() < newMap[k].GetOrdinal() {
				newMap[k] = v
			}
		} else {
			newMap[k] = v
		}
	}
	return newMap
}

