package helm

import (
	"fmt"
	"github.com/kballard/go-shellquote"
	"github.com/ozone2021/ozone/ozone-lib/config/config_variable"
	"github.com/ozone2021/ozone/ozone-lib/logger_lib"
	"github.com/ozone2021/ozone/ozone-lib/utils"
	"os"
	"os/exec"
	"strings"
)

func getHelmParams() []string {
	return []string{
		"INSTALL_NAME",
		"HELM_CHART",
		//"GITLAB_PROJECT_CODE",
		//"BUILD_ARGS",
	}
}

func Deploy(envVarMap *config_variable.VariableMap, logger *logger_lib.Logger) error {
	for _, arg := range getHelmParams() {
		if err := utils.ParamsOK("helmChart", arg, envVarMap); err != nil {
			return err
		}
	}

	ozoneWorkingDir, err := os.Getwd()
	if err != nil {
		logger.Println(err)
	}

	env := envVarMap.ConvertMapPongo()

	installName := env["INSTALL_NAME"]
	helmChart := env["HELM_CHART"]

	argsString := ""
	argsVar, ok := envVarMap.GetVariable("HELM_ARGS")
	if ok {
		argsStringSlice := argsVar.GetSliceValue()
		argsString = strings.Join(argsStringSlice, " ")
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
		argsString,
		helmChart,
	)

	logger.Printf("Helm cmd is: %s", cmdString)

	fields, err := shellquote.Split(cmdString)
	if err != nil {
		return fmt.Errorf("Error parsing command builder.go: %s", err.Error())
	}
	cmd := exec.Command(fields[0], fields[1:]...)
	cmd.Dir = ozoneWorkingDir
	cmd.Stdout = logger.File
	cmd.Stderr = logger.File
	if err := cmd.Run(); err != nil {
		fmt.Println("build docker err")
		return err
	}
	cmd.Wait()
	return nil
}
