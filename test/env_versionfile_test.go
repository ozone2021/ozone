package test

import (
	"github.com/ozone2021/ozone/ozone-lib/config/config_variable"
	"github.com/ozone2021/ozone/ozone-lib/env"
	"testing"
)

func TestVersionFile(t *testing.T) {
	varsMap := make(config_variable.VariableMap)

	varsMap["VERSION_FILE_DIR"] = config_variable.NewStringVariable("./version", 1)

	varsMap, err := env.FromVersionFile(1, varsMap)

	if err != nil {
		t.Error(err)
	}

	version, ok := varsMap["SERVICE_VERSION"]
	if !ok {
		t.Error("SERVICE_VERSION should be set.")
	}

	if version.String() != "1.2" {
		t.Error("Version should be 1.2")
	}
}
