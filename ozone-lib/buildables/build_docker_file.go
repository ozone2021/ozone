package buildables

import (
	"fmt"
	process_manager "github.com/JamesArthurHolland/ozone/ozone-daemon-lib/process-manager"
	"github.com/JamesArthurHolland/ozone/ozone-lib/utils"
	"log"
	"os"
	"os/exec"
)


func getParams() []string {
	return []string{
		"DIR",
		"FULL_TAG",
		"SERVICE",
		//"GITLAB_PROJECT_CODE",
		//"BUILD_ARGS",
	}
}


func BuildPushDockerContainer(varsMap map[string]string) error {
	for _, arg := range getParams() {
		if err := utils.ParamsOK(arg, varsMap); err != nil {
			return err
		}
	}

	dockerBuildDir := varsMap["DIR"]
	tag := varsMap["FULL_TAG"]

	buildArgs, ok := varsMap["BUILD_ARGS"]
	if !ok {
		buildArgs = ""
	}

	dockerfilePath, ok := varsMap["DOCKERFILE"]
	if ok {
		dockerfilePath = fmt.Sprintf("-f %s", dockerfilePath)
	} else {
		dockerfilePath = ""
	}

	cmdString := fmt.Sprintf("docker build -t %s %s %s %s",
		tag,
		buildArgs,
		dockerBuildDir,
		dockerfilePath,
	)

	log.Printf("Build cmd is: %s", cmdString)

	cmdFields, argFields := process_manager.CommandFromFields(cmdString)
	cmd := exec.Command(cmdFields[0], argFields...)
	cmd.Dir = dockerBuildDir
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
