package git_env

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	process_manager "github.com/ozone2021/ozone/ozone-daemon-lib/process-manager"
	. "github.com/ozone2021/ozone/ozone-lib/config/config_variable"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
)

func GitLogHash(ordinal int, varsMap, fromIncludeMap *VariableMap) error {
	ozoneWorkingDir, _ := varsMap.GetVariable("OZONE_WORKING_DIR")

	filesVar, ok := varsMap.GetVariable("GIT_SOURCE_FILES")
	if !ok {
		return errors.New("Source files must be set")
	}

	var hashBytes []byte
	for _, filePath := range filesVar.GetSliceValue() {
		absFilePath := path.Join(ozoneWorkingDir.String(), filePath)
		folderPath := filepath.Dir(absFilePath)
		fileHashBytes, err := gitLogHashOfFile(absFilePath, folderPath)
		if err != nil {
			return err
		}
		hashBytes = append(hashBytes, fileHashBytes...)
	}

	hash := md5.Sum(hashBytes)
	hashString := hex.EncodeToString(hash[:])

	fromIncludeMap.AddVariable(NewStringVariable("GIT_LOG_HASH", string(hashString)), ordinal)

	return nil
}

func gitLogHashOfFile(filePath string, dirPath string) ([]byte, error) {
	cmdString := fmt.Sprintf("git log -n 1 --pretty=format:%%H --  %s",
		filePath,
	)

	cmdFields, argFields := process_manager.CommandFromFields(cmdString)
	cmd := exec.Command(cmdFields[0], argFields...)
	cmd.Dir = dirPath
	cmd.Stderr = os.Stderr
	byteData, err := cmd.Output()
	if err != nil {
		log.Println("GitLogHash")
		return nil, err
	}
	if len(byteData) == 0 {
		log.Printf("Command: %s", cmdString)
		return nil, errors.New(fmt.Sprintf("No git hash, maybe some files don't exist / relative path issue?: %s", filePath))
	}
	return byteData, nil
}
