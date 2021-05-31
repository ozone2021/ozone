package buildables

import (
	"fmt"
	process_manager "github.com/JamesArthurHolland/ozone/ozone-daemon-lib/process-manager"
	"github.com/JamesArthurHolland/ozone/ozone-lib/utils"
	"log"
	"net/rpc"
	"os"
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
	serviceName := varsMap["SERVICE"]

	ozoneWorkingDir, err := os.Getwd()
	if err != nil {
		log.Println(err)
	}
	cmdString := fmt.Sprintf("docker push %s",
		tag,
	)

	query := &process_manager.ProcessCreateQuery{
		serviceName,
		ozoneWorkingDir,
		ozoneWorkingDir,
		cmdString,
		true,
		false,
		varsMap,
	}

	client, err := rpc.DialHTTP("tcp", ":8000")
	if err != nil {
		log.Fatal("dialing:", err)
	}
	err = client.Call("ProcessManager.AddProcess", query, nil)
	if err != nil {
		log.Fatal("arith error:", err)
	}

	return nil
}
