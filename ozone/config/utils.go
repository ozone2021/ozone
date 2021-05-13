package config

import (
	"github.com/flosch/pongo2/v4"
	"log"
)

func VarsToMap(vars []*Var) map[string]string {
	result := make(map[string]string)

	for _, v := range vars {
		result[v.Name] = v.Value
	}

	return result
}

func convertMap(originalMap interface{}) pongo2.Context {
	convertedMap := make(map[string]interface{})
	for key, value := range originalMap.(map[string]string) {
		convertedMap[key] = value
	}

	return convertedMap
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

func RenderNoMerge(base map[string]string, scope map[string]string) map[string]string {
	for k, v := range base {
		base[k] = renderVars(v, scope)
	}
	return base
}

func MergeMaps(base map[string]string, overwrite map[string]string) map[string]string {
	for k, v := range overwrite {
		base[k] = renderVars(v, base)
	}
	return base
}
