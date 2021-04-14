package cli

import (
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"log"
	"net/rpc"
	"os"
	process_manager "ozone-daemon-lib/process-manager"
	"ozone-lib/buildables"
	ozoneConfig "ozone-lib/config"
	"ozone-lib/env"
	"ozone-lib/executable"
	_go "ozone-lib/go"
)

func init() {
	rootCmd.AddCommand(buildCmd)
}

func build(config *ozoneConfig.OzoneConfig, context string) {
	fmt.Println("here3")
	fmt.Println(context)

	for _, b := range config.Builds {
		buildVarsMap := ozoneConfig.VarsToMap(config.BuildVars)
		fmt.Println(b.Name)
		fmt.Println("-")
		for _, es := range b.ContextSteps {
			fmt.Println("ContextSteps")
			if es.Context == context {
				fmt.Println("Context")
				for _, step := range es.Steps {
					fmt.Printf("step %s", step.Type)
					buildVarsMap["SERVICE"] = b.Service
					varsMap, err := fetchEnvs(config, es.WithEnv, buildVarsMap)
					if err != nil {
						log.Fatalln(err)
					}
					if step.Type == "builtin" {
						fmt.Println("gogo")
						switch step.Value {
						case "go":
							fmt.Println("gogo")
							_go.Build(
								b.Service,
								"micro-a",
								"main.go",
								varsMap,
							)
						case "buildPushDockerImage":
							fmt.Println("Building docker image.")
							buildables.BuildPushDockerContainer(varsMap)
						}
					}
				}
			}
		}
	}
}

func mergeMaps(base map[string]string, overwrite map[string]string) map[string]string {
	for k, v := range overwrite {
		base[k] = v
	}
	return base
}
func fetchBuiltinEnvFromInclude(config *ozoneConfig.OzoneConfig, envName string, varsMap map[string]string) (map[string]string, error) {
	var err error
	fromIncludeMap := make(map[string]string)

	switch envName {
	case "env/from_k8s_secret":
		fromIncludeMap, err = env.FromSecret(varsMap)
	case "env/from_env_file":
		fromIncludeMap, err = env.FromEnvFile(varsMap)
	case "env/docker_submodule_git_hash":
		fromIncludeMap, err = env.FromGitSubmoduleBranchHash(varsMap)
	}

	if err != nil {
		return nil, err
	}

	return fromIncludeMap, nil
}

func fetchEnv(config *ozoneConfig.OzoneConfig, envName string, varsMap map[string]string) (map[string]string, error) {
	nameFound := false
	for _, e := range config.Environments {
		if e.Name == envName {
			nameFound = true
			if len(e.Includes) != 0 {
				for _, incl := range e.Includes {
					var inclVarsMap map[string]string
					var err error
					if incl.Type == "builtin" {
						inclParamVarsMap := mergeMaps(ozoneConfig.VarsToMap(incl.WithVars), varsMap)
						inclVarsMap, err = fetchBuiltinEnvFromInclude(config, incl.Name, inclParamVarsMap)
					} else {
						inclVarsMap, err = fetchEnv(config, incl.Name, varsMap)
					}
					if err != nil {
						return nil, err
					}
					varsMap = mergeMaps(varsMap, inclVarsMap)
				}
			}
			varsMap = mergeMaps(varsMap, ozoneConfig.VarsToMap(e.WithVars))
		}
	}
	if nameFound == false {
		return nil, errors.New(fmt.Sprintf("Environment %s not found \n", envName))
	}

	return varsMap, nil
}

func fetchEnvs(config *ozoneConfig.OzoneConfig, envList []string, varsMap map[string]string) (map[string]string, error) {
	for _, env := range envList {
		fetchedMap, err := fetchEnv(config, env, varsMap)
		if err != nil {
			return nil, err
		}
		varsMap = mergeMaps(varsMap, fetchedMap)
	}
	return varsMap, nil
}

func deploy(config *ozoneConfig.OzoneConfig, context string) {
	//varsMap := ozoneConfig.VarsToMap(config.BuildVars)
	fmt.Println("here3")
	fmt.Println(context)

	for _, b := range config.Deploys {
		fmt.Println(b.Name)
		fmt.Println("-")
		for _, es := range b.ContextSteps {
			fmt.Println("ContextSteps")
			if es.Context == context {
			 	buildVars := ozoneConfig.VarsToMap(config.BuildVars)
				varsMap, err := fetchEnvs(config, es.WithEnv, buildVars)
				varsMap = mergeMaps(buildVars, varsMap)
				if err != nil {
					log.Fatalln(err)
				}

				fmt.Println("Context")
				for _, step := range es.Steps {
					fmt.Printf("step %s", step.Type)
					if step.Type == "builtin" {
						fmt.Println("gogo")
						switch step.Value {
						case "executable":
							fmt.Println("gogo")
							executable.Build(b.Service, varsMap)
						}
					}
				}
			}
		}
	}
}



func fetchContext(defaultContext string) (string, error) {
	ozoneWorkingDir, err := os.Getwd()
	if err != nil {
		log.Println(err)
	}

	request := process_manager.ContextQueryRequest{
		ozoneWorkingDir,
		"",
		defaultContext,
	}

	var response process_manager.ContextQueryResponse

	client, err := rpc.DialHTTP("tcp", ":8000")
	if err != nil {
		log.Fatal("dialing:", err)
	}
	defer client.Close()
	err = client.Call("ProcessManager.ContextQuery", request, &response)

	return response.Context, err
}

var buildCmd = &cobra.Command{
	Use:   "b",
	Long:  `List running processes`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Building...")

		config := ozoneConfig.ReadConfig()
		context, err := fetchContext(config.ContextInfo.Default)

		if err != nil {
			log.Fatal("FetchContext error:", err)
		}

		for _, arg := range args {
			if config.HasRunnable(arg) {
				continue
			} else {
				log.Fatalf("Config doesn't have runnable: %s \n", arg)
			}
		}
		build(config, context)
		deploy(config, context)

		//_go.Build("microA", "micro-a", "main.go")
		//
	},
}