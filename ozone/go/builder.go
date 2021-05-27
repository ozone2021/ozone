package _go

import (
	"fmt"
	"log"
	"os"
	"ozone-daemon-lib/process-manager"
	process_manager_client "ozone-daemon-lib/process-manager-client"
)



func Build(serviceName string, relativeDir string, file string, varsMap map[string]string) error {
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
		varsMap,
	}

	if err := process_manager_client.AddProcess(query); err != nil{
		return err
	}
	return nil
}