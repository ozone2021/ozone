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
	"github.com/ozone2021/ozone/ozone-lib/logger_lib"
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
	Children     []Node
	Name         string               `yaml:"name"`
	Cache        bool                 `yaml:"cache"`
	Ordinal      int                  `yaml:"ordinal"`
	Service      string               `yaml:"service"`
	SourceFiles  []string             `yaml:"source_files"`
	Dir          string               `yaml:"dir"`
	BuildScope   *DifferentialScope   `yaml:"scope"`
	Conditionals *RunspecConditionals `yaml:"conditionals"`
	Steps        []*RunspecStep       `yaml:"steps"`
	Type         config.RunnableType  `yaml:"RunnableType"`
}

func (r *RunspecRunnable) GetRunnable() *RunspecRunnable {
	return r
}

type CallStack struct {
	RootRunnableName string              `yaml:"root_runnable_name"`
	RootRunnableType config.RunnableType `yaml:"root_runnable_type"`
	HasCaching       bool                `yaml:"has_caching"`
	Hash             string              `yaml:"hash"`
	SourceFiles      []string            `yaml:"source_files"`
	RootRunnable     *RunspecRunnable
}

func (c *CallStack) GetType() string {
	return "callstack"
}

func (c *CallStack) GetRunnable() *RunspecRunnable {
	return c.RootRunnable
}

func (c *CallStack) getSourceFiles() []string {
	return c.SourceFiles
}

func (r *RunspecRunnable) getSourceFiles() []string {
	return r.SourceFiles
}

func (r *RunspecRunnable) GetType() string {
	return "runspecRunnable"
}

type Node interface {
	GetRunnable() *RunspecRunnable
	GetType() string
	ConditionalsSatisfied() bool
	getSourceFiles() []string
	hasCaching() bool
}

func (cs *CallStack) hasCaching() bool {
	if cs.RootRunnableType == config.BuildType && cs.HasCaching == true {
		return true
	}
	return false
}

func (cs *CallStack) ConditionalsSatisfied() bool {
	return cs.RootRunnable.Conditionals.Satisfied
}

func (r RunspecRunnable) ConditionalsSatisfied() bool {
	return r.Conditionals.Satisfied
}

func getBuildHash(node Node, ozoneWorkingDir string) (string, error) {
	ozonefilePath := path.Join(ozoneWorkingDir, "Ozonefile")

	ozonefileEditTime, err := cache.FileLastEdit(ozonefilePath)

	if err != nil {
		return "", err
	}

	filesDirsLastEditTimes := []int64{ozonefileEditTime}

	for _, filePath := range node.getSourceFiles() {
		wildcardExpandedFiles, err := filepath.Glob(filePath)

		if err != nil {
			fmt.Println(err)
		}

		for _, match := range wildcardExpandedFiles {
			editTime, err := cache.FileLastEdit(match)

			if err != nil {
				return "", errors.New(fmt.Sprintf("Source file %s for runnable %s is missing.", filePath, node.GetRunnable().Name))
			}

			filesDirsLastEditTimes = append(filesDirsLastEditTimes, editTime)
		}

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
	if !runnable.DropContextEnv {
		fetchedEnvs, err := wt.config.FetchEnvs(ordinal, []string{wt.Context}, buildScope)
		if err != nil {
			return nil, err
		}
		contextEnvVars.MergeVariableMaps(fetchedEnvs)
	}
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

func (wt *Runspec) ContextStepsFlatten(configRunnable *config.Runnable, scope *VariableMap, ordinal int) ([]*RunspecStep, error) {
	var steps []*RunspecStep
	for _, cs := range configRunnable.ContextSteps {
		match, err := config_utils.ContextInPattern(wt.Context, cs.Context, scope)
		if err != nil {
			log.Fatalln("Err in ContextStep context %s in runnable %s, err: %s", cs.Context, configRunnable.Name, err)
		}
		if !match {
			continue
		}

		contextStepVars, err := wt.config.FetchEnvs(ordinal, cs.WithEnv, scope)
		contextStepVars.MergeVariableMaps(scope)
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

			steps = append(steps, &RunspecStep{
				Type: step.Type,
				Name: step.Name,
				Scope: &DifferentialScope{
					parentScope: scope,
					scope:       stepVars,
				},
				VarOutputAs: nil,
			})
		}
	}

	return steps, nil
}

func (wt *Runspec) StepsToRunspecSteps(configRunnable *config.Runnable, buildscope *VariableMap, ordinal int) ([]*RunspecStep, error) {
	var steps []*RunspecStep
	for _, step := range configRunnable.Steps {
		stepVars := CopyOrCreateNew(step.WithVars)
		stepVars.IncrementOrdinal(ordinal) // TODO should be part of Copy/CreateNew
		stepVars.MergeVariableMaps(buildscope)

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

	return steps, nil
}

func (wt *Runspec) ConvertConfigRunnableStackItemToRunspecRunnable(configRunnableStackItem *ConfigRunnableStackItem, ordinal int, logger *logger_lib.Logger) (*RunspecRunnable, error) {
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
	runnableBuildScope.MergeVariableMaps(configRunnable.WithVars)
	diffBuildScope := NewDifferentialScope(parentScope, runnableBuildScope)

	contextSteps, err := wt.ContextStepsFlatten(configRunnable, runnableBuildScope, ordinal)
	if err != nil {
		return nil, err
	}

	steps, err := wt.StepsToRunspecSteps(configRunnable, buildScope, ordinal)
	if err != nil {
		return nil, err
	}

	combinedSteps := append(contextSteps, steps...)

	runspecRunnable := &RunspecRunnable{
		Name:         configRunnable.Name,
		Ordinal:      ordinal,
		Children:     []Node{},
		Cache:        configRunnable.Cache,
		Service:      service,
		SourceFiles:  sourceFiles,
		Dir:          dir,
		BuildScope:   diffBuildScope,
		Conditionals: ConvertContextConditional(runnableBuildScope, configRunnable, wt.Context, logger),
		Steps:        combinedSteps,
		Type:         configRunnable.Type,
	}

	return runspecRunnable, nil
}

type ConfigRunnableStackItem struct {
	Parent         *RunspecRunnable
	ConfigRunnable *config.Runnable
	parentScope    *VariableMap
	buildScope     *VariableMap
}

func NewConfigRunnableStackItem(parent *RunspecRunnable, configRunnable *config.Runnable, parentScope *VariableMap, buildScope *VariableMap) *ConfigRunnableStackItem {
	return &ConfigRunnableStackItem{
		Parent:         parent,
		ConfigRunnable: configRunnable,
		parentScope:    parentScope,
		buildScope:     buildScope,
	}
}

func (r *RunspecRunnable) addChild(child Node) {
	r.Children = append(r.Children, child)
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

func (wt *Runspec) ExecuteCallstacks() *RunResult {
	runOrder := []config.RunnableType{
		config.PreUtilityType,
		config.BuildType,
		config.DeployType,
		config.TestType,
		config.PipelineType,
		config.PostUtilityType,
	}
	runResult := NewRunResult()

	for _, runnableType := range runOrder {
		for _, callstack := range wt.CallStacks[runnableType] {
			callstackLogger, err := logger_lib.New(wt.OzoneWorkDir, callstack.RootRunnableName, wt.config.Headless)
			if err != nil {
				log.Fatalln(err)
			}
			runResult.AddRootCallstack(callstack, callstackLogger)
			callstackResults, err := wt.CheckCacheAndExecute(callstack, callstackLogger)
			err = runResult.AddCallstackResult(callstack.RootRunnableName, callstackResults, err)
			if err != nil {
				log.Fatalln("Error: %s", err)
			}
		}
	}

	return runResult
}

func (wt *Runspec) CheckCacheAndExecute(rootCallstack *CallStack, logger *logger_lib.Logger) ([]*CallstackResult, error) {
	nodeInputStack := lane.NewStack[Node]()
	nodeInputStack.Push(rootCallstack)

	workQueue := lane.NewDeque[*RunspecRunnable]()

	var results []*CallstackResult

	for nodeInputStack.Size() != 0 {
		node, ok := nodeInputStack.Pop()
		if !ok {
			log.Fatalf("Error: runnable stack is empty. \n")
		}
		cached := false
		hash := ""

		if node.hasCaching() && wt.config.Headless == false { // TODO do only callstacks have caching?
			cached, hash = wt.checkNodeCache(node)

			if node.ConditionalsSatisfied() == true && cached == true {
				logger.Println("--------------------")
				logger.Printf("Cache Info: build files for %s %s unchanged from cache. \n", node.GetType(), node.GetRunnable().Name)
				logger.Println("--------------------")
				results = append(results, NewCachedCallstackResult(node.GetRunnable().Name, nil))
				continue
			}
		}

		switch node.(type) {
		case *CallStack:
			callstack, _ := node.(*CallStack)
			callstackLogger, err := logger_lib.New(wt.OzoneWorkDir, callstack.RootRunnableName, wt.config.Headless)
			if err != nil {
				log.Fatalln(err)
			}
			if callstack.ConditionalsSatisfied() == false {
				log.Printf("Skipping callstack %s because conditionals not satisfied \n", callstack.RootRunnableName)
				results = append(results, NewSucceededCallstackResult(callstack.RootRunnable.Name, callstackLogger))
				continue
			}
			log.Printf("Executing callstack %s \n", callstack.RootRunnableName)
			err = callstack.execute(wt.config.Headless, callstackLogger) // TODO call recursive
			if err != nil {
				results = append(results, NewFailedCallstackResult(callstack.RootRunnable.Name, err, callstackLogger))
				if wt.config.Headless {
					logger.Fatalf("Error: %s in runnable \n", err, callstack.RootRunnable.Name)
				}
			} else {
				results = append(results, NewSucceededCallstackResult(callstack.RootRunnable.Name, callstackLogger))
			}

			if err == nil && callstack.RootRunnable.hasCaching() && wt.config.Headless == false {
				process_manager_client.CacheUpdate(wt.OzoneWorkDir, callstack.RootRunnable.Name, hash) // TODO
			}
		case *RunspecRunnable:
			runspecRunnable, _ := node.(*RunspecRunnable)
			workQueue.Prepend(runspecRunnable)
			for i := len(runspecRunnable.Children) - 1; i >= 0; i-- {
				nodeInputStack.Push(runspecRunnable.Children[i].GetRunnable())
			}
		default:
			logger.Fatalln("Error: node is not a CallStack or Runnable")
		}
	}

	workQueueResults, err := executeWorkQueue(false, wt.config.Headless, logger, workQueue)
	results = append(results, workQueueResults...)

	return results, err
}

func (cs *CallStack) execute(headless bool, logger *logger_lib.Logger) error {
	inStack := lane.NewDeque[*RunspecRunnable]()

	inStack.Prepend(cs.RootRunnable)

	log.Printf("Executing callstack: %s \n", cs.RootRunnable.Name)

	workQueue := lane.NewDeque[*RunspecRunnable]()

	for inStack.Size() != 0 {
		current, ok := inStack.Shift()
		if !ok {
			logger.Fatalf("Error: runnable work stack is empty. \n")
		}

		workQueue.Prepend(current)
		for _, child := range current.Children {
			inStack.Prepend(child.GetRunnable())
		}
	}

	_, err := executeWorkQueue(true, headless, logger, workQueue)
	if err != nil {
		return err
	}
	return nil
}

func executeWorkQueue(returnOnErr, headless bool, logger *logger_lib.Logger, workQueue *lane.Deque[*RunspecRunnable]) ([]*CallstackResult, error) {
	var results []*CallstackResult

	for workQueue.Size() != 0 {
		runspecRunnable, ok := workQueue.Shift()
		if !ok {
			logger.Fatalf("Error: runnable work queue is empty. \n")
		}
		log.Printf("  - %s \n", runspecRunnable.Name)

		if runspecRunnable.ConditionalsSatisfied() == false {
			log.Printf("Skipping runnable %s because conditionals not satisfied \n", runspecRunnable.Name)
			continue
		}

		err := runspecRunnable.RunSteps(logger)
		if err != nil {
			if headless {
				logger.Fatalf("Error in step: %s in runnable \n", err, runspecRunnable.Name)
			}
			if returnOnErr {
				results = append(results, NewFailedCallstackResult(runspecRunnable.Name, err, logger))
				return results, err
			}
		}
	}
	return results, nil
}

func (wtr *RunspecRunnable) hasCaching() bool {
	return wtr.Cache
}

func (wtr *RunspecRunnable) RunSteps(logger *logger_lib.Logger) error {
	for _, step := range wtr.Steps {
		step.Scope.scope.MergeVariableMaps(wtr.BuildScope.scope)
		logger.Printf("Running step %s \n", step.Name)
		err := step.RunStep(wtr.Type, logger)
		if err != nil {
			return err
		}
	}
	return nil
}

func (step *RunspecStep) RunStep(runnableType config.RunnableType, logger *logger_lib.Logger) error {
	if step.Type == "builtin" {
		switch runnableType {
		case config.PreUtilityType:
			return step.runUtility(logger)
		case config.BuildType:
			return step.runBuildable(logger)
		case config.DeployType:
			return step.runDeployables(logger)
		case config.TestType:
			return step.runTestable(logger)
		case config.PostUtilityType:
			return step.runUtility(logger)
		}
	}
	return nil
}

func (step *RunspecStep) runBuildable(logger *logger_lib.Logger) error {
	switch step.Name {
	case "buildDockerImage":
		fmt.Println("Building docker image.")
		return buildables.BuildDockerContainer(step.Scope.scope, logger)
	case "bashScript":
		script, ok := step.Scope.scope.GetVariable("SCRIPT")
		if !ok {
			return errors.New(fmt.Sprintf("Script not set for runnable step %s", "r.Name"))
		}
		_, err := utilities.RunBashScript(script.String(), step.Scope.scope, logger)
		return err
	case "pushDockerImage":
		fmt.Println("Building docker image.")
		return buildables.PushDockerImage(step.Scope.scope, logger)
	case "tagDockerImageAs":
		fmt.Println("Tagging docker image.")
		return buildables.TagDockerImageAs(step.Scope.scope, logger)
	}
	return nil
}

func (step *RunspecStep) runTestable(logger *logger_lib.Logger) error {
	switch step.Name {
	case "bashScript":
		script, ok := step.Scope.scope.GetVariable("SCRIPT")
		if !ok {
			return errors.New(fmt.Sprintf("Script not set for runnable step %s", step.Name))
		}
		_, err := utilities.RunBashScript(script.String(), step.Scope.scope, logger)
		if err != nil {
			logger.Fatalln(err)
		}
	default:
		return errors.New(fmt.Sprintf("Testable value not found: %s \n", step.Name))
	}
	return nil
}

func (step *RunspecStep) runUtility(logger *logger_lib.Logger) error {
	switch step.Name {
	case "bashScript":
		script, ok := step.Scope.scope.GetVariable("SCRIPT")
		if !ok {
			return errors.New(fmt.Sprintf("Script not set for runnable step %s", step.Name)) // TODO wrap err
		}
		_, err := utilities.RunBashScript(script.String(), step.Scope.scope, logger)
		return err
	default:
		return errors.New(fmt.Sprintf("Utility value not found: %s \n", step.Name))
	}
	return nil
}

func (step *RunspecStep) runDeployables(logger *logger_lib.Logger) error {
	if step.Type == "builtin" {
		switch step.Name {
		case "executable":
			return executable.Build(step.Scope.scope, logger)
		case "helm":
			return helm.Deploy(step.Scope.scope, logger)
		case "runDockerImage":
			return docker.Build(step.Scope.scope, logger)
		case "bashScript":
			script, ok := step.Scope.scope.GetVariable("SCRIPT")
			if !ok {
				errors.New(fmt.Sprintf("Script not set for runnable step %s", step.Name))
			}
			_, err := utilities.RunBashScript(script.String(), step.Scope.scope, logger)
			return err
		default:
			return errors.New(fmt.Sprintf("Builtin value not found: %s \n", step.Name))
		}
	}
	return nil
}

//func (wtr *RunspecRunnable) runPipeline(pipelines []*ozoneConfig.Runnable, config *ozoneConfig.OzoneConfig, context string) {
//	for _, pipeline := range pipelines {
//		var runnables []*ozoneConfig.Runnable
//		for _, dependency := range pipeline.Depends {
//			exists, dependencyRunnable := config.FetchRunnable(dependency.Name)
//			if !exists {
//
//				log.Fatalf("Dependency %s on pipeline %s doesn't exist", dependency.Name, pipeline.Name)
//			}
//			runnables = append(runnables, dependencyRunnable)
//		}
//		run(runnables, config, context)
//	}
//}

// True means cache hit
func (wt *Runspec) checkNodeCache(node Node) (bool, string) {
	if wt.config.Headless == true || node.hasCaching() == false {
		return false, ""
	}
	hash, err := getBuildHash(node, wt.OzoneWorkDir)
	if err != nil {
		log.Fatalf("CheckNodeCache error: %s", err)
		return false, ""
	}
	if hash == "" {
		return false, ""
	}

	cachedHash := process_manager_client.CacheCheck(wt.OzoneWorkDir, node.GetRunnable().Name)
	return cachedHash == hash, hash
}

func (wt *Runspec) AddCallstacks(runnables []*config.Runnable, ozoneConfig *config.OzoneConfig, context string) {
	ordinal := 0

	topLevelScope := CopyOrCreateNew(ozoneConfig.BuildVars)
	topLevelScope.SelfRender()
	topLevelScope.AddVariable(NewStringVariable("CONTEXT", context), ordinal)
	topLevelScope.AddVariable(NewStringVariable("OZONE_WORKING_DIR", wt.OzoneWorkDir), ordinal)

	for _, r := range runnables {
		asOutput := make(map[string]string)
		var err error
		var callstack *CallStack

		// Treat pipelines as a special case, each runnable dependency of the pipeline is an invidiual callstack.
		if r.Type == config.PipelineType {
			for _, dependency := range r.Depends {
				exists, dependencyRunnable := wt.config.FetchRunnable(dependency.Name)
				if !exists {
					log.Fatalf("Dependency %s on build %s doesn't exist", dependency.Name, r.Name)
				}
				callstack, err = wt.addCallstack(dependencyRunnable, ordinal, CopyOrCreateNew(topLevelScope), asOutput)
				wt.CallStacks[callstack.RootRunnableType] = append(wt.CallStacks[callstack.RootRunnableType], callstack)
			}
		} else {
			callstack, err = wt.addCallstack(r, ordinal, CopyOrCreateNew(topLevelScope), asOutput)
			wt.CallStacks[callstack.RootRunnableType] = append(wt.CallStacks[callstack.RootRunnableType], callstack)
		}
		if err != nil {
			log.Fatalf("Error %s in runnable %s", err, r.Name)
		}

		// TODO close callstack logger.
	}
}

func (wt *Runspec) addCallstack(rootConfigRunnable *config.Runnable, ordinal int, buildScope *VariableMap, asOutput map[string]string) (*CallStack, error) {
	var allSourceFiles []string
	var runspecRunnables []*RunspecRunnable
	var rootRunable *RunspecRunnable

	logger, err := logger_lib.New(wt.OzoneWorkDir, rootConfigRunnable.Name, wt.config.Headless)
	if err != nil {
		log.Fatalln(err)
	}

	dependencyRunnableDeque := lane.NewDeque[*ConfigRunnableStackItem]()
	dependencyRunnableDeque.Append(NewConfigRunnableStackItem(nil, rootConfigRunnable, buildScope, buildScope))

	for dependencyRunnableDeque.Empty() == false {
		ordinal++
		configRunnableStackItem, _ := dependencyRunnableDeque.Shift()
		configRunnable := configRunnableStackItem.ConfigRunnable

		if configRunnable.Cache == true && configRunnableStackItem.Parent != nil {
			childCallstack, err := wt.addCallstack(configRunnable, ordinal, CopyOrCreateNew(configRunnableStackItem.buildScope), asOutput)
			if err != nil {
				return nil, err
			}
			configRunnableStackItem.Parent.addChild(childCallstack)
			continue
		}
		runspecRunnable, err := wt.ConvertConfigRunnableStackItemToRunspecRunnable(configRunnableStackItem, ordinal, logger)
		if err != nil {
			return nil, err
		}
		parent := configRunnableStackItem.Parent
		if parent == nil {
			rootRunable = runspecRunnable
		} else {
			parent.Children = append(parent.Children, runspecRunnable)
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

			dependencyRunnableDeque.Prepend(NewConfigRunnableStackItem(runspecRunnable, dependencyRunnable, parentScope, dependencyScope))

			if err != nil {
				return nil, err
			}
			//outputVars.MergeVariableMaps(outputVarsFromDependentStep) TODO I don't think steps should inherit the env
			//fullEnvFromDependentStep.MergeVariableMaps(envFromDependentStep)
		}

		// TODO prepend queue with all configRunnableChildren
	}

	callStack := &CallStack{
		RootRunnableName: rootConfigRunnable.Name,
		RootRunnableType: rootConfigRunnable.Type,
		HasCaching:       rootConfigRunnable.Cache,
		Hash:             "", // All source files + system environment variables + build vars
		RootRunnable:     rootRunable,
		SourceFiles:      allSourceFiles,
	}

	err = logger.Close()
	if err != nil {
		return nil, err
	}

	return callStack, nil
}

func (wt *Runspec) PrintRunspec() {
	log.SetFlags(log.Flags() &^ (log.Ldate | log.Ltime))

	yamlEncoder := yaml.NewEncoder(os.Stdout)
	yamlEncoder.SetIndent(2)
	yamlEncoder.Encode(&wt)
}
