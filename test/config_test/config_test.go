package config_test

import (
	ozoneConfig "github.com/ozone2021/ozone/ozone-lib/config"
	"github.com/ozone2021/ozone/ozone-lib/config/config_variable"
	"log"
	"testing"
)

func TestWhitespaceRender(t *testing.T) {
	config := ozoneConfig.ReadConfig()

	topLevelScope := config_variable.CopyOrCreateNew(config.BuildVars)

	vm, err := config.FetchEnvs(1, []string{"prepend-123-all-services"}, topLevelScope)
	if err != nil {
		t.Error(err)
	}
	log.Println(vm)
	log.Println(config)

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
