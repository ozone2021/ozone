package env

import (
	"github.com/joho/godotenv"
	. "github.com/ozone2021/ozone/ozone-lib/config/config_variable"
)

func mapToVariableMap(ordinal int, envFile ...string) (VariableMap, error) {
	varsMapStringString, err := godotenv.Read(envFile...)

	if err != nil {
		return nil, err
	}

	varsMap := make(VariableMap)
	for k, v := range varsMapStringString {
		varsMap[k] = NewGenVariable(v, ordinal)
	}

	return varsMap, nil
}

func FromEnvFile(ordinal int, varsParamMap VariableMap) (VariableMap, error) {
	envFile, genErr := GenVarToString(varsParamMap, "ENV_FILE")

	if genErr != nil {
		return mapToVariableMap(ordinal)
	}
	return mapToVariableMap(ordinal, envFile)
}
