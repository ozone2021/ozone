package utilities

import (
	"fmt"
	"github.com/ozone2021/ozone/ozone-lib/config/config_variable"
	"github.com/ozone2021/ozone/ozone-lib/utils"
	"os"
	"os/exec"
)

func getParams() []string {
	return []string{
		"SCRIPT",
		//"WORKING_DIRECTORY" optional
	}
}

func RunBashScript(envVarMap config_variable.VariableMap) error {
	for _, arg := range getParams() {
		if err := utils.ParamsOK("RunBashScript", arg, envVarMap); err != nil {
			return err
		}
	}
	scriptPath := envVarMap["SCRIPT"].ToString()
	cmd := exec.Command("/bin/bash", scriptPath)

	env := config_variable.ConvertMap(envVarMap)

	workingDir, ok := env["WORKING_DIR"]
	if ok {
		cmd.Dir = fmt.Sprintf("%s/%s", env["OZONE_WORKING_DIR"], workingDir)
	}

	cmd.Env = os.Environ()
	for k, v := range env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()

	if err != nil {
		return err
	}
	//output := string(out)
	return nil
}
