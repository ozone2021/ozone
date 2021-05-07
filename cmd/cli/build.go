package cli

import (
	"errors"
	"fmt"
	"github.com/flosch/pongo2/v4"
	"github.com/spf13/cobra"
	"log"
	"net/rpc"
	"os"
	process_manager "ozone-daemon-lib/process-manager"
	"ozone-lib/buildables"
	ozoneConfig "ozone-lib/config"
	"ozone-lib/deployables/executable"
	"ozone-lib/deployables/helm"
	"ozone-lib/env"
	_go "ozone-lib/go"
)

func init() {
	rootCmd.AddCommand(buildCmd)
}


func build(builds []*ozoneConfig.Runnable, config *ozoneConfig.OzoneConfig, context string) {
	fmt.Println("here3")
	fmt.Println(context)

	for _, b := range builds {
		buildVarsMap := ozoneConfig.VarsToMap(config.BuildVars)
		fmt.Println(b.Name)
		fmt.Println("-")
		for _, es := range b.ContextSteps {
			if es.Context == context {
				for _, step := range es.Steps {
					fmt.Printf("step %s", step.Type)
					buildVarsMap["SERVICE"] = b.Service
					buildVarsMap["DIR"] = b.Dir
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
						case "buildDockerImage":
							fmt.Println("Building docker image.")
							err := buildables.BuildPushDockerContainer(varsMap)
							if err != nil {
								panic(err)
							}
						case "pushDockerImage":
							fmt.Println("Building docker image.")
							err := buildables.PushDockerImage(varsMap)
							if err != nil {
								panic(err)
							}
						}
					}
				}
			}
		}
	}
}

func convertMap(originalMap interface{}) pongo2.Context {
	convertedMap := make(map[string]interface{})
	for key, value := range originalMap.(map[string]string) {
		convertedMap[key] = value
	}

	return convertedMap
}

func renderVars(input string, varsMap map[string]string) string {
	//tpl, err := pongo2.FromString("Hello {{ name|capfirst }}!")
	tpl, err := pongo2.FromString(input)
	if err != nil {
		panic(err)
	}
	context := convertMap(varsMap)
	out, err := tpl.Execute(context)
	if err != nil {
		panic(err)
	}
	return out
}

func mergeMaps(base map[string]string, overwrite map[string]string) map[string]string {
	for k, v := range overwrite {
		base[k] = renderVars(v, base)
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

func deploy(deploys []*ozoneConfig.Runnable, config *ozoneConfig.OzoneConfig, context string) {
	//varsMap := ozoneConfig.VarsToMap(config.BuildVars)
	fmt.Println("Deploys")
	fmt.Println(context)

	for _, b := range deploys {
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
					// TODO merge in step.WithVars into varsMap
					varsMap = mergeMaps(varsMap, step.WithVars)
					if step.Type == "builtin" {
						switch step.Value {
						case "executable":
							fmt.Println("gogo")
							executable.Build(b.Service, varsMap)
						case "helm":
							helm.Deploy(b.Service, varsMap)
						default:
							log.Fatalf("Builtin value not found: %s \n", step.Value)
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

func isBuildable(name string, config *ozoneConfig.OzoneConfig) (bool, *ozoneConfig.Runnable) {
	for _, b := range config.Builds {
		if b.Name == name {
			return true, b
		}
	}
	return false, nil
}

func isDeployable(name string, config *ozoneConfig.OzoneConfig) bool {
	for _, b := range config.Deploys {
		if b.Name == name {
			return true
		}
	}
	return false
}

//func isTestable(name string, config *ozoneConfig.OzoneConfig) bool {
//	for _, b := range config.Tests {
//		if b.Name == name {
//			return true
//		}
//	}
//	return false
//}

func separateRunnables(args []string, config *ozoneConfig.OzoneConfig) ([]*ozoneConfig.Runnable,[]*ozoneConfig.Runnable,[]*ozoneConfig.Runnable) {
	var buildables []*ozoneConfig.Runnable
	var deployables []*ozoneConfig.Runnable
	var testables []*ozoneConfig.Runnable

	for _, runnableName := range args {
		if has, build := config.HasBuild(runnableName); has == true {
			buildables = append(buildables, build)
		}
		if has, deploy := config.HasDeploy(runnableName); has == true {
			deployables = append(deployables, deploy)
		}
		if has, test := config.HasTest(runnableName); has == true {
			deployables = append(testables, test)
		}
		//if isTest
	}

	return buildables, deployables, testables
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
			if has, _ := config.FetchRunnable(arg); has == true {
				continue
			} else {
				log.Fatalf("Config doesn't have runnable: %s \n", arg)
			}
		}

		builds, deploys, _ := separateRunnables(args, config)

		build(builds, config, context)
		deploy(deploys, config, context)
		//tests(tests, config, context)

	},
}