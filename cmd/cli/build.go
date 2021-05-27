package cli

import (
	"fmt"
	"github.com/common-nighthawk/go-figure"
	"github.com/spf13/cobra"
	"log"
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
	for _, b := range builds {
		figure.NewFigure(b.Name, "doom", true).Print()

		scope := config.BuildVars
		scope["CONTEXT"] = context
		scope["SERVICE"] = b.Service
		scope["DIR"] = b.Dir

		err := runIndividual(b, context, config, scope)
		if err != nil {
			log.Fatalln(err)
		}
	}
}

func runIndividual(b *ozoneConfig.Runnable, context string, config *ozoneConfig.OzoneConfig, scope map[string]string) error {
	for _, dependencyName := range b.Depends {
		exists, dependency := config.FetchRunnable(dependencyName)

		if !exists {
			log.Fatalf("Depencdency %s on build %s doesn't exist", dependencyName, b.Name)
		}

		runIndividual(dependency, context, config, scope)
	}

	for _, cs := range b.ContextSteps {
		match, err := regexp.Match(cs.Context, []byte(context))
		if err != nil {
			return err
		}
		if match {
			contextStepVars, err := config.FetchEnvs(cs.WithEnv, scope)
			if err != nil {
				return err
			}
			//scope = ozoneConfig.MergeMaps(scope, runtimeVars) TODO are runtimeVarsNeeded at build?
			for _, step := range cs.Steps {
				fmt.Printf("step %s", step.Type)

				stepVars := ozoneConfig.MergeMaps(contextStepVars, step.WithVars)

				if err != nil {
					return err
				}
				if step.Type == "builtin" {
					switch b.Type {
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
	return nil
}

func runBuildable(step *ozoneConfig.Step, r *ozoneConfig.Runnable, varsMap map[string]string) {
	switch step.Value {
	case "go":
		fmt.Println("gogo")
		err := _go.Build(
			r.Service,
			"micro-a",
			"main.go",
			varsMap,
		)
		if err != nil {
			log.Fatalln(err)
		}
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

			fmt.Println("after")
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
		contextBanner := fmt.Sprintf("context::: %s", context)
		figure.NewFigure(contextBanner, "doom", true).Print()
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