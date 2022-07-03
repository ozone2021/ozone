package test

import (
	"github.com/ozone2021/ozone/ozone-lib/config/config_variable"
	"github.com/ozone2021/ozone/ozone-lib/env"
	"testing"
)

func TestVariableMap(t *testing.T) {
	varsMap := config_variable.NewVariableMap()

	varsMap.AddVariable(config_variable.NewStringVariable("./version", "1"), 1)

	outputMap := config_variable.NewVariableMap()
	err := env.FromVersionFile(1, varsMap, outputMap)

	if err != nil {
		t.Error(err)
	}
}
