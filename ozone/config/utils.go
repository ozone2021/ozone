package config

import "fmt"

func VarsToMap(vars []*Var) map[string]string {
	result := make(map[string]string)

	for _, v := range vars {
		result[v.Name] = v.Value
	}

	return result
}

func ParamsOK(varName string, varsMap map[string]string) error {
	_, ok := varsMap[varName]
	if !ok {
		return fmt.Errorf("Var %s not present for buildPushDocker\n", varName)
	}
	return nil
}