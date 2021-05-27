package docker

import (
	"fmt"
	"log"
	"net/rpc"
	"os"
	process_manager "ozone-daemon-lib/process-manager"
	"ozone-lib/utils"
)

func getDockerRunParams() []string {
	return []string{
		"FULL_TAG",
		//"BUILD_ARGS",
	}
}

func VarsMapToDockerEnvString(varsMap map[string]string) string {
	envString := ""
	for key, value := range varsMap {
		envString = fmt.Sprintf("%s-e %s=%s ", envString, key, value)
	}
	return envString
}

func Build(serviceName string, env map[string]string) error {
	for _, arg := range getDockerRunParams() {
		if err := utils.ParamsOK(arg, env); err != nil {
			return err
		}
	}

	ozoneWorkingDir, err := os.Getwd()
	if err != nil {
		log.Println(err)
	}

	containerImage := env["FULL_TAG"]
	envString := VarsMapToDockerEnvString(env)

	cmdString := fmt.Sprintf("docker run --rm --user root --network host -d -p 5432:5432 %s %s",
		envString,
		containerImage,
	)

	query := &process_manager.ProcessCreateQuery{
		serviceName,
		ozoneWorkingDir,
		ozoneWorkingDir,
		cmdString,
		true,
		env,
	}

	client, err := rpc.DialHTTP("tcp", ":8000")
	if err != nil {
		log.Fatal("dialing:", err)
		return err
	}
	err = client.Call("ProcessManager.AddProcess", query, nil)
	if err != nil {
		log.Fatal("arith error:", err)
		return err
	}
	return nil
}
