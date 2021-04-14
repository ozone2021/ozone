package buildables

import (
	"fmt"
	"log"
	"net/rpc"
	"os"
	process_manager "ozone-daemon-lib/process-manager"
	"ozone-lib/config"
)


func getParams() []string {
	return []string{
		"DOCKERFILE_DIR",
		"FULL_TAG",
		"SERVICE",
		//"GITLAB_PROJECT_CODE",
		//"BUILD_ARGS",
	}
}


func BuildPushDockerContainer(varsMap map[string]string) error {
	for _, arg := range getParams() {
		if err := config.ParamsOK(arg, varsMap); err != nil {
			return err
		}
	}

	dockerFileDir := varsMap["DOCKERFILE_DIR"]
	serviceName := varsMap["SERVICE"]
	tag := varsMap["FULL_TAG"]

	buildArgs, ok := varsMap["BUILD_ARGS"]
	if !ok {
		buildArgs = ""
	}

	ozoneWorkingDir, err := os.Getwd()
	if err != nil {
		log.Println(err)
	}
	cmdString := fmt.Sprintf("docker build %s -t %s %s && docker push %s",
		buildArgs,
		tag,
		dockerFileDir,
		tag,
	)

	query := &process_manager.ProcessCreateQuery{
		serviceName,
		ozoneWorkingDir,
		ozoneWorkingDir,
		cmdString,
		true,
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
}
