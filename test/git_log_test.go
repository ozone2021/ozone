package test

import (
	"github.com/ozone2021/ozone/ozone-lib/config/config_keys"
	"github.com/ozone2021/ozone/ozone-lib/config/config_variable"
	"github.com/ozone2021/ozone/ozone-lib/env/git_env"
	"log"
	"testing"
)

func TestGitLog(t *testing.T) {

	varsMap := make(config_variable.VariableMap)
	genVarDir := config_variable.NewStringVariable(".", 1)
	genVarFiles := config_variable.NewSliceVariable([]string{"docs", "README.md"}, 1)

	varsMap["GIT_DIR"] = genVarDir
	varsMap[config_keys.SOURCE_FILES_KEY] = genVarFiles

	output, err := git_env.GitLogHash(varsMap)

	if err != nil {
		t.Error(err)
	}
	log.Printf("Output is %s", output)

}
