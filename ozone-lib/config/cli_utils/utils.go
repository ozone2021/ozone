package cli_utils

import (
	"fmt"
	"log"
)

const DEFAULT_INDENT = 2

type VariableStringCliOutput struct {
	Name    string `yaml:"name"`
	Value   string `yaml:"value"`
	Ordinal int    `yaml:"ordinal"`
}

type VariableStringSliceCliOutput struct {
	Name    string   `yaml:"name"`
	Value   []string `yaml:"value"`
	Ordinal int      `yaml:"ordinal"`
}

func IncreaseIndent(indent int) int {
	return indent + DEFAULT_INDENT
}

func printIndent(indent int) string {
	indentString := ""
	for i := 0; i < indent; i++ {
		indentString = fmt.Sprintf("%s ", indentString)
	}
	return indentString
}

func PrintWithIndent(s string, indent int) {
	log.Printf("%s%s\n", printIndent(indent), s)
}
