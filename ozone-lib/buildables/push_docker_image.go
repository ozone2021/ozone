package buildables

import (
	"fmt"
	process_manager "github.com/ozone2021/ozone/ozone-daemon-lib/process-manager"
	. "github.com/ozone2021/ozone/ozone-lib/config/config_variable"
	"github.com/ozone2021/ozone/ozone-lib/utils"
	"os"
	"os/exec"
)

func getPushDockerImageParams() []string {
	return []string{
		"DOCKER_FULL_TAG",
		"SERVICE",
	}
}

func PushDockerImage(varsMap VariableMap) error {
	for _, arg := range getPushDockerImageParams() {
		if err := utils.ParamsOK("PushDockerImage", arg, varsMap); err != nil {
			return err
		}
	}

	cmdString, _ := GenVarToFstring(varsMap, "DOCKER_FULL_TAG", "docker push %s")
	cmdFields, argFields := process_manager.CommandFromFields(cmdString)
	cmd := exec.Command(cmdFields[0], argFields...)
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
	//	ozoneWorkingDir,
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
