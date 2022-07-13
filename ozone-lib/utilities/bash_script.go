package utilities

import (
	"fmt"
	"github.com/ozone2021/ozone/ozone-lib/config/config_variable"
	"os"
	"os/exec"
)

// -1 is an error which can't be ignored in conditional
//
func RunBashScript(script string, envVarMap *config_variable.VariableMap) (int, error) {
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

	cmd.Env = combinedEnv
	for k, v := range env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			return exitError.ExitCode(), err
		}
	}

	//output := string(out)
	return 0, nil
}
