package test

import (
	"github.com/ozone2021/ozone/ozone-lib/config/config_variable"
	"testing"
)

func TestVersionFile(t *testing.T) {
	varsMap := config_variable.NewVariableMap()

	varsMap.AddVariable(config_variable.NewStringVariable("DOCKER_TAG", "the_docker_tag"), 1)

	asOutput := make(map[string]string)
	asOutput["DOCKER_TAG"] = "DOCKER_TAG_BASE"
	outputMap, err := varsMap.AsOutput(asOutput)
	if err != nil {
		t.Error(err)
	}

	value, exists := outputMap.GetVariable("DOCKER_TAG_BASE")
	if !exists || value.String() != "the_docker_tag" {
		t.Error("Value should exist and equal the_docker_tag")
	}
	ordinal, err := outputMap.GetOrdinal("DOCKER_TAG_BASE")
	if err != nil {
		t.Error(err)
	}
	if ordinal != 1 {
		t.Error("Ordinal should equal 1")
	}
}
