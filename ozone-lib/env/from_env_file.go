package env

import (
	"errors"
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
		varsMap[k] = NewStringVariable(v, ordinal)
	}

	return varsMap, nil
}

func FromEnvFile(ordinal int, varsParamMap VariableMap) (VariableMap, error) {
	envFile, ok := varsParamMap["ENV_FILE"]

	if !ok {
		return nil, errors.New("ENV_FILE needed.")
	}

	return mapToVariableMap(ordinal, envFile.String())
}
