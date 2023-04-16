package docker

import (
	"fmt"
	process_manager_client "github.com/ozone2021/ozone/ozone-daemon-lib/process-manager-client"
	process_manager "github.com/ozone2021/ozone/ozone-daemon-lib/process-manager-queries"
	. "github.com/ozone2021/ozone/ozone-lib/config/config_variable"
	"github.com/ozone2021/ozone/ozone-lib/utils"
	"log"
	"os"
)

func getDockerRunParams() []string {
	return []string{
		"DOCKER_FULL_TAG",
		"PORT",
		"NETWORK",
		"SERVICE",
		//"BUILD_ARGS",
	}
}

//func VarsMapToDockerEnvString(varsMap *VariableMap) string {
//	envString := ""
//	for key, value := range varsMap {
//		envString = fmt.Sprintf("%s-e %s=%s ", envString, key, value)
//	}
//	return envString
//}

func CreateNetworkIfNotExists(serviceName string, env *VariableMap) error {
	network, _ := env.GetVariable("NETWORK")

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

func DeleteContainerIfExists(serviceName string, env *VariableMap) error {
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

func Build(env *VariableMap) error {
	for _, arg := range getDockerRunParams() {
		if err := utils.ParamsOK("DeployDocker", arg, env); err != nil {
			return err
		}
	}

	ozoneWorkingDir, err := os.Getwd()
	if err != nil {
		log.Println(err)
	}

	serviceName, _ := env.GetVariable("SERVICE")

	CreateNetworkIfNotExists(serviceName.String(), env)
	DeleteContainerIfExists(serviceName.String(), env)

	containerImage, _ := env.GetVariable("DOCKER_FULL_TAG")
	network, _ := env.GetVariable("NETWORK")
	port, _ := env.GetVariable("PORT")
	//envString := VarsMapToDockerEnvString(env)

	cmdString := fmt.Sprintf("docker run --rm -t -v __OUTPUT__:__OUTPUT__ --network %s   -p %s:%s --name %s %s",
		network,
		//envString,
		port,
		port,
		serviceName,
		containerImage,
	)

	query := &process_manager.ProcessCreateQuery{
		serviceName.String(),
		"/",
		ozoneWorkingDir,
		cmdString,
		false,
		false,
		env,
	}

	if err := process_manager_client.AddProcess(query); err != nil {
		return err
	}

	return nil
}
