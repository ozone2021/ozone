package env

import (
	"fmt"
	. "github.com/ozone2021/ozone/ozone-lib/config/config_variable"
	"io/ioutil"
)

func FromVersionFile(ordinal int, varsMap VariableMap) (VariableMap, error) {
	vfileDir, ok := varsMap["VERSION_FILE_DIR"]
	if !ok {
		vfileDir = varsMap["OZONE_WORKING_DIR"]
	}

	versionContents, err := ioutil.ReadFile(fmt.Sprintf("%s/.version", vfileDir))
	if err != nil {
		return nil, fmt.Errorf("Version file error:  #%v . Did you set VERSION_FILE_DIR ?", err)
	}

	outputMap := make(VariableMap)
	outputMap["SERVICE_VERSION"] = NewStringVariable(string(versionContents), ordinal)

	return outputMap, nil
}
