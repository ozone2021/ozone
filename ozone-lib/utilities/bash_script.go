package utilities

import (
	"fmt"
	"github.com/ozone2021/ozone/ozone-lib/config/config_variable"
	"github.com/ozone2021/ozone/ozone-lib/logger_lib"
	"os"
	"os/exec"
)

// -1 is an error which can't be ignored in conditional
func RunBashScript(script string, envVarMap *config_variable.VariableMap, logger *logger_lib.Logger) (int, error) {
	cmd := exec.Command("/bin/bash", script)

	env := envVarMap.ConvertMap()

	workingDir, ok := env["WORKING_DIR"]
	if !ok {
		cmd.Dir = fmt.Sprintf("%s/%s", env["OZONE_WORKING_DIR"], workingDir)
	}

	combinedEnv := os.Environ()

	for key, value := range env {
		combinedEnv = append(combinedEnv, fmt.Sprintf("%s=%s", key, value)) // TODO this doesn't overwrite os.ENv vars
	}

	useEnv, ok := envVarMap.GetVariable("USE_ENV")
	if !ok || ok && useEnv.String() != "false" {
		cmd.Env = combinedEnv
	}
	cmd.Stdout = logger.File
	cmd.Stderr = logger.File
	if err := cmd.Run(); err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			return exitError.ExitCode(), err
		}
	}

	//output := string(out)
	return 0, nil
}
