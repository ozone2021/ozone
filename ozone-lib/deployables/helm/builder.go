package helm

import (
	"fmt"
	process_manager "github.com/ozone2021/ozone/ozone-daemon-lib/process-manager"
	"github.com/ozone2021/ozone/ozone-lib/config/config_variable"
	"github.com/ozone2021/ozone/ozone-lib/utils"
	"log"
	"os"
	"os/exec"
)

func getHelmParams() []string {
	return []string{
		"INSTALL_NAME",
		"CHART_DIR",
		//"GITLAB_PROJECT_CODE",
		//"BUILD_ARGS",
	}
}

func Deploy(serviceName string, envVarMap config_variable.VariableMap) error {
	for _, arg := range getHelmParams() {
		if err := utils.ParamsOK("helmChart", arg, envVarMap); err != nil {
			return err
		}
	}

	ozoneWorkingDir, err := os.Getwd()
	if err != nil {
		log.Println(err)
	}

	env := config_variable.ConvertMap(envVarMap)

	installName := env["INSTALL_NAME"]
	chartDir := env["CHART_DIR"]

	args, ok := env["HELM_ARGS"]
	if !ok {
		args = ""
	}

	valuesFile, ok := env["VALUES_FILE"]
	if ok {
		valuesFile = fmt.Sprintf("-f %s", valuesFile)
	} else {
		valuesFile = ""
	}

	cmdString := fmt.Sprintf("helm upgrade -i %s %s %s %s",
		installName,
		valuesFile,
		args,
		chartDir,
	)

	log.Printf("Helm cmd is: %s", cmdString)

	cmdFields, argFields := process_manager.CommandFromFields(cmdString)
	cmd := exec.Command(cmdFields[0], argFields...)
	cmd.Dir = ozoneWorkingDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stdout
	if err := cmd.Run(); err != nil {
		fmt.Println("build docker err")
		return err
	}
	cmd.Wait()
	return nil
}
