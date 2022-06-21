package config_utils

import (
	"errors"
	. "github.com/ozone2021/ozone/ozone-lib/config/config_variable"
	"log"
	"os"
	"strings"
)

func ContextInPattern(context, pattern string, scope VariableMap) (bool, error) {
	pattern, err := RenderSingleString(pattern, scope)
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

func OSEnvToVarsMap(ordinal int) VariableMap {
	newMap := make(VariableMap)
	for _, kvString := range os.Environ() {
		parts := strings.Split(kvString, "=")
		key, value := parts[0], parts[1]
		newMap[key] = NewGenVariable[string](value, ordinal)
	}
	return newMap
}

func RenderSentence(sentence string, scope VariableMap) string {
	genvar := NewGenVariable[string](sentence, 1)
	genvar.Render(scope)
	return genvar.GetValue()
}

func RenderNoMerge(ordinal int, base VariableMap, scope VariableMap) VariableMap {
	combinedScope := CopyVariableMap(scope)
	osEnv := OSEnvToVarsMap(ordinal)
	combinedScope, err := MergeVariableMaps(combinedScope, osEnv)
	if err != nil {
		log.Fatalln(err)
	}
	newMap := CopyVariableMap(base)
	for k, variable := range newMap {
		switch variable.(type) {
		case *GenVariable[string]:
			genVar := variable.(*GenVariable[string])
			genVar.Render(combinedScope)
		case *GenVariable[[]string]:
			genVar := variable.(*GenVariable[[]string])
			genVar.Render(combinedScope)
		}

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

func RenderFilters(base *VariableMap) error {
	mapWithFiltersApplied := CopyVariableMap(*base)

	for _, variable := range mapWithFiltersApplied {
		emptyVarsMap := make(VariableMap)
		switch variable.(type) {
		case *GenVariable[string]:
			genVar := variable.(*GenVariable[string])
			genVar.Render(emptyVarsMap)
		case *GenVariable[[]string]:
			// TODO
		}

	}

	return nil
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

func CopyCopy[GenVar GenVarType](genvarIface interface{}, ptr *GenVar) {
	switch genvarIface.(type) {
	case *GenVariable[string]:
		genVar := genvarIface.(GenVariable[string])
		genVarPtr := any(ptr).(*GenVariable[string])
		genVarPtr = &genVar
		ptr = any(genVarPtr).(*GenVar)
	case *GenVariable[[]string]:
		genVar := genvarIface.(GenVariable[[]string])
		genVarPtr := any(ptr).(*GenVariable[[]string])
		genVarPtr = &genVar
		ptr = any(genVarPtr).(*GenVar)
	}
}

func GetOrdinal[Genvar GenVarType](genvar Genvar) (int, error) {
	return genvar.GetOrdinal(), nil
}

func GetOrdinalInterface(iface interface{}) (int, error) {
	stringGenVar, ok := iface.(*GenVariable[string])
	if ok {
		return GetOrdinal(stringGenVar)
	}
	sliceGenVar, ok := iface.(*GenVariable[[]string])
	if ok {
		return GetOrdinal(sliceGenVar)
	}
	return 0, errors.New("GetOrdinalInterface error")
}

func MergeVariableMaps(base VariableMap, overwrite VariableMap) (VariableMap, error) {
	if base == nil {
		return CopyVariableMap(overwrite), nil
	}
	newMap := CopyVariableMap(base)
	for key, overwriteValue := range overwrite {
		_, exists := newMap[key]
		if exists {
			overwriteOrdinal, err := GetOrdinalInterface(overwriteValue)
			if err != nil {
				return nil, err
			}

			baseOrdinal, err := GetOrdinalInterface(base)
			if err != nil {
				return nil, err
			}

			if overwriteOrdinal < baseOrdinal {
				newMap[key] = overwriteValue
			}
		} else {
			newMap[key] = overwriteValue
		}
	}
	return newMap, nil
}
