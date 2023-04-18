package worktree

import (
	"errors"
	"fmt"
	"github.com/oleiade/lane/v2"
	"github.com/ozone2021/ozone/ozone-daemon-lib/cache"
	process_manager_client "github.com/ozone2021/ozone/ozone-daemon-lib/process-manager-client"
	"github.com/ozone2021/ozone/ozone-lib/buildables"
	"github.com/ozone2021/ozone/ozone-lib/config"
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

type WorktreeStep struct {
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

//	ContextConditionals []*ContextConditional `yaml:"context_conditionals"` # TODO save whether satisified
//  Steps is the depends and contextSteps merged
type WorktreeRunnable struct {
	Name         string                `yaml:"name"`
	Ordinal      int                   `yaml:"ordinal"`
	Service      string                `yaml:"service"`
	SourceFiles  []string              `yaml:"source_files"`
	Dir          string                `yaml:"dir"`
	BuildScope   *DifferentialScope    `yaml:"scope"`
	Conditionals *WorktreeConditionals `yaml:"conditionals"`
	Steps        []*WorktreeStep       `yaml:"steps"`
	Type         config.RunnableType   `yaml:"RunnableType"`
}

type CallStack struct {
	RootRunnableName string              `yaml:"root_runnable_name"`
	RootRunnableType config.RunnableType `yaml:"root_runnable_type"`
	Hash             string              `yaml:"hash"`
	SourceFiles      []string            `yaml:"source_files"`
	Runnables        []*WorktreeRunnable `yaml:"runnables"`
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

type Worktree struct {
	config       *config.OzoneConfig
	ProjectName  string       `yaml:"project"`
	Context      string       `yaml:"context"`
	OzoneWorkDir string       `yaml:"work_dir"` // move into config
	BuildVars    *VariableMap `yaml:"build_vars"`
	CallStacks   []*CallStack `yaml:"call_stack"`
}

func NewWorktree(context, ozoneWorkingDir string, config *config.OzoneConfig) *Worktree {
	systemEnvVars := OSEnvToVarsMap()

	renderedBuildVars := config.BuildVars
	renderedBuildVars.RenderNoMerge(systemEnvVars)
	renderedBuildVars.SelfRender()

	worktree := &Worktree{
		config:       config,
		ProjectName:  config.ProjectName,
		Context:      context,
		OzoneWorkDir: ozoneWorkingDir,
		BuildVars:    renderedBuildVars,
	}

	return worktree
}

func (wt *Worktree) getConfigRunnableSourceFiles(configRunnable *config.Runnable, buildScope *VariableMap) []string {
	var sourceFiles []string
	for _, file := range configRunnable.SourceFiles {
		rendered, err := buildScope.RenderSentence(file)
		if err != nil {
			log.Fatalf("Error: %s while getting source file: %s in configRunnable: %s", err, file, configRunnable.Name)
		}
		sourceFiles = append(sourceFiles, filepath.Join(wt.OzoneWorkDir, rendered))
	}
	buildScope.SelfRender()

	return sourceFiles
}

func (wt *Worktree) FetchContextEnvs(ordinal int, buildScope *VariableMap, runnable *config.Runnable) (*VariableMap, error) {
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

func (wt *Worktree) ContextStepsFlatten(configRunnable *config.Runnable, buildscope *VariableMap, ordinal int) ([]*WorktreeStep, error) {
	var steps []*WorktreeStep
	for _, cs := range configRunnable.ContextSteps {
		match, err := config_utils.ContextInPattern(wt.Context, cs.Context, buildscope)
		if err != nil || !match {
			return nil, err
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

			steps = append(steps, &WorktreeStep{
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

func (wt *Worktree) ConvertConfigRunnableStackItemToWorktreeRunnable(configRunnableStackItem *ConfigRunnableStackItem, ordinal int) (*WorktreeRunnable, error) {
	configRunnable := configRunnableStackItem.ConfigRunnable
	buildScope := configRunnableStackItem.buildScope
	parentScope := configRunnableStackItem.parentScope

	addCallstackScopeVars(configRunnable, buildScope, ordinal)

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

	workTreeRunnable := &WorktreeRunnable{
		Name:         configRunnable.Name,
		Ordinal:      ordinal,
		Service:      configRunnable.Service,
		SourceFiles:  wt.getConfigRunnableSourceFiles(configRunnable, buildScope),
		Dir:          configRunnable.Dir,
		BuildScope:   diffBuildScope,
		Conditionals: ConvertContextConditional(runnableBuildScope, configRunnable, wt.Context),
		Steps:        steps,
		Type:         configRunnable.Type,
	}

	return workTreeRunnable, nil
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

func addCallstackScopeVars(runnable *config.Runnable, buildScope *VariableMap, ordinal int) {
	if runnable.Service != "" {
		buildScope.AddVariableWithoutOrdinality(NewStringVariable("SERVICE", runnable.Service))
	}
	if runnable.Dir != "" {
		buildScope.AddVariable(NewStringVariable("DIR", runnable.Dir), ordinal)
	}
	buildScope.AddVariable(NewStringVariable("NAME", runnable.Name), ordinal)
}

func (wt *Worktree) ExecuteCallstacks() error {
	for _, callstack := range wt.CallStacks {
		if callstack.RootRunnableType == config.BuildType && wt.checkCache(callstack) == true {
			log.Printf("Info: build files for %s unchanged from cache. \n", callstack.RootRunnableName)
			continue
		}
		for _, worktreeRunnable := range callstack.Runnables {
			worktreeRunnable.RunSteps()
		}
	}
	return nil
}

func (wtr *WorktreeRunnable) RunSteps() {
	for _, step := range wtr.Steps {
		step.RunStep(wtr.Type)
	}
}

func (step *WorktreeStep) RunStep(runnableType config.RunnableType) {
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

func (step *WorktreeStep) runBuildable() {
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

func (step *WorktreeStep) runTestable() {
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

func (step *WorktreeStep) runUtility() {
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

func (step *WorktreeStep) runDeployables() {
	serviceVar, ok := step.Scope.scope.GetVariable("SERVICE")
	if !ok {
		log.Fatalf("SERVICE not set for runnable step %s", step.Name)
	}
	service := serviceVar.GetStringValue()
	if step.Type == "builtin" {
		var err error
		switch step.Name {
		case "executable":
			err = executable.Build(service, step.Scope.scope)
		case "helm":
			err = helm.Deploy(service, step.Scope.scope)
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

//func (wtr *WorktreeRunnable) runPipeline(pipelines []*ozoneConfig.Runnable, config *ozoneConfig.OzoneConfig, context string) {
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
func (wt *Worktree) checkCache(callstack *CallStack) bool {
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

func (wt *Worktree) AddCallstacks(builds []*config.Runnable, config *config.OzoneConfig, context string) {
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

func (wt *Worktree) addCallstack(rootConfigRunnable *config.Runnable, ordinal int, context string, ozoneConfig *config.OzoneConfig, buildScope *VariableMap, asOutput map[string]string) error {
	var allSourceFiles []string
	var worktreeRunnables []*WorktreeRunnable

	configRunnableDeque := lane.NewDeque[*ConfigRunnableStackItem]()
	configRunnableDeque.Append(NewConfigRunnableStackItem(rootConfigRunnable, buildScope, buildScope))

	for configRunnableDeque.Empty() == false {
		ordinal++
		configRunnableStackItem, _ := configRunnableDeque.Shift()
		configRunnable := configRunnableStackItem.ConfigRunnable

		worktreeRunnable, err := wt.ConvertConfigRunnableStackItemToWorktreeRunnable(configRunnableStackItem, ordinal)
		if err != nil {
			return err
		}
		allSourceFiles = append(allSourceFiles, worktreeRunnable.SourceFiles...)

		worktreeRunnables = append(worktreeRunnables, worktreeRunnable)

		for _, dependency := range configRunnable.Depends {
			exists, dependencyRunnable := wt.config.FetchRunnable(dependency.Name)

			if !exists {
				log.Fatalf("Dependency %s on build %s doesn't exist", dependency.Name, worktreeRunnable.Name)
			}

			parentScope := CopyOrCreateNew(worktreeRunnable.BuildScope.GetScope())
			dependencyScope := CopyOrCreateNew(parentScope)
			//dependencyScope.MergeVariableMaps(contextEnvVars) TODO think this happens now in ConvertConfigRunnableStackItemToWorktreeRunnable
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
		Runnables:        worktreeRunnables,
		SourceFiles:      allSourceFiles,
	}
	wt.CallStacks = append(wt.CallStacks, callStack)

	return nil
}

func (wt *Worktree) PrintWorktree() {
	log.SetFlags(log.Flags() &^ (log.Ldate | log.Ltime))

	yamlEncoder := yaml.NewEncoder(os.Stdout)
	yamlEncoder.SetIndent(2)
	yamlEncoder.Encode(&wt)
}
