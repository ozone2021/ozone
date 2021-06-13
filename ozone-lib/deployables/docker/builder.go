package docker

import (
	"fmt"
	process_manager "github.com/JamesArthurHolland/ozone/ozone-daemon-lib/process-manager"
	process_manager_client "github.com/JamesArthurHolland/ozone/ozone-daemon-lib/process-manager-client"
	"github.com/JamesArthurHolland/ozone/ozone-lib/utils"
	"log"
	"os"
)

func getDockerRunParams() []string {
	return []string{
		"FULL_TAG",
		"PORT",
		"NETWORK",
		"SERVICE",
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

func CreateNetworkIfNotExists(serviceName string, env map[string]string) error {
	network := env["NETWORK"]

	cmdString := fmt.Sprintf("docker network create -d bridge %s",
		network,
	)

	ozoneWorkingDir, err := os.Getwd()
	if err != nil {
		log.Println(err)
	}

	query := &process_manager.ProcessCreateQuery{
		serviceName,
		"/",
		ozoneWorkingDir,
		cmdString,
		true,
		true,
		env,
	}

	process_manager_client.AddProcess(query)

	return nil
}

func DeleteContainerIfExists(serviceName string, env map[string]string) error {
	cmdString := fmt.Sprintf("docker rm -f %s",
		serviceName,
	)

	ozoneWorkingDir, err := os.Getwd()
	if err != nil {
		log.Println(err)
	}

	query := &process_manager.ProcessCreateQuery{
		serviceName,
		"/",
		ozoneWorkingDir,
		cmdString,
		true,
		true,
		env,
	}

	process_manager_client.AddProcess(query)
	return nil
}



func Build(env map[string]string) error {
	for _, arg := range getDockerRunParams() {
		if err := utils.ParamsOK("DeployDocker", arg, env); err != nil {
			return err
		}
	}

	ozoneWorkingDir, err := os.Getwd()
	if err != nil {
		log.Println(err)
	}

	serviceName := env["SERVICE"]
	CreateNetworkIfNotExists(serviceName, env)
	DeleteContainerIfExists(serviceName, env)

	containerImage := env["FULL_TAG"]
	network := env["NETWORK"]
	port := env["PORT"]
	envString := VarsMapToDockerEnvString(env)

	cmdString := fmt.Sprintf("docker run --rm -t -v __OUTPUT__:__OUTPUT__ --network %s %s -p %s:%s --name %s %s",
		network,
		envString,
		port,
		port,
		serviceName,
		containerImage,
	)

	query := &process_manager.ProcessCreateQuery{
		serviceName,
		"/",
		ozoneWorkingDir,
		cmdString,
		false,
		false,
		env,
	}

	if err := process_manager_client.AddProcess(query); err != nil{
		return err
	}

	return nil
}
