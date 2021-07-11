package utils

import (
    "fmt"
    "log"
    "os"
    "strings"
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

func ContextInPattern(context, pattern string) bool {
    patternArray := strings.Split(pattern, "|")
    for _, v := range patternArray {
        if context == v {
            return true
        }
    }
    return false
}

func OSEnvToVarsMap() map[string]string {
    newMap := make(map[string]string)
    for _, kvString := range os.Environ() {
        parts := strings.Split(kvString, "=")
        key, value := parts[0], parts[1]
        newMap[key] = value
    }
    return newMap
}