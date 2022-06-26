package config_variable

import (
	"errors"
	"fmt"
	"github.com/flosch/pongo2/v4"
	"log"
	"regexp"
	"strings"
)

type VariableMap map[string]Variable

const VariablePattern = `\{\{\s*([^}|\s]*)\s*(\s*\\|\s*[^}]*)?\s*\}\}`
const WhiteSpace = `\S(\s+)`
const ReplacementSymbol = `®`

//func (vm *VariableMap) Explode() []string {
//for _, variable := range v.GetValue() {
//iface := any(variable)
//switch variable.(type) {
//case *GenVariable[string]:
//genvar := iface.(*GenVariable[string])
//output = append(output, genvar.GetValue())
//case *GenVariable[[]string]:
//genvar := iface.(*GenVariable[[]string])
//
//for _, i := range genvar.GetValue() {
//output = append(output, i)
//}
//}
//}

// Must set ordinal first.
func (vm *VariableMap) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var yamlObj map[string]interface{}
	if err := unmarshal(&yamlObj); err != nil {
		return err
	}

	(*vm) = make(VariableMap)

	for name, value := range yamlObj {
		switch x := value.(type) {
		case string:
			stringVal, _ := value.(string)
			(*vm)[name] = NewStringVariable(stringVal, 0)
		case []interface{}:
			var stringSlice []string
			for _, item := range x {
				stringVal, _ := item.(string)
				stringSlice = append(stringSlice, stringVal)
			}
			(*vm)[name] = NewSliceVariable(stringSlice, 0)
		}
	}

	return nil
}

type VarDeclaration struct {
	Declaration string
	VarName     string
	Filter      string
}

type VarType int

const (
	StringType VarType = iota
	SliceType  VarType = iota
)

type Variable struct {
	value []string `yaml:"value"`
	VarType
	ordinal int `yaml:"ordinal"`
}

func NewStringVariable(value string, ordinal int) Variable {
	return Variable{
		value:   []string{value},
		VarType: StringType,
		ordinal: ordinal,
	}
}

func NewSliceVariable(value []string, ordinal int) Variable {
	return Variable{
		value:   value,
		VarType: SliceType,
		ordinal: ordinal,
	}
}

func (v Variable) ToString() string {
	return v.Fstring("%s")
}

func (v Variable) Fstring(format string, seperators ...string) string {
	separator := ""
	switch len(seperators) {
	case 0:
		separator = ";"
	case 1:
		separator = seperators[0]
	default:
		log.Fatalln("Either one seperator or none must be passed.")
	}
	switch v.VarType {
	case StringType:
		return fmt.Sprintf(format, v.value[0])
	case SliceType:
		return strings.Join(v.GetSliceValue(), separator)
	default:
		log.Fatalln("Unknown type in variable ToString.")
	}
	log.Fatalln("Error: variable ToString.")
	return ""
}

//func(vm *VariableMap) RenderFilters() error {
//	for _, variable := range *vm {
//		emptyVarsMap := make(VariableMap)
//		variable.Render(emptyVarsMap)
//	}
//
//	return nil
//}

func (v Variable) Render(varsMap VariableMap) error {
	switch v.VarType {
	case StringType:
		renderedValue, err := RenderSentence(v.GetStringValue(), varsMap)
		if err != nil {
			return err
		}
		v.value = []string{renderedValue}
	case SliceType:
		var newArray []string

		for key, item := range v.GetSliceValue() {
			var err error
			newArray[key], err = RenderSentence(item, varsMap)
			if err != nil {
				return err
			}
		}
		v.value = newArray
	default:
		return errors.New("Unknown type in variable render.")
	}

	return nil
}

func (v Variable) SetStringValue(value string) {
	v.value = []string{value}
}

func (v Variable) GetStringValue() string {
	return v.value[0]
}

func (v Variable) GetSliceValue() []string {
	return v.value
}

func (v Variable) GetOrdinal() int {
	return v.ordinal
}

// TODO this is where we convert the lists to exploded semi colons.
// Normal env vars go straight across.
func ConvertMap(originalMap VariableMap) pongo2.Context {
	convertedMap := make(map[string]interface{})
	for key, variable := range originalMap {
		convertedMap[key] = variable.ToString()
	}

	return convertedMap
}

func collectVariableAndFilters(sentence string) []*VarDeclaration {
	r := regexp.MustCompile(VariablePattern)
	subs := r.FindAllStringSubmatch(sentence, -1)

	var collectedVars []*VarDeclaration
	for _, sub := range subs {
		collectedVars = append(collectedVars, &VarDeclaration{
			Declaration: sub[0],
			VarName:     sub[1],
			Filter:      sub[2],
		})
	}

	return collectedVars
}

func replaceVariablesWithSpecial(sentence string, collectedVarsWithBraces []*VarDeclaration) string {
	for _, variableDeclaration := range collectedVarsWithBraces {
		sentence = strings.ReplaceAll(sentence, variableDeclaration.Declaration, ReplacementSymbol)
	}

	return sentence
}

func RenderSentence(sentence string, varsMap VariableMap) (string, error) {
	collectedVars := collectVariableAndFilters(sentence)
	replacedWithSpecialChar := replaceVariablesWithSpecial(sentence, collectedVars)

	output := replacedWithSpecialChar
	for _, varDeclaration := range collectedVars {
		_, exists := varsMap[varDeclaration.VarName]
		var err error
		replacement := varDeclaration.Declaration
		if exists || varDeclaration.Filter != "" {
			replacement, err = PongoRender(replacement, varsMap)
			if err != nil {
				return "", err
			}
		}
		output = strings.Replace(output, ReplacementSymbol, replacement, 1)
	}
	return output, nil
}

func PongoRender(input string, varsMap VariableMap) (string, error) {
	tpl, err := pongo2.FromString(input)
	if err != nil {
		return "", err
	}
	context := ConvertMap(varsMap)
	out, err := tpl.Execute(context)
	if err != nil {
		return "", err
	}
	if out == "" {
		return input, nil
	}
	return out, nil
}
