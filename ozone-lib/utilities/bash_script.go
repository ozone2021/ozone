package utilities

import (
	"fmt"
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

func RunBashScript(env map[string]string) error {
	for _, arg := range getParams() {
		if err := utils.ParamsOK("RunBashScript", arg, env); err != nil {
			return err
		}
	}
	scriptPath := env["SCRIPT"]
	cmd := exec.Command("/bin/bash", scriptPath)

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
