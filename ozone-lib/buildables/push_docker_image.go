package buildables

import (
	"fmt"
	"github.com/kballard/go-shellquote"
	. "github.com/ozone2021/ozone/ozone-lib/config/config_variable"
	"github.com/ozone2021/ozone/ozone-lib/logger_lib"
	"github.com/ozone2021/ozone/ozone-lib/utils"
	"os/exec"
)

func getPushDockerImageParams() []string {
	return []string{
		"DOCKER_FULL_TAG",
		"SERVICE",
	}
}

func PushDockerImage(varsMap *VariableMap, logger *logger_lib.Logger) error {
	for _, arg := range getPushDockerImageParams() {
		if err := utils.ParamsOK("PushDockerImage", arg, varsMap); err != nil {
			return err
		}
	}

	tag, _ := varsMap.GetVariable("DOCKER_FULL_TAG")
	cmdString := fmt.Sprintf("docker push %s", tag)
	fields, err := shellquote.Split(cmdString)
	if err != nil {
		return fmt.Errorf("Error parsing command push_docker_image.go: %s", err.Error())
	}
	cmd := exec.Command(fields[0], fields[1:]...)
	cmd.Stdout = logger.File
	cmd.Stderr = logger.File
	err = cmd.Run()
	if err != nil {
		logger.Fatalln(fmt.Sprintf("Error in push_docker_image.go: %s", err.Error()))
		return err
	}

	return nil
}
