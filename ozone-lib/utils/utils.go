package utils

import "fmt"

func ParamsOK(varName string, varsMap map[string]string) error {
    _, ok := varsMap[varName]
    if !ok {
        return fmt.Errorf("Var %s not present for buildPushDocker\n", varName)
    }
    return nil
}
