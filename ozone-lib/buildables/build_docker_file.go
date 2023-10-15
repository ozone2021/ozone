package buildables

import (
	"fmt"
	process_manager "github.com/ozone2021/ozone/ozone-daemon-lib/process-manager"
	. "github.com/ozone2021/ozone/ozone-lib/config/config_variable"
	"github.com/ozone2021/ozone/ozone-lib/logger_lib"
	"github.com/ozone2021/ozone/ozone-lib/utils"
	"os/exec"
)

func getParams() []string {
	return []string{
		//"DIR",
		"DOCKER_FULL_TAG",
		"SERVICE",
		//"DOCKER_BUILD_DIR",
		//"GITLAB_PROJECT_CODE",
		//"BUILD_ARGS",
	}
}

func BuildDockerContainer(varsMap *VariableMap, logger *logger_lib.Logger) error {
	for _, arg := range getParams() {
		if err := utils.ParamsOK("BuildDockerContainer", arg, varsMap); err != nil {
			return err
		}
	}

	dockerBuildDir, ok := varsMap.GetVariable("DOCKER_BUILD_DIR")
	if !ok {
		dockerBuildDir, _ = varsMap.GetVariable("OZONE_WORKING_DIR")
	}

	//sourceDirArg := varsMap.GetVariable("DIR").Fstring("--build-arg DIR=%s")

	cmdCallDir, _ := varsMap.GetVariable("OZONE_WORKING_DIR")
	tag, _ := varsMap.GetVariable("DOCKER_FULL_TAG")

	buildArgs := ""
	buildArgsVar, ok := varsMap.GetVariable("DOCKER_BUILD_ARGS")
	if ok {
		buildArgs = fmt.Sprintf("%s", buildArgsVar)
	}

	dockerfilePath := ""
	dockerfilePathVar, ok := varsMap.GetVariable("DOCKERFILE")
	if ok {
		buildArgs = fmt.Sprintf("%s -f %s", buildArgs, dockerfilePathVar)
	}

	cmdString := fmt.Sprintf("docker build -t %s %s %s %s",
		tag,
		buildArgs,
		dockerBuildDir,
		dockerfilePath,
	)

	logger.Printf("Build cmd is: %s \n", cmdString)

	cmdFields, argFields := process_manager.CommandFromFields(cmdString)
	cmd := exec.Command(cmdFields[0], argFields...)
	cmd.Dir = cmdCallDir.String()

	cmd.Stdout = logger.File
	cmd.Stderr = logger.File
	err := cmd.Run()
	if err != nil {
		logger.Println("build docker err for image %s ", tag)
		return err
	}
	cmd.Wait()

	//query := &process_manager.ProcessCreateQuery{
	//	serviceName,
	//	"/",
	//	ozoneWorkingDir,
	//	cmdString,
	//	true,
	//	false,
	//	varsMap,
	//}
	//
	//process_manager_client.AddProcess(query)

	return nil
}
