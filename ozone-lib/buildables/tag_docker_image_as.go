package buildables

import (
	"errors"
	"fmt"
	process_manager "github.com/ozone2021/ozone/ozone-daemon-lib/process-manager"
	. "github.com/ozone2021/ozone/ozone-lib/config/config_variable"
	"github.com/ozone2021/ozone/ozone-lib/logger_lib"
	"github.com/ozone2021/ozone/ozone-lib/utils"
	"os"
	"os/exec"
	"strings"
)

func getTagDockerImageAsParams() []string {
	return []string{
		"SOURCE_TAG",
		"TARGET_TAG",
	}
}

func TagDockerImageAs(varsMap *VariableMap, logger *logger_lib.Logger) error {
	for _, arg := range getTagDockerImageAsParams() {
		if err := utils.ParamsOK("TagDockerImageAs", arg, varsMap); err != nil {
			return err
		}
	}

	sourceTag, _ := varsMap.GetVariable("SOURCE_TAG")
	targetTag, _ := varsMap.GetVariable("TARGET_TAG")

	if strings.Contains(targetTag.String(), "Merged branch is not release branch.") {
		return errors.New("Merged branch is not release branch.")
	}
	logger.Printf("Pulling docker image %s \n", sourceTag)
	pullCmdString := fmt.Sprintf("docker pull %s", sourceTag)
	cmdFields, argFields := process_manager.CommandFromFields(pullCmdString)
	pullCmd := exec.Command(cmdFields[0], argFields...)
	pullCmd.Stdout = os.Stdout
	pullCmd.Stderr = os.Stdout
	err := pullCmd.Run()
	if err != nil {
		fmt.Printf("%s\n", pullCmdString)
		return err
	}
	pullCmd.Wait()

	logger.Printf("Tagging docker image %s as %s \n", sourceTag, targetTag)
	tagCmdString := fmt.Sprintf("docker tag %s %s", sourceTag, targetTag)
	tagCmdFields, argFields := process_manager.CommandFromFields(tagCmdString)
	tagCmd := exec.Command(tagCmdFields[0], argFields...)
	tagCmd.Stdout = os.Stdout
	tagCmd.Stderr = os.Stdout
	err = tagCmd.Run()
	if err != nil {
		fmt.Printf("%s\n", tagCmdString)
		return err
	}
	tagCmd.Wait()

	return nil
}
