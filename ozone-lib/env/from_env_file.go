package env

import (
	"errors"
	"github.com/joho/godotenv"
	. "github.com/ozone2021/ozone/ozone-lib/config/config_variable"
)

func FromEnvFile(ordinal int, varsMap, fromIncludeMap *VariableMap) error {
	envFile, ok := varsMap.GetVariable("ENV_FILE")

	if !ok {
		return errors.New("ENV_FILE needed.")
	}

	varsMapStringString, err := godotenv.Read(envFile.String())

	if err != nil {
		return err
	}

	for k, v := range varsMapStringString {
		fromIncludeMap.AddVariable(NewStringVariable(k, v), ordinal)
	}

	return nil
}
