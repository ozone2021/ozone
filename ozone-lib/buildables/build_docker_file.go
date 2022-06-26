package buildables

import (
	"fmt"
	process_manager "github.com/ozone2021/ozone/ozone-daemon-lib/process-manager"
	. "github.com/ozone2021/ozone/ozone-lib/config/config_variable"
	"github.com/ozone2021/ozone/ozone-lib/utils"
	"log"
	"os"
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

func BuildDockerContainer(varsMap VariableMap) error {
	for _, arg := range getParams() {
		if err := utils.ParamsOK("BuildDockerContainer", arg, varsMap); err != nil {
			return err
		}
	}

	dockerBuildDir, ok := varsMap["DOCKER_BUILD_DIR"]
	if !ok {
		dockerBuildDir = varsMap["OZONE_WORKING_DIR"]
	}
	sourceDirArg := varsMap["DIR"].Fstring("--build-arg DIR=%s")

	cmdCallDir, _ := varsMap["OZONE_WORKING_DIR"]
	tag, _ := varsMap["DOCKER_FULL_TAG"]

	buildArgs := varsMap["DOCKER_BUILD_ARGS"].ToString()
	buildArgs = fmt.Sprintf("%s %s", buildArgs, sourceDirArg)

	dockerfilePath := varsMap["DOCKERFILE"].Fstring("-f %s")

	cmdString := fmt.Sprintf("docker build -t %s %s %s %s",
		tag,
		buildArgs,
		dockerBuildDir,
		dockerfilePath,
	)

	log.Printf("Build cmd is: %s", cmdString)

	cmdFields, argFields := process_manager.CommandFromFields(cmdString)
	cmd := exec.Command(cmdFields[0], argFields...)
	cmd.Dir = cmdCallDir.ToString()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stdout
	err := cmd.Run()
	if err != nil {
		fmt.Println("build docker err")
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
