package env

import (
	"errors"
	"github.com/joho/godotenv"
	"github.com/mitchellh/go-homedir"
	. "github.com/ozone2021/ozone/ozone-lib/config/config_variable"
)

func FromEnvFile(ordinal int, varsMap, fromIncludeMap *VariableMap) error {
	envFile, ok := varsMap.GetVariable("ENV_FILE")

	if !ok {
		return errors.New("ENV_FILE needed.")
	}

	expanded, err := homedir.Expand(envFile.String())

	if err != nil {
		return err
	}

	varsMapStringString, err := godotenv.Read(expanded)

	if err != nil {
		return err
	}

	for k, v := range varsMapStringString {
		fromIncludeMap.AddVariable(NewStringVariable(k, v), ordinal)
	}

	return nil
}
