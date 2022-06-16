package config_variable

import (
	"encoding/json"
	"errors"
	"github.com/flosch/pongo2/v4"
)

type Variable[T VariableDataFormats] interface {
	SetValue(value T)
	GetValue() T
	GetOrdinal() int
	Copy() *GenVariable[T]
	UnmarshalJSON(bytes []byte) error
}

type GenVariable[T VariableDataFormats, GenericVarType GenVarType] struct {
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

func NewGenVariable[T](value T, ordinal int) *GenVariable[T] {
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
}

func RenderVariable[variableType GenVarType](variable variableType, varsMap map[string]Variable) error {
	switch interface{}(variable).(type) {
	case *GenVariable[string]:
		val, ok := variable.(*GenVariable[string])
		if !ok {
			return errors.New("Should be string")
		}
		renderedValue, err := RenderSingleString(val.GetValue(), varsMap)
		if err != nil {
			return err
		}
		val.SetValue(renderedValue)

	case *GenVariable[[]string]:
		castVar, ok := variable.(*GenVariable[[]string])
		if !ok {
			return errors.New("Should be string array")
		}
		var newArray []string

		for _, listItem := range castVar.GetValue() {
			newItem, err := RenderSingleString(listItem, varsMap)
			if err != nil {
				return err
			}
			newArray = append(newArray, newItem)
		}
		castVar.SetValue(newArray)
		return nil
	default:
		return errors.New("Unknown type in variable render.")
	}

	return errors.New("Unknown type in variable render.")
}

//func (v *GenVariable[T]) RenderVariable(varsMap map[string]Variable) (Variable, error) {
//	switch v.GetVarType() {
//	case StringType:
//		//val, ok := v.GetValue()
//		//if !ok {
//		//	return nil, errors.New("Should be string")
//		//}
//		val := v.GetValue()
//		renderedValue, err := RenderSingleString(val, varsMap)
//		if err != nil {
//			return nil, err
//		}
//		v.SetValue(renderedValue)
//
//	case GenVariable[[]string]:
//		var newArray []string
//
//		for k, v := range varsMap {
//			newArray[k] = RenderSingleString(v.GetValue(), varsMap)
//		}
//		v.Value = newArray
//	default:
//		return nil, errors.New("Unknown type in variable render.")
//	}
//
//	return nil, errors.New("Unknown type in variable render.")
//}

func convertMap(originalMap interface{}) pongo2.Context {
	convertedMap := make(map[string]interface{})
	for key, variable := range originalMap.(map[string]Variable) {
		convertedMap[key] = variable.GetValue() // TODO does this break for arrays?
	}

	return convertedMap
}

func RenderSingleString(input string, varsMap map[string]Variable) (string, error) {
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
