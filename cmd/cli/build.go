package cli

import (
	"fmt"
	"github.com/spf13/cobra"
	"log"
	"net/rpc"
	"os"
	process_manager "ozone-daemon-lib/process-manager"
	"ozone-lib/buildables"
	ozoneConfig "ozone-lib/config"
	"ozone-lib/deployables/docker"
	"ozone-lib/deployables/executable"
	"ozone-lib/deployables/helm"
	_go "ozone-lib/go"
	"regexp"
)

func init() {
	rootCmd.AddCommand(buildCmd)
}

func run(builds []*ozoneConfig.Runnable, config *ozoneConfig.OzoneConfig, context string, runType ozoneConfig.RunnableType) {
	fmt.Println("here3")
	fmt.Println(context)

	for _, b := range builds {
		scope := config.BuildVars
		scope["CONTEXT"] = context
		scope["SERVICE"] = b.Service
		scope["DIR"] = b.Dir
		fmt.Println(b.Name)
		fmt.Println("-")
		for _, cs := range b.ContextSteps {
			match, err := regexp.Match(cs.Context, []byte(context))
			if err != nil {
				log.Fatalln(err)
				return
			}
			if match {
				contextStepVars, err := config.FetchEnvs(cs.WithEnv, scope)
				if err != nil {
					log.Fatalln(err)
					return
				}
				//scope = ozoneConfig.MergeMaps(scope, runtimeVars) TODO are runtimeVarsNeeded at build?
				for _, step := range cs.Steps {
					fmt.Printf("step %s", step.Type)

					stepVars := ozoneConfig.MergeMaps(contextStepVars, step.WithVars)

					if err != nil {
						log.Fatalln(err)
					}
					if step.Type == "builtin" {
						fmt.Println("gogo")
						switch runType {
						case ozoneConfig.BuildType:
							runBuildable(step, b, stepVars)
						case ozoneConfig.DeployType:
							runDeployables(step, b, stepVars)
						//case ozoneConfig.TestTypeType:
						//	runTestables(step, b, stepVars)
						}
					}
				}
			}
		}
	}
}

func runBuildable(step *ozoneConfig.Step, r *ozoneConfig.Runnable, varsMap map[string]string) {
	switch step.Value {
	case "go":
		fmt.Println("gogo")
		_go.Build(
			r.Service,
			"micro-a",
			"main.go",
			varsMap,
		)
	case "buildDockerImage":
		fmt.Println("Building docker image.")
		err := buildables.BuildPushDockerContainer(varsMap)
		if err != nil {
			log.Fatalln(err)
		}
	case "pushDockerImage":
		fmt.Println("Building docker image.")
		err := buildables.PushDockerImage(varsMap)
		if err != nil {
			log.Fatalln(err)
		}
	}
}

func runDeployables(step *ozoneConfig.Step, r *ozoneConfig.Runnable, varsMap map[string]string) {
	if step.Type == "builtin" {
		switch step.Value {
		case "executable":
			fmt.Println("gogo")
			executable.Build(r.Service, varsMap)
		case "helm":
			helm.Deploy(r.Service, varsMap)
		case "runDockerImage":
			err := docker.Build(r.Service, varsMap)
			if err != nil {
				log.Fatalln(err)
			}
		default:
			log.Fatalf("Builtin value not found: %s \n", step.Value)
		}
	}
}






//func deploy(deploys []*ozoneConfig.Runnable, config *ozoneConfig.OzoneConfig, context string) {
//	//varsMap := ozoneConfig.VarsToMap(config.BuildVars)
//	fmt.Println("Deploys")
//	fmt.Println(context)
//
//	for _, b := range deploys {
//		fmt.Println(b.Name)
//		fmt.Println("-")
//		for _, es := range b.ContextSteps {
//			fmt.Printf("Context: %s \n", context)
//			if es.Context == context {
//			 	buildVars := ozoneConfig.VarsToMap(config.BuildVars)
//				varsMap, err := fetchEnvs(config, es.WithEnv, buildVars)
//				varsMap = mergeMaps(buildVars, varsMap)
//				if err != nil {
//					log.Fatalln(err)
//				}
//
//				fmt.Println("Context")
//				for _, step := range es.Steps {
//					fmt.Printf("step %s", step.Type)
//					// TODO merge in step.WithVars into varsMap
//					stepVars := mergeMaps(varsMap, step.WithVars)
//				}
//			}
//		}
//	}
//}

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

		run(builds, config, context, ozoneConfig.BuildType)
		run(deploys, config, context, ozoneConfig.DeployType)
		//tests(tests, config, context)

	},
}