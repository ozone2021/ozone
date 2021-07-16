package utils

import (
    "fmt"
    "log"
)

func ParamsOK(runnable string, varName string, varsMap map[string]string) error {
    param, ok := varsMap[varName]
    if !ok || param == "" {
        return fmt.Errorf("Var %s not present for %s \n", varName, runnable)
    }
    return nil
}

func WarnIfNullVar(service, varValue, varName string) {
    if varValue == "" {
        log.Printf("WARNING: Should %s be nil for helm, service: %s \n", varName, service)
    }
}

