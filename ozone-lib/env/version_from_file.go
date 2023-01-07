package env

import (
	"fmt"
	. "github.com/ozone2021/ozone/ozone-lib/config/config_variable"
	"io/ioutil"
)

func FromVersionFile(ordinal int, varsMap, fromIncludeMap *VariableMap) error {
	vfileDir, ok := varsMap.GetVariable("VERSION_FILE_DIR")
	if !ok {
		vfileDir, _ = varsMap.GetVariable("OZONE_WORKING_DIR")
	}

	versionContents, err := ioutil.ReadFile(fmt.Sprintf("%s/.version", vfileDir))
	if err != nil {
		return fmt.Errorf("Version file error:  #%v . Did you set VERSION_FILE_DIR ?", err)
	}

	fromIncludeMap.AddVariable(NewStringVariable("SERVICE_VERSION", string(versionContents)), ordinal)

	return nil
}
