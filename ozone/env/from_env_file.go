package env

import (
	"fmt"
	"github.com/joho/godotenv"
)

func FromEnvFile(varsParamMap map[string]string) (map[string]string, error) {
	envFile, ok := varsParamMap["ENV_FILE"]
	var varsMap map[string]string
	var err error
	if ok && envFile != "" {
		varsMap, err = godotenv.Read()
		if err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("fromEnvFile needs ENV_FILE")
	}
	return varsMap, nil
}