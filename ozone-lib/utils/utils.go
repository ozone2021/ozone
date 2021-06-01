package utils

import "fmt"

func ParamsOK(varName string, varsMap map[string]string) error {
    param, ok := varsMap[varName]
    if !ok || param == "" {
        return fmt.Errorf("Var %s not present for buildPushDocker\n", varName)
    }
    return nil
}
