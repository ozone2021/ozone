package runspec

import (
	"errors"
	"fmt"
	"github.com/oleiade/lane/v2"
	"github.com/ozone2021/ozone/ozone-daemon-lib/cache"
	process_manager_client "github.com/ozone2021/ozone/ozone-daemon-lib/process-manager-client"
	"github.com/ozone2021/ozone/ozone-lib/buildables"
	"github.com/ozone2021/ozone/ozone-lib/config"
	"github.com/ozone2021/ozone/ozone-lib/config/config_keys"
	"github.com/ozone2021/ozone/ozone-lib/config/config_utils"
	. "github.com/ozone2021/ozone/ozone-lib/config/config_variable"
	"github.com/ozone2021/ozone/ozone-lib/deployables/docker"
	"github.com/ozone2021/ozone/ozone-lib/deployables/executable"
	"github.com/ozone2021/ozone/ozone-lib/deployables/helm"
	"github.com/ozone2021/ozone/ozone-lib/utilities"
	"gopkg.in/yaml.v3"
	"log"
	"os"
	"path"
	"path/filepath"
)

type RunspecStep struct {
	Type        string             `yaml:"type"`
	Name        string             `yaml:"name"`
	Scope       *DifferentialScope `yaml:"scope"`
	VarOutputAs map[string]string  `yaml:"var_output_as"`
}

type DifferentialScope struct {
	parentScope *VariableMap
	scope       *VariableMap
}

func NewDifferentialScope(parentScope *VariableMap, scope *VariableMap) *DifferentialScope {
	return &DifferentialScope{
		parentScope: parentScope,
		scope:       scope,
	}
}

func (ds *DifferentialScope) GetScope() *VariableMap {
	return ds.scope
}

func (ds *DifferentialScope) MarshalYAML() (interface{}, error) {
	if ds == nil {
		return nil, nil
	}
	diff, err := ds.parentScope.Diff(ds.scope)
	if err != nil {
		return nil, err
	}

	b, err := yaml.Marshal(diff)

	return string(b), err
}

//		ContextConditionals []*ContextConditional `yaml:"context_conditionals"` # TODO save whether satisified
//	 Steps is the depends and contextSteps merged
type RunspecRunnable struct {
	Name         string               `yaml:"name"`
	Ordinal      int                  `yaml:"ordinal"`
	Service      string               `yaml:"service"`
	SourceFiles  []string             `yaml:"source_files"`
	Dir          string               `yaml:"dir"`
	BuildScope   *DifferentialScope   `yaml:"scope"`
	Conditionals *RunspecConditionals `yaml:"conditionals"`
	Steps        []*RunspecStep       `yaml:"steps"`
	Type         config.RunnableType  `yaml:"RunnableType"`
}

type CallStack struct {
	RootRunnableName string              `yaml:"root_runnable_name"`
	RootRunnableType config.RunnableType `yaml:"root_runnable_type"`
	Hash             string              `yaml:"hash"`
	SourceFiles      []string            `yaml:"source_files"`
	Runnables        []*RunspecRunnable  `yaml:"runnables"`
}

func (cs *CallStack) hasCaching() bool {
	return cs.SourceFiles == nil
}

func (cs *CallStack) getBuildHash(ozoneWorkingDir string) (string, error) {
	ozonefilePath := path.Join(ozoneWorkingDir, "Ozonefile")

	ozonefileEditTime, err := cache.FileLastEdit(ozonefilePath)

	if err != nil {
		return "", err
	}

	filesDirsLastEditTimes := []int64{ozonefileEditTime}

	for _, filePath := range cs.SourceFiles {
		editTime, err := cache.FileLastEdit(filePath)

		if err != nil {
			return "", errors.New(fmt.Sprintf("Source file %s for runnable %s is missing.", filePath, cs.RootRunnableName))
		}

		filesDirsLastEditTimes = append(filesDirsLastEditTimes, editTime)
	}

	hash := cache.Hash(filesDirsLastEditTimes...)
	return hash, nil
}

type Runspec struct {
	config       *config.OzoneConfig
	ProjectName  string                               `yaml:"project"`
	Context      string                               `yaml:"context"`
	OzoneWorkDir string                               `yaml:"work_dir"` // move into config
	BuildVars    *VariableMap                         `yaml:"build_vars"`
	CallStacks   map[config.RunnableType][]*CallStack `yaml:"call_stack"`
}

func NewRunspec(context, ozoneWorkingDir string, ozoneConfig *config.OzoneConfig) *Runspec {
	systemEnvVars := OSEnvToVarsMap()

	renderedBuildVars := ozoneConfig.BuildVars
	renderedBuildVars.RenderNoMerge(systemEnvVars)
	renderedBuildVars.SelfRender()

	runspec := &Runspec{
		config:       ozoneConfig,
		ProjectName:  ozoneConfig.ProjectName,
		Context:      context,
		OzoneWorkDir: ozoneWorkingDir,
		BuildVars:    renderedBuildVars,
		CallStacks:   make(map[config.RunnableType][]*CallStack),
	}

	return runspec
}

func (wt *Runspec) getConfigRunnableSourceFiles(configRunnable *config.Runnable, buildScope *VariableMap, ordinal int) []string {
	sourceFiles, err := buildScope.RenderList(configRunnable.SourceFiles)
	if err != nil {
		log.Fatalln("Couldn't render source files for runnable %s with err: %s", configRunnable.Name, err)
	}

	sourceFilesVar := NewSliceVariable(config_keys.SOURCE_FILES_KEY, sourceFiles)
	buildScope.AddVariable(sourceFilesVar, ordinal)

	outputSourceFiles := make([]string, len(sourceFiles))
	for k, v := range sourceFiles {
		renderedPrepend, err := buildScope.RenderSentence(configRunnable.SourceFilesPrepend)
		if err != nil {
			log.Fatalln("Couldn't render sourceFilePrepend for runnable %s with err: %s", configRunnable.Name, err)
		}
		outputSourceFiles[k] = filepath.Join(wt.OzoneWorkDir, renderedPrepend, v)
	}

	return outputSourceFiles
}

func (wt *Runspec) FetchContextEnvs(ordinal int, buildScope *VariableMap, runnable *config.Runnable) (*VariableMap, error) {
	contextEnvVars := NewVariableMap()
	for _, contextEnv := range runnable.ContextEnv {
		buildScope.MergeVariableMaps(contextEnv.WithVars)
		inPattern, err := config_utils.ContextInPattern(wt.Context, contextEnv.Context, buildScope)

		if err != nil {
			return nil, err
		}
		if inPattern {
			fetchedEnvs, err := wt.config.FetchEnvs(ordinal, contextEnv.WithEnv, buildScope)
			if err != nil {
				return nil, err
			}
			contextEnvVars.MergeVariableMaps(fetchedEnvs)
		}
	}
	contextEnvVars.IncrementOrdinal(ordinal)

	return contextEnvVars, nil
}

func (wt *Runspec) ContextStepsFlatten(configRunnable *config.Runnable, buildscope *VariableMap, ordinal int) ([]*RunspecStep, error) {
	var steps []*RunspecStep
	for _, cs := range configRunnable.ContextSteps {
		match, err := config_utils.ContextInPattern(wt.Context, cs.Context, buildscope)
		if err != nil {
			log.Fatalln("Err in ContextStep context %s in runnable %s, err: %s", cs.Context, configRunnable.Name, err)
		}
		if !match {
			continue
		}

		contextStepVars, err := wt.config.FetchEnvs(ordinal, cs.WithEnv, buildscope)
		contextStepVars.MergeVariableMaps(buildscope)
		if err != nil {
			return nil, err
		}

		for _, step := range cs.Steps {
			stepVars := CopyOrCreateNew(step.WithVars)
			stepVars.IncrementOrdinal(ordinal) // TODO should be part of Copy/CreateNew
			stepVars.MergeVariableMaps(contextStepVars)

			//stepOutputVars, err := stepVars.AsOutput(step.VarOutputAs)
			//if err != nil {
			//	return nil, err
			//}
			//outputVars.MergeVariableMaps(stepOutputVars)
			fmt.Printf("Step: %s \n", step.Name)

			steps = append(steps, &RunspecStep{
				Type: step.Type,
				Name: step.Name,
				Scope: &DifferentialScope{
					parentScope: buildscope,
					scope:       stepVars,
				},
				VarOutputAs: nil,
			})
		}
	}

	return steps, nil
}

func (wt *Runspec) ConvertConfigRunnableStackItemToRunspecRunnable(configRunnableStackItem *ConfigRunnableStackItem, ordinal int) (*RunspecRunnable, error) {
	configRunnable := configRunnableStackItem.ConfigRunnable
	buildScope := configRunnableStackItem.buildScope
	parentScope := configRunnableStackItem.parentScope

	service, dir := addCallstackScopeVars(configRunnable, buildScope, ordinal)

	sourceFiles := wt.getConfigRunnableSourceFiles(configRunnable, buildScope, ordinal)
	buildScope.AddVariable(NewSliceVariable(config_keys.SOURCE_FILES_KEY, sourceFiles), ordinal)

	contextEnvs, err := wt.FetchContextEnvs(ordinal, buildScope, configRunnable)
	if err != nil {
		return nil, err
	}

	runnableBuildScope := CopyOrCreateNew(contextEnvs)
	runnableBuildScope.MergeVariableMaps(buildScope)
	diffBuildScope := NewDifferentialScope(parentScope, runnableBuildScope)

	steps, err := wt.ContextStepsFlatten(configRunnable, buildScope, ordinal)
	if err != nil {
		return nil, err
	}

	runspecRunnable := &RunspecRunnable{
		Name:         configRunnable.Name,
		Ordinal:      ordinal,
		Service:      service,
		SourceFiles:  sourceFiles,
		Dir:          dir,
		BuildScope:   diffBuildScope,
		Conditionals: ConvertContextConditional(runnableBuildScope, configRunnable, wt.Context),
		Steps:        steps,
		Type:         configRunnable.Type,
	}

	return runspecRunnable, nil
}

type ConfigRunnableStackItem struct {
	ConfigRunnable *config.Runnable
	parentScope    *VariableMap
	buildScope     *VariableMap
}

func NewConfigRunnableStackItem(configRunnable *config.Runnable, parentScope *VariableMap, buildScope *VariableMap) *ConfigRunnableStackItem {
	return &ConfigRunnableStackItem{
		ConfigRunnable: configRunnable,
		parentScope:    parentScope,
		buildScope:     buildScope,
	}
}

func addCallstackScopeVars(runnable *config.Runnable, buildScope *VariableMap, ordinal int) (string, string) {
	service, _ := buildScope.RenderSentence(runnable.Service)
	dir, _ := buildScope.RenderSentence(runnable.Dir)

	if runnable.Service != "" {
		buildScope.AddVariableWithoutOrdinality(NewStringVariable("SERVICE", service))
	}
	if runnable.Dir != "" {
		buildScope.AddVariable(NewStringVariable("DIR", dir), ordinal)
	}
	buildScope.AddVariable(NewStringVariable("NAME", runnable.Name), ordinal)

	buildScope.SelfRender()

	return service, dir
}

func (wt *Runspec) ExecuteCallstacks() error {
	runOrder := []config.RunnableType{
		config.PreUtilityType,
		config.BuildType,
		config.DeployType,
		config.TestType,
		config.PipelineType,
		config.PostUtilityType,
	}
	for _, runnableType := range runOrder {
		for _, callstack := range wt.CallStacks[runnableType] {
			if callstack.RootRunnableType == config.BuildType && wt.checkCache(callstack) == true {
				log.Printf("Info: build files for %s unchanged from cache. \n", callstack.RootRunnableName)
				continue
			}
			for _, runspecRunnable := range callstack.Runnables {
				if runspecRunnable.Conditionals.Satisfied == true {
					runspecRunnable.RunSteps()
				}
			}
		}
	}

	return nil
}

func (wtr *RunspecRunnable) RunSteps() {
	for _, step := range wtr.Steps {
		step.Scope.scope.MergeVariableMaps(wtr.BuildScope.scope)
		log.Printf("Running step %s \n", step.Name)
		step.RunStep(wtr.Type)
	}
}

func (step *RunspecStep) RunStep(runnableType config.RunnableType) {
	if step.Type == "builtin" {
		switch runnableType {
		case config.PreUtilityType:
			step.runUtility()
		case config.BuildType:
			step.runBuildable()
		case config.DeployType:
			step.runDeployables()
		case config.TestType:
			step.runTestable()
		case config.PostUtilityType:
			step.runUtility()
		}
	}
}

func (step *RunspecStep) runBuildable() {
	switch step.Name {
	case "buildDockerImage":
		fmt.Println("Building docker image.")
		err := buildables.BuildDockerContainer(step.Scope.scope)
		if err != nil {
			log.Fatalln(err)
		}
	case "bashScript":
		script, ok := step.Scope.scope.GetVariable("SCRIPT")
		if !ok {
			log.Fatalf("Script not set for runnable step %s", "r.Name") // TODO wrap err
		}
		_, err := utilities.RunBashScript(script.String(), step.Scope.scope)
		if err != nil {
			log.Fatalln(err)
		}
	case "pushDockerImage":
		fmt.Println("Building docker image.")
		err := buildables.PushDockerImage(step.Scope.scope)
		if err != nil {
			log.Fatalln(err)
		}
	}
}

func (step *RunspecStep) runTestable() {
	switch step.Name {
	case "bashScript":
		script, ok := step.Scope.scope.GetVariable("SCRIPT")
		if !ok {
			log.Fatalf("Script not set for runnable step %s", step.Name) // TODO wrap err
		}
		_, err := utilities.RunBashScript(script.String(), step.Scope.scope)
		if err != nil {
			log.Fatalln(err)
		}
	default:
		log.Fatalf("Testable value not found: %s \n", step.Name)
	}
}

func (step *RunspecStep) runUtility() {
	switch step.Name {
	case "bashScript":
		script, ok := step.Scope.scope.GetVariable("SCRIPT")
		if !ok {
			log.Fatalf("Script not set for runnable step %s", step.Name) // TODO wrap err
		}
		_, err := utilities.RunBashScript(script.String(), step.Scope.scope)
		if err != nil {
			log.Fatalln(err)
		}
	default:
		log.Fatalf("Utility value not found: %s \n", step.Name)
	}
}

func (step *RunspecStep) runDeployables() {
	if step.Type == "builtin" {
		var err error
		switch step.Name {
		case "executable":
			err = executable.Build(step.Scope.scope)
		case "helm":
			err = helm.Deploy(step.Scope.scope)
		case "runDockerImage":
			err = docker.Build(step.Scope.scope)
		case "bashScript":
			script, ok := step.Scope.scope.GetVariable("SCRIPT")
			if !ok {
				log.Fatalf("Script not set for runnable step %s", step.Name)
			}
			_, err = utilities.RunBashScript(script.String(), step.Scope.scope)
		default:
			log.Fatalf("Builtin value not found: %s \n", step.Name)
		}
		if err != nil {
			log.Fatalln(err)
		}
	}
}

//func (wtr *RunspecRunnable) runPipeline(pipelines []*ozoneConfig.Runnable, config *ozoneConfig.OzoneConfig, context string) {
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

// True means cache hit
func (wt *Runspec) checkCache(callstack *CallStack) bool {
	if wt.config.Headless == true || callstack.hasCaching() == false {
		return false
	}
	hash, err := callstack.getBuildHash(wt.OzoneWorkDir)
	if err != nil {
		log.Fatalln(err)
		return false
	}
	if hash == "" {
		return false
	}

	runnableName := callstack.RootRunnableName
	cachedHash := process_manager_client.CacheCheck(wt.OzoneWorkDir, runnableName)
	return cachedHash == hash
}

func (wt *Runspec) AddCallstacks(builds []*config.Runnable, config *config.OzoneConfig, context string) {
	ordinal := 0

	topLevelScope := CopyOrCreateNew(config.BuildVars)
	topLevelScope.SelfRender()
	topLevelScope.AddVariable(NewStringVariable("CONTEXT", context), ordinal)
	topLevelScope.AddVariable(NewStringVariable("OZONE_WORKING_DIR", wt.OzoneWorkDir), ordinal)

	for _, b := range builds {
		asOutput := make(map[string]string)
		err := wt.addCallstack(b, ordinal, context, config, CopyOrCreateNew(topLevelScope), asOutput)
		if err != nil {
			log.Fatalf("Error %s in runnable %s", err, b.Name)
		}
	}
}

func (wt *Runspec) addCallstack(rootConfigRunnable *config.Runnable, ordinal int, context string, ozoneConfig *config.OzoneConfig, buildScope *VariableMap, asOutput map[string]string) error {
	var allSourceFiles []string
	var runspecRunnables []*RunspecRunnable

	configRunnableDeque := lane.NewDeque[*ConfigRunnableStackItem]()
	configRunnableDeque.Append(NewConfigRunnableStackItem(rootConfigRunnable, buildScope, buildScope))

	for configRunnableDeque.Empty() == false {
		ordinal++
		configRunnableStackItem, _ := configRunnableDeque.Shift()
		configRunnable := configRunnableStackItem.ConfigRunnable

		runspecRunnable, err := wt.ConvertConfigRunnableStackItemToRunspecRunnable(configRunnableStackItem, ordinal)
		if err != nil {
			return err
		}
		allSourceFiles = append(allSourceFiles, runspecRunnable.SourceFiles...)

		runspecRunnables = append(runspecRunnables, runspecRunnable)

		// Prepend in reverse order so that the first dependency is top of the stack.
		for i := len(configRunnable.Depends) - 1; i >= 0; i-- {
			dependency := configRunnable.Depends[i]
			exists, dependencyRunnable := wt.config.FetchRunnable(dependency.Name)

			if !exists {
				log.Fatalf("Dependency %s on build %s doesn't exist", dependency.Name, runspecRunnable.Name)
			}

			parentScope := CopyOrCreateNew(runspecRunnable.BuildScope.GetScope())
			dependencyScope := CopyOrCreateNew(parentScope)
			//dependencyScope.MergeVariableMaps(contextEnvVars) TODO think this happens now in ConvertConfigRunnableStackItemToRunspecRunnable
			dependencyWithVars := CopyOrCreateNew(dependency.WithVars)
			dependencyWithVars.IncrementOrdinal(ordinal) // TODO test this
			dependencyScope.MergeVariableMaps(dependencyWithVars)
			var err error

			//dependencyVarAsOutput := copyMapStringString(dependency.VarOutputAs)
			//mergeMapStringString(dependencyVarAsOutput, asOutput)
			//dependencyScope.MergeVariableMaps(outputVars) TODO I don't think steps should inherit the env

			configRunnableDeque.Prepend(NewConfigRunnableStackItem(dependencyRunnable, parentScope, dependencyScope))

			if err != nil {
				return err
			}
			//outputVars.MergeVariableMaps(outputVarsFromDependentStep) TODO I don't think steps should inherit the env
			//fullEnvFromDependentStep.MergeVariableMaps(envFromDependentStep)
		}

		// TODO prepend queue with all configRunnableChildren
	}

	callStack := &CallStack{
		RootRunnableName: rootConfigRunnable.Name,
		RootRunnableType: rootConfigRunnable.Type,
		Hash:             "", // All source files + system environment variables + build vars
		Runnables:        runspecRunnables,
		SourceFiles:      allSourceFiles,
	}
	wt.CallStacks[callStack.RootRunnableType] = append(wt.CallStacks[callStack.RootRunnableType], callStack)

	return nil
}

func (wt *Runspec) PrintRunspec() {
	log.SetFlags(log.Flags() &^ (log.Ldate | log.Ltime))

	yamlEncoder := yaml.NewEncoder(os.Stdout)
	yamlEncoder.SetIndent(2)
	yamlEncoder.Encode(&wt)
}
