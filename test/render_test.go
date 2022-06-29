package test

import (
	"github.com/ozone2021/ozone/ozone-lib/config/config_variable"
	"log"
	"regexp"
	"testing"
)

// Make sure non-existent vars don't get erased.

func TestWhitespaceRender(t *testing.T) {
	r := regexp.MustCompile(config_variable.WhiteSpace)
	var output []string
	matches := r.FindAllString("aaa x  x    aaaa", -1)
	for _, match := range matches {
		output = append(output, match[1:])
	}

	log.Println(matches)
}

//

func TestSpecialCharRender(t *testing.T) {

}

func TestSentenceRender(t *testing.T) {
	//varsMap := make(map[string]interface{})
	////genVarNeedRendered := config_variable.NewGenVariable[string]("its {{new}}/{{nothere}} !!!", 1)
	//sentence := `aaa {{here}} {{ heretoo }}  {{DOCKER_REGISTRY | default_if_none:"registry.local"}} {{should_stay}} aaaa`
	//firstGenvar := config_variable.NewGenVariable[string]("first", 1)
	//secondGenvar := config_variable.NewGenVariable[string]("second", 1)
	//varsMap["here"] = firstGenvar
	//varsMap["heretoo"] = secondGenvar
	//
	//collectedVars := collectVariableAndFilters(sentence)
	//replacedWithSpecialChar := replaceVariablesWithSpecial(sentence, collectedVars)
	//
	//for _, varDeclaration := range collectedVars {
	//	_, exists := varsMap[varDeclaration.VarName]
	//	var err error
	//	replacement := varDeclaration.Declaration
	//	if exists || varDeclaration.Filter != "" {
	//		replacement, err = config_variable.PongoRender(replacement, varsMap)
	//		if err != nil {
	//			t.Error(err)
	//		}
	//	}
	//	replacedWithSpecialChar = strings.Replace(replacedWithSpecialChar, config_variable.ReplacementSymbol, replacement, 1)
	//}
	//
	//log.Println(collectedVars)
	//log.Println(replacedWithSpecialChar)
}

func TestFilter(t *testing.T) {
	//input := `{{DOMAIN | default_if_none:"virtuosoqa.local"}}`
	input := "{{DOCKER_REGISTRY | default_if_none:\"registry.local\"}}"
	vm := config_variable.NewVariableMap()
	output, err := config_variable.PongoRender(input, vm.ConvertMap())
	if err != nil {
		t.Error(err)
	}
	log.Println(output)
}

func TestVarMatch(t *testing.T) {

	r := regexp.MustCompile(config_variable.VariablePattern)
	matches := r.FindAllString(`aaa {{here}} {{ heretoo }}  aaaa`, -1)
	subs := r.FindAllStringSubmatch(`aaa {{here}} {{ heretoo }}  aaaa`, -1)
	log.Println(subs)

	if len(matches) != 2 {
		t.Error("Should be 2 matches")
	}

	//varsMap := make(map[string]interface{})
	//genVarDir := config_variable.NewGenVariable[string](".", 1)
	//genVarFiles := config_variable.NewGenVariable[[]string]([]string{"docs", "README.md"}, 1)
	//
	//varsMap["GIT_DIR"] = genVarDir
	//varsMap[config.SOURCE_FILES_KEY] = genVarFiles
	//
	//output, err := git_env.GitLogHash(varsMap)
	//
	//if err != nil {
	//	t.Error(err)
	//}
	//log.Printf("Output is %s", output)

}
