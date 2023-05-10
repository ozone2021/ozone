package buildables

import (
	"fmt"
	process_manager "github.com/ozone2021/ozone/ozone-daemon-lib/process-manager"
	. "github.com/ozone2021/ozone/ozone-lib/config/config_variable"
	"github.com/ozone2021/ozone/ozone-lib/utils"
	"os"
	"os/exec"
)

func getTagDockerImageAsParams() []string {
	return []string{
		"SOURCE_TAG",
		"TARGET_TAG",
	}
}

func TagDockerImageAs(varsMap *VariableMap) error {
	for _, arg := range getTagDockerImageAsParams() {
		if err := utils.ParamsOK("TagDockerImageAs", arg, varsMap); err != nil {
			return err
		}
	}

	sourceTag, _ := varsMap.GetVariable("SOURCE_TAG")
	targetTag, _ := varsMap.GetVariable("TARGET_TAG")
	cmdString := fmt.Sprintf("docker tag %s %s", sourceTag, targetTag)
	cmdFields, argFields := process_manager.CommandFromFields(cmdString)
	cmd := exec.Command(cmdFields[0], argFields...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stdout
	err := cmd.Run()
	if err != nil {
		fmt.Println("docker TagDockerImageAs err")
		return err
	}
	cmd.Wait()

	return nil
}
