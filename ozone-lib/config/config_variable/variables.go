package config_variable

import (
	"errors"
	"fmt"
	"github.com/dlclark/regexp2"
	"github.com/flosch/pongo2/v4"
	"github.com/ozone2021/ozone/ozone-lib/config/cli_utils"
	"html"
	"log"
	"math"
	"os"
	"reflect"
	"strings"
)

const VariablePattern = `\{{2}(?!{)\s*([^}|^{|\s]*)\s*(\s*\\|\s*[^}]*)?\s*\}\}`
const WhiteSpace = `\S(\s+)`
const ReplacementSymbol = `®`
const ConfigOrdinal = math.MaxInt

type VariableMap struct {
	variables map[string]*Variable
	ordinals  map[string]int
}

const OrdinalityTag = "ordinality"

func NewVariableMap() *VariableMap {
	return &VariableMap{
		variables: make(map[string]*Variable),
		ordinals:  make(map[string]int),
	}
}

func (vm *VariableMap) IsEmpty() bool {
	return len(vm.variables) == 0
}

func (vm *VariableMap) AddVariable(variable *Variable, ordinal int) {
	if variable == nil {
		return
	}
	_, exists := vm.variables[variable.name]
	if !exists || exists && ordinal <= vm.ordinals[variable.name] {
		vm.variables[variable.name] = variable
		vm.ordinals[variable.name] = ordinal
	}
}

func (vm *VariableMap) DeleteVariableByName(variableName string) error {
	_, exists := vm.variables[variableName]
	if !exists {
		return errors.New(fmt.Sprintf("Variable %s doesn't exist in DeleteVariableByName \n", variableName))
	}
	delete(vm.variables, variableName)
	delete(vm.ordinals, variableName)

	return nil
}

func (vm *VariableMap) Print(indent int) {
	indent = cli_utils.IncreaseIndent(indent)
	for _, variable := range vm.variables {
		cli_utils.PrintWithIndent(fmt.Sprintf("%s=%s", variable.name, variable.value), indent)
	}
}

func (vm *VariableMap) AddVariableWithoutOrdinality(variable *Variable) {
	if variable == nil {
		return
	}
	//rendered, err := vm.Render(variable)
	//if err != nil {
	//	return err
	//}
	vm.variables[variable.name] = variable
	vm.ordinals[variable.name] = 1
	return
}

// TODO this isn't really a diff, it only shows what exists
func (vm *VariableMap) Diff(otherMap *VariableMap) (*VariableMap, error) {
	diffMap := NewVariableMap()
	for name, otherVariable := range otherMap.variables {
		ourVariable, exists := vm.GetVariable(name)
		if !exists || !otherVariable.Equals(ourVariable) {
			ordinal, err := otherMap.GetOrdinal(name)
			if err != nil {
				return nil, err
			}
			diffMap.AddVariable(otherVariable, ordinal)
		}
	}
	return diffMap, nil
}

func (vm *VariableMap) GetOrdinal(name string) (int, error) {
	ordinal, exists := vm.ordinals[name]
	if !exists {
		return 0, errors.New(fmt.Sprintf("Ordinal for %s doesn't exist", name))
	}
	return ordinal, nil
}

func (vm *VariableMap) GetVariable(name string) (*Variable, bool) {
	variable, ok := vm.variables[name]
	return variable, ok
}
func (vm *VariableMap) AsOutput(varOutputAs map[string]string) (*VariableMap, error) {
	outputMap := NewVariableMap()
	for currentName, targetName := range varOutputAs {
		variable, exists := vm.GetVariable(currentName)
		if exists {
			copiedVar := variable.Copy()
			copiedVar.name = targetName
			ordinal, err := vm.GetOrdinal(currentName)
			if err != nil {
				return nil, err
			}
			outputMap.AddVariable(copiedVar, ordinal)
		} else {
			return nil, errors.New(fmt.Sprintf("Variable %s doesn't exist in VariableMap", currentName))
		}
	}
	return outputMap, nil
}

//func (vm *VariableMap) getOrdinal(name string) int {
//	return vm.ordinals[name]
//}

func (vm *VariableMap) IncrementOrdinal(incrementBy int) {
	for key, ordinal := range vm.ordinals {
		if ordinal == math.MaxInt {
			vm.ordinals[key] = incrementBy
		} else {
			vm.ordinals[key] += incrementBy
		}
	}
}

func (vm *VariableMap) ConvertMap() map[string]string {
	convertedMap := make(map[string]string)
	for key, variable := range vm.variables {
		convertedMap[key] = variable.String()
	}

	return convertedMap
}

func (vm *VariableMap) ConvertMapPongo() pongo2.Context {
	convertedMap := make(map[string]interface{})
	for key, variable := range vm.variables {
		convertedMap[key] = variable.String()
	}

	return convertedMap
}

func (vm *VariableMap) MarshalYAML() (interface{}, error) {
	var variables []interface{}

	for name, value := range vm.variables {
		ordinal, ok := vm.ordinals[name]
		if !ok {
			return nil, errors.New(fmt.Sprintf("Couldn't find ordinal for variable: %s", name))
		}
		switch value.varType {
		case StringType:
			variables = append(variables, &cli_utils.VariableStringCliOutput{
				Name:    name,
				Value:   value.GetStringValue(),
				Ordinal: ordinal,
			})
		case SliceType:
			variables = append(variables, &cli_utils.VariableStringSliceCliOutput{
				Name:    name,
				Value:   value.GetSliceValue(),
				Ordinal: ordinal,
			})
		}
	}
	return variables, nil
}

// Must set ordinal first.
func (vm *VariableMap) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var yamlObj map[string]interface{}
	if err := unmarshal(&yamlObj); err != nil {
		return err
	}

	*vm = *NewVariableMap()
	for name, value := range yamlObj {
		switch x := value.(type) {
		case string:
			vm.AddVariable(NewStringVariable(name, x), ConfigOrdinal)
		case []interface{}:
			var stringSlice []string
			for _, item := range x {
				stringVal, _ := item.(string)
				stringSlice = append(stringSlice, stringVal)
			}
			vm.AddVariable(NewSliceVariable(name, stringSlice), ConfigOrdinal)
		}
	}

	return nil
}

// Does this mess the overwrite varmap up?
func (vm *VariableMap) MergeVariableMaps(overwrite *VariableMap) error {
	if overwrite == nil {
		return nil
	}
	for _, overwriteVariable := range overwrite.variables {
		variable, err := vm.Render(overwriteVariable)
		if err != nil {
			return err
		}
		vm.AddVariable(variable, overwrite.ordinals[overwriteVariable.name])
	}
	for i := 0; i < 2; i++ { // VariableVariables
		err := vm.SelfRender()
		if err != nil {
			return err
		}
	}

	return nil
}

func (vm *VariableMap) RenderNoMerge(scope *VariableMap) error {
	combinedScope := scope.copy()
	//osEnv := OSEnvToVarsMap()
	//err := combinedScope.MergeVariableMaps(osEnv)
	//if err != nil {
	//	return err
	//}
	for _, variable := range vm.variables {
		rendered, err := combinedScope.Render(variable)
		if err != nil {
			return err
		}
		*variable = *rendered
	}
	return nil
}

func CopyOrCreateNew(vm *VariableMap) *VariableMap { // Should create with ordinal
	if vm == nil {
		return NewVariableMap()
	}
	return vm.copy()
}

func (vm *VariableMap) copy() *VariableMap {
	newMap := NewVariableMap()
	for name, variable := range vm.variables {
		newMap.variables[name] = variable.Copy()
		newMap.ordinals[name] = vm.ordinals[name]
	}
	return newMap
}

func OSEnvToVarsMap() *VariableMap {
	newMap := NewVariableMap()
	for _, kvString := range os.Environ() {
		parts := strings.Split(kvString, "=")
		key, value := parts[0], parts[1]
		newMap.AddVariable(NewStringVariable(key, value), ConfigOrdinal)
	}
	return newMap
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
	name    string
	value   []string `yaml:"value"`
	varType VarType
}

func NewStringVariable(name, value string) *Variable {
	return &Variable{
		name:    name,
		value:   []string{value},
		varType: StringType,
	}
}

func NewSliceVariable(name string, value []string) *Variable {
	return &Variable{
		name:    name,
		value:   value,
		varType: SliceType,
	}
}

func (v *Variable) Equals(other *Variable) bool {
	if v.name != other.name || v.varType != other.varType {
		return false
	}
	for k, val := range v.value {
		if val != other.value[k] {
			return false
		}
	}
	return true
}

func (v *Variable) Copy() *Variable {
	return &Variable{
		name:    v.name,
		value:   v.value,
		varType: v.varType,
	}
}

func (v *Variable) GetVarType() VarType {
	return v.varType
}

func (v *Variable) SetVarType(t VarType) {
	v.varType = t
}

func (v *Variable) String() string {
	return v.Fstring("%s")
}

func (v *Variable) Fstring(format string, seperators ...string) string {
	separator := ""
	switch len(seperators) {
	case 0:
		separator = ";"
	case 1:
		separator = seperators[0]
	default:
		log.Fatalln("Either one seperator or none must be passed.")
	}
	if v == nil {
		log.Printf("here")
	}
	switch v.GetVarType() {
	case StringType:
		if len(v.value) == 0 {
			return ""
		}
		return fmt.Sprintf(format, v.value[0])
	case SliceType:
		return strings.Join(v.GetSliceValue(), separator)
	default:
		log.Fatalln("Unknown type in variable ToString.")
	}
	log.Fatalln("Error: variable ToString.")
	return ""
}

// TODO maybe a custom tag of ozone?
func (v *Variable) getYamlTag() (string, error) {
	t := reflect.TypeOf(*v)

	// Get the type and kind of our user variable
	fmt.Println("Type:", t.Name())
	fmt.Println("Kind:", t.Kind())

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("yaml")
		log.Println(tag)
		fmt.Printf("%d. %v (%v), tag: '%v'\n", i+1, field.Name, field.Type.Name(), tag)
	}
	return "", errors.New("Could not get yaml tag.")
}

func (v *Variable) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var yamlObj interface{}
	if err := unmarshal(&yamlObj); err != nil {
		return err
	}

	switch value := yamlObj.(type) {
	case string:
		v.SetStringValue(value)
	case []interface{}:
		var stringSlice []string
		for _, item := range value {
			stringVal, _ := item.(string)
			stringSlice = append(stringSlice, stringVal)
		}
		v.SetSliceValue(stringSlice)
	}

	tag, err := v.getYamlTag()
	if err != nil {
		return err
	}
	v.name = tag

	return nil
}

func (vm *VariableMap) RenderFilters() error {
	emptyVm := OSEnvToVarsMap()
	for _, variable := range vm.variables {
		switch variable.GetVarType() {
		case StringType:
			rendered, err := PongoRender(variable.String(), emptyVm.ConvertMapPongo())
			if err != nil {
				return err
			}
			variable.SetStringValue(rendered)
		case SliceType:
			var newArray []string
			for _, item := range variable.GetSliceValue() {
				rendered, err := PongoRender(item, emptyVm.ConvertMapPongo())
				if err != nil {
					return err
				}
				newArray = append(newArray, rendered)
			}
			variable.SetSliceValue(newArray)
		default:
			return errors.New("Unknown type in variable render.")
		}
	}
	return nil
}

func (vm *VariableMap) SelfRender() error {
	for _, variable := range vm.variables {
		rendered, err := vm.Render(variable)
		if err != nil {
			return err
		}
		*variable = *rendered
	}
	return nil
}

func (vm *VariableMap) Render(v *Variable) (*Variable, error) {
	output := v.Copy()
	switch v.GetVarType() {
	case StringType:
		renderedValue, err := vm.RenderSentence(v.GetStringValue())
		if err != nil {
			return nil, err
		}
		output.value = []string{renderedValue}
	case SliceType:
		rendered, err := vm.RenderList(v.GetSliceValue())
		if err != nil {
			return nil, err
		}
		output.value = rendered
	default:
		return nil, errors.New("Unknown type in variable render.")
	}

	var err error
	output.name, err = vm.RenderSentence(v.name)
	if err != nil {
		return nil, err
	}

	return output, nil
}

//func (v *Variable) Render(varsMap VariableMap) error {
//	switch v.GetVarType() {
//	case StringType:
//		renderedValue, err := RenderSentence(v.GetStringValue(), varsMap)
//		if err != nil {
//			return err
//		}
//		v.value = []string{renderedValue}
//	case SliceType:
//		var newArray []string
//
//		for _, item := range v.GetSliceValue() {
//			rendered, err := RenderSentence(item, varsMap)
//			if err != nil {
//				return err
//			}
//			newArray = append(newArray, rendered)
//		}
//		v.value = newArray
//	default:
//		return errors.New("Unknown type in variable render.")
//	}
//
//	return nil
//}

func (vm *VariableMap) RenderList(list []string) ([]string, error) {
	var parts []string
	for _, item := range list {
		renderedSentence, err := vm.RenderSentence(item)
		if err != nil {
			return nil, err // TODO wrap error
		}

		if strings.Contains(renderedSentence, ";") {
			parts = append(parts, strings.Split(renderedSentence, ";")...)
		} else {
			parts = append(parts, renderedSentence)
		}
	}
	return parts, nil
}

func (vm *VariableMap) RenderSentence(sentence string) (string, error) {
	collectedVars := collectVariableAndFilters(sentence)
	replacedWithSpecialChar := replaceVariablesWithSpecial(sentence, collectedVars)

	output := replacedWithSpecialChar
	for _, varDeclaration := range collectedVars {
		_, exists := vm.variables[varDeclaration.VarName]
		var err error
		replacement := varDeclaration.Declaration
		if exists || varDeclaration.Filter != "" {
			replacement, err = PongoRender(replacement, vm.ConvertMapPongo())
			if err != nil {
				return "", err
			}
		}
		output = strings.Replace(output, ReplacementSymbol, replacement, 1)
	}
	return output, nil
}

func (v *Variable) SetStringValue(value string) {
	v.value = []string{value}
}

func (v *Variable) SetSliceValue(value []string) {
	v.value = value
}

func (v *Variable) GetStringValue() string {
	return v.value[0]
}

func (v *Variable) GetSliceValue() []string {
	return v.value
}

// TODO this is where we convert the lists to exploded semi colons.
// Normal env vars go straight across.

func regexp2FindAllString(re *regexp2.Regexp, s string) []*regexp2.Match {
	var matches []*regexp2.Match
	m, _ := re.FindStringMatch(s)
	for m != nil {
		matches = append(matches, m)
		m, _ = re.FindNextMatch(m)
	}
	return matches
}

func collectVariableAndFilters(sentence string) []*VarDeclaration {
	r := regexp2.MustCompile(VariablePattern, 0)
	matches := regexp2FindAllString(r, sentence)

	var collectedVars []*VarDeclaration

	if matches != nil {
		for _, match := range matches {
			subs := match.Groups()
			collectedVars = append(collectedVars, &VarDeclaration{
				Declaration: subs[0].Captures[0].String(),
				VarName:     subs[1].Captures[0].String(),
				Filter:      subs[2].Captures[0].String(),
			})
		}
	}

	return collectedVars
}

func replaceVariablesWithSpecial(sentence string, collectedVarsWithBraces []*VarDeclaration) string {
	for _, variableDeclaration := range collectedVarsWithBraces {
		sentence = strings.ReplaceAll(sentence, variableDeclaration.Declaration, ReplacementSymbol)
	}

	return sentence
}

func PongoRender(input string, context pongo2.Context) (string, error) {
	tpl, err := pongo2.FromString(input)
	if err != nil {
		return "", err
	}
	out, err := tpl.Execute(context)
	if err != nil {
		return "", err
	}
	containsDefaultIfNoneFilter := strings.Contains(input, "default_if_none")
	if out == "" && containsDefaultIfNoneFilter == false {
		return input, nil
	}
	return html.UnescapeString(out), nil
}
