package _go

import (
	"fmt"
	"github.com/ozone2021/ozone/ozone-daemon-lib/process-manager"
	process_manager_client "github.com/ozone2021/ozone/ozone-daemon-lib/process-manager-client"
	"github.com/ozone2021/ozone/ozone-lib/config/config_variable"
	"log"
	"os"
)

func Build(serviceName string, relativeDir string, file string, varsMap config_variable.VariableMap) error {
	ozoneWorkingDir, err := os.Getwd()
	if err != nil {
		log.Println(err)
	}
	cmdString := fmt.Sprintf("go build -o __OUTPUT__/executable %s/%s",
		relativeDir,
		file)

	query := &process_manager.ProcessCreateQuery{
		serviceName,
		ozoneWorkingDir,
		ozoneWorkingDir,
		cmdString,
		true,
		false,
		varsMap,
	}

	if err := process_manager_client.AddProcess(query); err != nil {
		return err
	}
	return nil
}
