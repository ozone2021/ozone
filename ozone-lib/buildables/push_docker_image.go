package buildables

import (
	"fmt"
	process_manager "github.com/JamesArthurHolland/ozone/ozone-daemon-lib/process-manager"
	"github.com/JamesArthurHolland/ozone/ozone-lib/utils"
	"os"
	"os/exec"
)


func getPushDockerImageParams() []string {
	return []string{
		"FULL_TAG",
		"SERVICE",
	}
}


func PushDockerImage(varsMap map[string]string) error {
	for _, arg := range getPushDockerImageParams() {
		if err := utils.ParamsOK(arg, varsMap); err != nil {
			return err
		}
	}

	tag := varsMap["FULL_TAG"]
	cmdString := fmt.Sprintf("docker push %s",
		tag,
	)
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
