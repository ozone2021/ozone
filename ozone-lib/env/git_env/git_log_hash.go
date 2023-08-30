package git_env

import (
	"errors"
	"fmt"
	process_manager "github.com/ozone2021/ozone/ozone-daemon-lib/process-manager"
	. "github.com/ozone2021/ozone/ozone-lib/config/config_variable"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
)

func GitLogHash(ordinal int, varsMap, fromIncludeMap *VariableMap) error {
	dirPath := "./"
	dirPathVar, ok := varsMap.GetVariable("GIT_DIR")
	if ok {
		ozoneWorkingDir, _ := varsMap.GetVariable("OZONE_WORKING_DIR")
		dirPath = path.Join(ozoneWorkingDir.String(), dirPathVar.String())
	}

	filesVar, ok := varsMap.GetVariable("GIT_SOURCE_FILES")
	if !ok {
		return errors.New("Source files must be set")
	}

	filesJoined := strings.Join(filesVar.GetSliceValue(), " ")

	cmdString := fmt.Sprintf("git log -n 1 --pretty=format:%%H --  %s",
		filesJoined,
	)

	cmdFields, argFields := process_manager.CommandFromFields(cmdString)
	cmd := exec.Command(cmdFields[0], argFields...)
	cmd.Dir = dirPath
	cmd.Stderr = os.Stderr
	byteData, err := cmd.Output()
	if err != nil {
		log.Println("GitLogHash")
		return err
	}

	if len(byteData) == 0 {
		log.Printf("Command: %s", cmdString)
		return errors.New(fmt.Sprintf("No git hash, maybe some files don't exist / relative path issue?: %s", filesJoined))
	}

	fromIncludeMap.AddVariable(NewStringVariable("GIT_LOG_HASH", string(byteData)), ordinal)

	return nil
}
