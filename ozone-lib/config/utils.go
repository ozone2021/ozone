package config

import (
	"github.com/flosch/pongo2/v4"
	"log"
	"os"
	"strings"
)

func convertMap(originalMap interface{}) pongo2.Context {
	convertedMap := make(map[string]interface{})
	for key, value := range originalMap.(map[string]string) {
		convertedMap[key] = value
	}

	return convertedMap
}

func ContextInPattern(context, pattern string, scope map[string]string) bool {
	pattern = renderVars(pattern, scope)
	patternArray := strings.Split(pattern, "|")
	for _, v := range patternArray {
		if context == v {
			return true
		}
	}
	return false
}

func renderVars(input string, varsMap map[string]string) string {
	//tpl, err := pongo2.FromString("Hello {{ name|capfirst }}!")
	tpl, err := pongo2.FromString(input)
	if err != nil {
		log.Fatalln(err)
	}
	context := convertMap(varsMap)
	out, err := tpl.Execute(context)
	if err != nil {
		log.Fatalln(err)
	}
	return out
}

func OSEnvToVarsMap() map[string]string {
	newMap := make(map[string]string)
	for _, kvString := range os.Environ() {
		parts := strings.Split(kvString, "=")
		key, value := parts[0], parts[1]
		newMap[key] = value
	}
	return newMap
}

func RenderNoMerge(base map[string]string, scope map[string]string) map[string]string {
	combinedScope := CopyMap(scope)
	osEnv := OSEnvToVarsMap()
	combinedScope = MergeMaps(combinedScope, osEnv)
	newMap := CopyMap(base)
	for k, v := range base {
		newMap[k] = renderVars(v, combinedScope)
	}
	return newMap
}

func CopyMap(toCopy map[string]string) map[string]string {
	newMap := make(map[string]string)
	for k,v := range toCopy {
		newMap[k] = v
	}
	return newMap
}

func MergeMapsSelfRender(base map[string]string, overwrite map[string]string) map[string]string {
	if base == nil {
		return CopyMap(overwrite)
	}
	newMap := CopyMap(base)
	for k, v := range overwrite {
		newMap[k] = v
	}
	newMap = RenderNoMerge(newMap, newMap)
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
