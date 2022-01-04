package env

import (
	"github.com/joho/godotenv"
)

func FromEnvFile(varsParamMap map[string]string) (map[string]string, error) {
	envFile, ok := varsParamMap["ENV_FILE"]
	var varsMap map[string]string
	var err error
	if ok {
		varsMap, err = godotenv.Read(envFile)
	} else {
		varsMap, err = godotenv.Read()
	}
	if err != nil {
		return nil, err
	}
	return varsMap, nil
}