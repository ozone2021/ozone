package cli

//
//import (
//	"errors"
//	"fmt"
//	"github.com/common-nighthawk/go-figure"
//	"github.com/ozone2021/ozone/ozone-daemon-lib/cache"
//	process_manager_client "github.com/ozone2021/ozone/ozone-daemon-lib/process-manager-client"
//	"github.com/ozone2021/ozone/ozone-lib/buildables"
//	ozoneConfig "github.com/ozone2021/ozone/ozone-lib/config"
//	"github.com/ozone2021/ozone/ozone-lib/config/config_keys"
//	"github.com/ozone2021/ozone/ozone-lib/config/config_utils"
//	"github.com/ozone2021/ozone/ozone-lib/config/config_variable"
//	"github.com/ozone2021/ozone/ozone-lib/deployables/docker"
//	"github.com/ozone2021/ozone/ozone-lib/deployables/executable"
//	"github.com/ozone2021/ozone/ozone-lib/deployables/helm"
//	_go "github.com/ozone2021/ozone/ozone-lib/go"
//	"github.com/ozone2021/ozone/ozone-lib/utilities"
//	"github.com/spf13/cobra"
//	"log"
//	"path"
//	"path/filepath"
//)
//
//func init() {
//	rootCmd.AddCommand(runCmd)
//}
//
//func hasCaching(runnable *ozoneConfig.Runnable) bool {
//	return runnable.SourceFiles != nil
//}
//
//func checkCache(runnable *ozoneConfig.Runnable, sourceFiles []string) bool {
//	if headless == true || hasCaching(runnable) == false {
//		return false
//	}
//	hash, err := getBuildHash(runnable.Name, sourceFiles)
//	if err != nil {
//		log.Fatalln(err)
//		return false
//	}
//	if hash == "" {
//		return false
//	}
//
//	runnableName := runnable.Name
//	cachedHash := process_manager_client.CacheCheck(ozoneWorkingDir, runnableName)
//	return cachedHash == hash
//}
//
//func getBuildHash(runnableName string, sourceFiles []string) (string, error) {
//	ozonefilePath := path.Join(ozoneWorkingDir, "Ozonefile")
//
//	ozonefileEditTime, err := cache.FileLastEdit(ozonefilePath)
//
//	if err != nil {
//		return "", err
//	}
//
//	filesDirsLastEditTimes := []int64{ozonefileEditTime}
//
//	for _, filePath := range sourceFiles {
//		editTime, err := cache.FileLastEdit(filePath)
//
//		if err != nil {
//			return "", errors.New(fmt.Sprintf("Source file %s for runnable %s is missing.", filePath, runnableName))
//		}
//
//		filesDirsLastEditTimes = append(filesDirsLastEditTimes, editTime)
//	}
//
//	hash := cache.Hash(filesDirsLastEditTimes...)
//	return hash, nil
//}
//
//func whenScript(script string, varsMap *config_variable.VariableMap) (bool, error) {
//	exitCode, err := utilities.RunBashScript(script, varsMap)
//	if err != nil {
//		log.Printf("WARNING: code %d, potential err=%s", exitCode, err.Error())
//	}
//	switch exitCode {
//	case 0:
//		return true, nil
//	case -1:
//		return false, err
//	default:
//		return false, nil
//	}
//}
//
//func mergeMapStringString(map1, map2 map[string]string) {
//	for k, v := range map2 {
//		map1[k] = v
//	}
//}
//
//func copyMapStringString(m map[string]string) map[string]string {
//	newMap := make(map[string]string)
//	for k, v := range m {
//		newMap[k] = v
//	}
//	return newMap
//}
//
//func runPipeline(pipelines []*ozoneConfig.Runnable, config *ozoneConfig.OzoneConfig, context string) {
//	for _, pipeline := range pipelines {
//		var runnables []*ozoneConfig.Runnable
//		for _, dependency := range pipeline.Depends {
//			exists, dependencyRunnable := config.FetchRunnable(dependency.Name)
//
//			if !exists {
//				log.Fatalf("Dependency %s on pipeline %s doesn't exist", dependency.Name, pipeline.Name)
//			}
//			runnables = append(runnables, dependencyRunnable)
//		}
//		run(runnables, config, context)
//	}
//}
//
//func run(builds []*ozoneConfig.Runnable, config *ozoneConfig.OzoneConfig, context string) {
//	ordinal := 0
//
//	topLevelScope := config_variable.CopyOrCreateNew(config.BuildVars)
//	topLevelScope.AddVariable(config_variable.NewStringVariable("CONTEXT", context), ordinal)
//	topLevelScope.AddVariable(config_variable.NewStringVariable("OZONE_WORKING_DIR", ozoneWorkingDir), ordinal)
//
//	for _, b := range builds {
//		asOutput := make(map[string]string)
//		_, _, err := runIndividual(b, ordinal, context, config, config_variable.CopyOrCreateNew(topLevelScope), asOutput, true)
//		if err != nil {
//			log.Fatalf("Error %s in runnable %s", err, b.Name)
//		}
//	}
//}
//
//func runIndividual(runnable *ozoneConfig.Runnable, ordinal int, context string, config *ozoneConfig.OzoneConfig, buildScope *config_variable.VariableMap, asOutput map[string]string, shouldRun bool) (*config_variable.VariableMap, *config_variable.VariableMap, error) {
//	ordinal++
//
//	if runnable.Service != "" {
//		buildScope.AddVariableWithoutOrdinality(config_variable.NewStringVariable("SERVICE", runnable.Service))
//	}
//	if runnable.Dir != "" {
//		buildScope.AddVariable(config_variable.NewStringVariable("DIR", runnable.Dir), ordinal)
//	}
//	buildScope.AddVariable(config_variable.NewStringVariable("NAME", runnable.Name), ordinal)
//	//runnable.SourceFiles. TODO set name
//
//	var sourceFiles []string
//	for _, file := range runnable.SourceFiles {
//		rendered, err := buildScope.RenderSentence(file)
//		if err != nil {
//			return nil, nil, err
//		}
//		sourceFiles = append(sourceFiles, filepath.Join(ozoneWorkingDir, rendered))
//	}
//	if sourceFiles != nil {
//		buildScope.AddVariable(config_variable.NewSliceVariable(config_keys.SOURCE_FILES_KEY, sourceFiles), ordinal)
//	}
//	buildScope.SelfRender()
//
//	// TODO add support for list variables.
//
//	if hasCaching(runnable) {
//		cacheHash, err := getBuildHash(runnable.Name, sourceFiles)
//		if err != nil {
//			return nil, nil, err
//		}
//		buildScope.AddVariable(config_variable.NewStringVariable("CACHE_HASH_ENTIRE", cacheHash), ordinal)
//	}
//
//	if shouldRun == true {
//		figure.NewFigure(runnable.Name, "doom", true).Print()
//	}
//	if runnable.Type == ozoneConfig.BuildType && checkCache(runnable, sourceFiles) == true {
//		log.Printf("Info: build files for %s unchanged from cache. \n", runnable.Name)
//		return nil, nil, nil
//	}
//
//	contextEnvVars := config_variable.NewVariableMap()
//	for _, contextEnv := range runnable.ContextEnv {
//		buildScope.MergeVariableMaps(contextEnv.WithVars)
//		inPattern, err := config_utils.ContextInPattern(context, contextEnv.Context, buildScope)
//
//		if err != nil {
//			return nil, nil, err
//		}
//		if inPattern {
//			fetchedEnvs, err := config.FetchEnvs(ordinal, contextEnv.WithEnv, buildScope)
//			if err != nil {
//				return nil, nil, err
//			}
//			contextEnvVars.MergeVariableMaps(fetchedEnvs)
//		}
//	}
//	contextEnvVars.IncrementOrdinal(ordinal)
//
//	//runnableVars, err := config.FetchEnvs(runnable.WithEnv, buildScope)
//	//if err != nil  {
//	//	return err
//	//}
//	runnableBuildScope := config_variable.CopyOrCreateNew(contextEnvVars)
//	runnableBuildScope.MergeVariableMaps(buildScope)
//
//	outputVars := config_variable.NewVariableMap()
//	contextOutputVars, err := runnableBuildScope.AsOutput(asOutput)
//	if err != nil {
//		return nil, nil, err
//	}
//	outputVars.MergeVariableMaps(contextOutputVars)
//
//	for _, contextConditional := range runnable.ContextConditionals {
//		if shouldRun == false {
//			break
//		}
//		inPattern, err := config_utils.ContextInPattern(context, contextConditional.Context, runnableBuildScope)
//		if err != nil {
//			return nil, nil, err
//		}
//		if inPattern {
//			// WhenScript
//			for _, script := range contextConditional.WhenScript {
//				exitCode, err := utilities.RunBashScript(script, runnableBuildScope)
//				switch exitCode {
//				case 0:
//					shouldRun = true
//					continue
//				case 3:
//					return nil, nil, err
//				default:
//					shouldRun = false
//					log.Printf("Not running, contextConditional whenScript not satisfied: %s", script)
//					break
//				}
//			}
//			// When Not script
//			if shouldRun == false {
//				break
//			}
//			for _, script := range contextConditional.WhenNotScript {
//				exitCode, err := utilities.RunBashScript(script, runnableBuildScope)
//				switch exitCode {
//				case 0:
//					shouldRun = false
//					log.Printf("Not running, contextConditional whenNotScript not satisfied: %s", script)
//					break
//				case 3:
//					return nil, nil, err
//				default:
//					shouldRun = true
//					continue
//				}
//			}
//		}
//	}
//
//	outputVarsFromDependentStep := config_variable.NewVariableMap()
//	fullEnvFromDependentStep := config_variable.CopyOrCreateNew(runnableBuildScope)
//	for _, dependency := range runnable.Depends {
//		exists, dependencyRunnable := config.FetchRunnable(dependency.Name)
//
//		if !exists {
//			log.Fatalf("Dependency %s on build %s doesn't exist", dependency.Name, runnable.Name)
//		}
//
//		dependencyScope := config_variable.CopyOrCreateNew(runnableBuildScope)
//		dependencyScope.MergeVariableMaps(contextEnvVars)
//		dependencyWithVars := config_variable.CopyOrCreateNew(dependency.WithVars)
//		dependencyWithVars.IncrementOrdinal(ordinal)
//		dependencyScope.MergeVariableMaps(dependencyWithVars)
//		var err error
//
//		dependencyVarAsOutput := copyMapStringString(dependency.VarOutputAs)
//		mergeMapStringString(dependencyVarAsOutput, asOutput)
//		dependencyScope.MergeVariableMaps(outputVars)
//
//		envFromDependentStep := config_variable.CopyOrCreateNew(dependencyScope)
//		outputVarsFromDependentStep, envFromDependentStep, err = runIndividual(dependencyRunnable, ordinal, context, config, dependencyScope, dependencyVarAsOutput, shouldRun)
//		if err != nil {
//			return nil, nil, err
//		}
//		outputVars.MergeVariableMaps(outputVarsFromDependentStep)
//		fullEnvFromDependentStep.MergeVariableMaps(envFromDependentStep)
//	}
//
//	runnableBuildScope.MergeVariableMaps(outputVarsFromDependentStep)
//	fullEnvFromDependentStep.MergeVariableMaps(runnableBuildScope)
//
//	if shouldRun == false {
//		return outputVars, config_variable.CopyOrCreateNew(fullEnvFromDependentStep), nil
//	}
//
//	for _, cs := range runnable.ContextSteps {
//		match, err := config_utils.ContextInPattern(context, cs.Context, runnableBuildScope)
//		if err != nil {
//			return nil, nil, err
//		}
//		if match {
//			//contextStepVars, err := config.FetchEnvs(ordinal, cs.WithEnv, runnableBuildScope)
//			//contextStepVars = config_utils.MergeMapsSelfRender(ordinal, contextEnvVars, contextStepVars)
//			//contextStepBuildScope := config_utils.MergeMapsSelfRender(ordinal, buildScope, contextStepVars)
//			contextStepVars, err := config.FetchEnvs(ordinal, cs.WithEnv, runnableBuildScope)
//			contextStepVars.MergeVariableMaps(contextEnvVars)
//			contextStepBuildScope := config_variable.CopyOrCreateNew(buildScope)
//			contextStepBuildScope.MergeVariableMaps(contextStepVars)
//			if err != nil {
//				return nil, nil, err
//			}
//			//scope = ozoneConfig.MergeMapsSelfRender(scope, runtimeVars) TODO are runtimeVarsNeeded at build?
//			for _, step := range cs.Steps {
//				stepVars := config_variable.CopyOrCreateNew(step.WithVars)
//				stepVars.IncrementOrdinal(ordinal)
//				stepVars.MergeVariableMaps(contextStepBuildScope)
//				stepVars.MergeVariableMaps(contextStepVars)
//
//				stepOutputVars, err := stepVars.AsOutput(step.VarOutputAs)
//				if err != nil {
//					return nil, nil, err
//				}
//				outputVars.MergeVariableMaps(stepOutputVars)
//				fmt.Printf("Step: %s \n", step.Name)
//
//				if err != nil {
//					return nil, nil, err
//				}
//				if step.Type == "builtin" {
//					switch runnable.Type {
//					case ozoneConfig.PreUtilityType:
//						runUtility(step, runnable, stepVars)
//					case ozoneConfig.BuildType:
//						runBuildable(step, runnable, stepVars)
//					case ozoneConfig.DeployType:
//						runDeployables(step, runnable, stepVars)
//					case ozoneConfig.TestType:
//						runTestable(step, runnable, stepVars)
//					case ozoneConfig.PostUtilityType:
//						runUtility(step, runnable, stepVars)
//					}
//				}
//			}
//		}
//	}
//	// TODO update cache
//	if headless == false && runnable.Type == ozoneConfig.BuildType && hasCaching(runnable) {
//		updateCache(runnable, sourceFiles)
//	}
//
//	return outputVars, fullEnvFromDependentStep, nil
//}
//
//func updateCache(runnable *ozoneConfig.Runnable, sourceFiles []string) {
//	hash, err := getBuildHash(runnable.Name, sourceFiles)
//	if err != nil {
//		log.Fatalln(err)
//	}
//
//	log.Printf("Cache updated for %s \n", runnable.Name)
//	process_manager_client.CacheUpdate(ozoneWorkingDir, runnable.Name, hash)
//}
//
//func runBuildable(step *ozoneConfig.Step, r *ozoneConfig.Runnable, varsMap *config_variable.VariableMap) {
//	switch step.Name {
//	case "go":
//		fmt.Println("gogo")
//		err := _go.Build(
//			r.Service,
//			"micro-a",
//			"main.go",
//			varsMap,
//		)
//		if err != nil {
//			log.Fatalln(err)
//		}
//	case "buildDockerImage":
//		fmt.Println("Building docker image.")
//		err := buildables.BuildDockerContainer(varsMap)
//		if err != nil {
//			log.Fatalln(err)
//		}
//	case "bashScript":
//		script, ok := varsMap.GetVariable("SCRIPT")
//		if !ok {
//			log.Fatalf("Script not set for runnable step %s", r.Name)
//		}
//		_, err := utilities.RunBashScript(script.String(), varsMap)
//		if err != nil {
//			log.Fatalln(err)
//		}
//	case "pushDockerImage":
//		fmt.Println("Building docker image.")
//		err := buildables.PushDockerImage(varsMap)
//		if err != nil {
//			log.Fatalln(err)
//		}
//	}
//}
//
//func runTestable(step *ozoneConfig.Step, r *ozoneConfig.Runnable, varsMap *config_variable.VariableMap) {
//	switch step.Name {
//	case "bashScript":
//		script, ok := varsMap.GetVariable("SCRIPT")
//		if !ok {
//			log.Fatalf("Script not set for runnable step %s", r.Name)
//		}
//		_, err := utilities.RunBashScript(script.String(), varsMap)
//		if err != nil {
//			log.Fatalln(err)
//		}
//	default:
//		log.Fatalf("Testable value not found: %s \n", step.Name)
//	}
//}
//
//func runUtility(step *ozoneConfig.Step, r *ozoneConfig.Runnable, varsMap *config_variable.VariableMap) {
//	switch step.Name {
//	case "bashScript":
//		script, ok := varsMap.GetVariable("SCRIPT")
//		if !ok {
//			log.Fatalf("Script not set for runnable step %s", r.Name)
//		}
//		_, err := utilities.RunBashScript(script.String(), varsMap)
//		if err != nil {
//			log.Fatalln(err)
//		}
//	default:
//		log.Fatalf("Utility value not found: %s \n", step.Name)
//	}
//}
//
//func runDeployables(step *ozoneConfig.Step, r *ozoneConfig.Runnable, varsMap *config_variable.VariableMap) {
//	if step.Type == "builtin" {
//		var err error
//		switch step.Name {
//		case "executable":
//			err = executable.Build(r.Service, varsMap)
//		case "helm":
//			err = helm.Deploy(r.Service, varsMap)
//		case "runDockerImage":
//			err = docker.Build(varsMap)
//		case "bashScript":
//			script, ok := varsMap.GetVariable("SCRIPT")
//			if !ok {
//				log.Fatalf("Script not set for runnable step %s", r.Name)
//			}
//			_, err = utilities.RunBashScript(script.String(), varsMap)
//		default:
//			log.Fatalf("Builtin value not found: %s \n", step.Name)
//		}
//		if err != nil {
//			log.Fatalln(err)
//		}
//	}
//}
//
////func deploy(deploys []*ozoneConfig.Runnable, config *ozoneConfig.OzoneConfig, context string) {
////	//varsMap := ozoneConfig.VarsToMap(config.BuildVars)
////	fmt.Println("Deploys")
////	fmt.Println(context)
////
////	for _, b := range deploys {
////		fmt.Println(b.Name)
////		fmt.Println("-")
////		for _, es := range b.ContextSteps {
////			fmt.Printf("Context: %s \n", context)
////			if es.Context == context {
////			 	buildVars := ozoneConfig.VarsToMap(config.BuildVars)
////				varsMap, err := fetchEnvs(config, es.WithEnv, buildVars)
////				varsMap = mergeMaps(buildVars, varsMap)
////				if err != nil {
////					log.Fatalln(err)
////				}
////
////				fmt.Println("Context")
////				for _, step := range es.Steps {
////					fmt.Printf("step %s", step.Type)
////					// TODO merge in step.InputVars into varsMap
////					stepVars := mergeMaps(varsMap, step.InputVars)
////				}
////			}
////		}
////	}
////}
//
//func separateRunnables(args []string, config *ozoneConfig.OzoneConfig) ([]*ozoneConfig.Runnable, []*ozoneConfig.Runnable, []*ozoneConfig.Runnable, []*ozoneConfig.Runnable, []*ozoneConfig.Runnable, []*ozoneConfig.Runnable) {
//	var preUtilities []*ozoneConfig.Runnable
//	var buildables []*ozoneConfig.Runnable
//	var deployables []*ozoneConfig.Runnable
//	var testables []*ozoneConfig.Runnable
//	var pipelines []*ozoneConfig.Runnable
//	var postUtilities []*ozoneConfig.Runnable
//
//	for _, runnableName := range args {
//		if has, utility := config.HasPreUtility(runnableName); has == true {
//			preUtilities = append(preUtilities, utility)
//		}
//		if has, build := config.HasBuild(runnableName); has == true {
//			buildables = append(buildables, build)
//		}
//		if has, deploy := config.HasDeploy(runnableName); has == true {
//			deployables = append(deployables, deploy)
//		}
//		if has, pipeline := config.HasPipeline(runnableName); has == true {
//			pipelines = append(pipelines, pipeline)
//		}
//		if has, test := config.HasTest(runnableName); has == true {
//			testables = append(testables, test)
//		}
//		if has, utility := config.HasPostUtility(runnableName); has == true {
//			postUtilities = append(postUtilities, utility)
//		}
//
//	}
//
//	return preUtilities, buildables, deployables, testables, pipelines, postUtilities
//}
//
//var runCmd = &cobra.Command{
//	Use:  "r",
//	Long: `List running processes`,
//	Run: func(cmd *cobra.Command, args []string) {
//		context := config_utils.FetchContext(cmd, ozoneWorkingDir, config)
//
//		contextBanner := fmt.Sprintf("context::: %s", context)
//		figure.NewFigure(contextBanner, "doom", true).Print()
//		for _, arg := range args {
//			if has, _ := config.FetchRunnable(arg); has == true {
//				continue
//			} else {
//				log.Fatalf("Config doesn't have runnable: %s \n", arg)
//			}
//		}
//
//		preUtilities, builds, deploys, tests, pipelines, postUtilities := separateRunnables(args, config)
//
//		run(preUtilities, config, context)
//		run(builds, config, context)
//		run(deploys, config, context)
//		runPipeline(pipelines, config, context)
//		run(tests, config, context)
//		run(postUtilities, config, context)
//		//tests(tests, config, context)
//
//	},
//}
