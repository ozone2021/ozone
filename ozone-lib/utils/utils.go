package utils

import (
	"fmt"
	"github.com/ozone2021/ozone/ozone-lib/config/config_variable"
	"log"
)

func ParamsOK(runnable string, varName string, varsMap *config_variable.VariableMap) error {
	param, ok := varsMap.GetVariable(varName)
	if !ok || param.String() == "" { // TODO cast
		return fmt.Errorf("Var %s not present for %s \n", varName, runnable)
	}
	return nil
}

func WarnIfNullVar(service, varValue, varName string) {
	if varValue == "" {
		log.Printf("WARNING: Should %s be nil for helm, service: %s \n", varName, service)
	}
}
