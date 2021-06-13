package utils

import "fmt"

func ParamsOK(runnable string, varName string, varsMap map[string]string) error {
    param, ok := varsMap[varName]
    if !ok || param == "" {
        return fmt.Errorf("Var %s not present for %s \n", varName, runnable)
    }
    return nil
}
