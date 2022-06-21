package config_variable

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/flosch/pongo2/v4"
	"log"
	"strings"
)

type VariableMap map[string]interface{}

func (vm *VariableMap) Explode() []string {
	var output []string

	for _, variable := range *vm {
		iface := any(variable)
		switch variable.(type) {
		case *GenVariable[string]:
			genvar := iface.(*GenVariable[string])
			output = append(output, genvar.GetValue())
		case *GenVariable[[]string]:
			genvar := iface.(*GenVariable[[]string])

			for _, i := range genvar.GetValue() {
				output = append(output, i)
			}
		}
	}

	return output
}

// Must set ordinal first.
func (vm *VariableMap) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var yamlObj map[string]interface{}
	if err := unmarshal(&yamlObj); err != nil {
		return err
	}

	log.Println(yamlObj)

	(*vm) = make(VariableMap)

	for name, value := range yamlObj {
		switch x := value.(type) {
		case string:
			stringVal, _ := value.(string)
			(*vm)[name] = NewGenVariable[string](stringVal, 0)
		case []interface{}:
			var stringSlice []string
			for _, item := range x {
				stringVal, _ := item.(string)
				stringSlice = append(stringSlice, stringVal)
			}
			(*vm)[name] = NewGenVariable[[]string](stringSlice, 0)
		}
	}

	return nil
}

type GenVariable[T VariableDataFormats] struct {
	value   T   `yaml:"value"`
	ordinal int `yaml:"ordinal"`
}

func (v *GenVariable[T]) SetValue(value T) {
	v.value = value
}

func (v *GenVariable[T]) GetValue() T {
	return v.value
}

func (v *GenVariable[T]) GetOrdinal() int {
	return v.ordinal
}

func (v *GenVariable[T]) UnmarshalJSON(bytes []byte) error {
	var value T
	err := json.Unmarshal(bytes, &value)
	if err != nil {
		return err
	}
	v.value = value
	return nil
}

func (v *GenVariable[T]) Copy() *GenVariable[T] {
	return &GenVariable[T]{
		value:   v.value,
		ordinal: v.ordinal,
	}
}

func InterfaceToGenvar[Genvar GenVarType](iface interface{}) Genvar {
	switch iface.(type) {
	case *GenVariable[string]:
		return iface.(Genvar)
	case *GenVariable[[]string]:
		return iface.(Genvar)
	}
	return nil
}

func GenVarToString(varMap VariableMap, key string) (string, error) {
	genvarInterface, ok := varMap[key]

	if !ok {
		return "", nil
	}

	genvar, ok := genvarInterface.(*GenVariable[string])

	if !ok {
		return "", errors.New(fmt.Sprintf("Key %s not a string value", key))
	}

	return genvar.GetValue(), nil
}

func GenVarToFstring(varMap VariableMap, key string, format string) (string, error) {
	varString, err := GenVarToString(varMap, key)

	if err != nil {
		return "", nil
	}

	return fmt.Sprintf(format, varString), nil
}

func NewGenVariable[T VariableDataFormats](value T, ordinal int) *GenVariable[T] {
	return &GenVariable[T]{
		value:   value,
		ordinal: ordinal,
	}
}

type VariableDataFormats interface {
	string | []string
}

type GenVarType interface {
	*GenVariable[string] | *GenVariable[[]string]
	GetOrdinal() int
}

func (v *GenVariable[T]) Render(varsMap VariableMap) (*GenVariable[T], error) {
	var ret T
	switch val := any(&ret).(type) {
	case *string:
		asStringGen := any(v).(*GenVariable[string])
		renderedValue, err := RenderSingleString(asStringGen.GetValue(), varsMap)
		if err != nil {
			return nil, err
		}
		asStringGen.SetValue(renderedValue)
	case *[]string:
		var newArray []string

		for key, item := range *val {
			var err error
			newArray[key], err = RenderSingleString(item, varsMap)
			if err != nil {
				return nil, err
			}
		}
		asStringGen := any(v).(*GenVariable[[]string])
		asStringGen.SetValue(newArray)
	default:
		return nil, errors.New("Unknown type in variable render.")
	}

	return v, nil
}

// TODO this is where we convert the lists to exploded semi colons.
// Normal env vars go straight across.
func convertMap(originalMap VariableMap) pongo2.Context {
	convertedMap := make(map[string]interface{})
	for key, variable := range originalMap {
		switch variable.(type) {
		case *GenVariable[string]:
			genVar := variable.(*GenVariable[string])
			convertedMap[key] = genVar.GetValue()
		case *GenVariable[[]string]:
			genVar := variable.(*GenVariable[[]string])
			convertedMap[key] = strings.Join(genVar.GetValue(), ";")
		}
	}

	return convertedMap
}

func RenderSingleString(input string, varsMap VariableMap) (string, error) {
	tpl, err := pongo2.FromString(input)
	if err != nil {
		return "", err
	}
	context := convertMap(varsMap)
	out, err := tpl.Execute(context)
	if err != nil {
		return "", err
	}
	if out == "" {
		return input, nil
	}
	return out, nil
}

//func NewListGenVariable(value []string, ordinal int) *Variable[[]string] {
//	return &GenVariable[[]string]{
//		value:   value,
//		ordinal: ordinal,
//	}
//}

//func Copy[E Variable](variable E) E {
//	return &E{
//		Value:   v.Value,
//		ordinal: 0,
//	}
//}
