package git_env

import (
	"errors"
	"fmt"
	process_manager "github.com/ozone2021/ozone/ozone-daemon-lib/process-manager"
	"github.com/ozone2021/ozone/ozone-lib/config/config_keys"
	. "github.com/ozone2021/ozone/ozone-lib/config/config_variable"
	"os"
	"os/exec"
	"strings"
)

func GitLogHash(ordinal int, varsMap VariableMap) (VariableMap, error) {
	dirPath := "./"
	dirPathVar, ok := varsMap["GIT_DIR"]
	if ok {
		dirPath = dirPathVar.ToString()
	}

	filesVar, ok := varsMap[config_keys.SOURCE_FILES_KEY]
	if !ok {
		return nil, errors.New("Source files must be set")
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
		return nil, err
	}

	if len(byteData) == 0 {
		return nil, errors.New(fmt.Sprintf("No git hash, maybe some files don't exist / relative path issue?: %s", filesJoined))
	}

	varsMap["GIT_LOG_HASH"] = NewStringVariable(string(byteData), ordinal)

	return varsMap, nil
}
