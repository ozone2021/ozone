package worktree

import (
	"github.com/oleiade/lane/v2"
	"github.com/ozone2021/ozone/ozone-lib/config"
	"github.com/ozone2021/ozone/ozone-lib/config/config_utils"
	. "github.com/ozone2021/ozone/ozone-lib/config/config_variable"
	"gopkg.in/yaml.v3"
	"log"
	"os"
	"path/filepath"
)

type WorktreeStep struct {
	Steps   []*config.Step `yaml:"steps"`
	WithEnv []string       `yaml:"with_env"`
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
	Type         config.RunnableType   `yaml:"RunnableType"`
}

type CallStack struct {
	Hash        string              `yaml:"hash"`
	SourceFiles []string            `yaml:"source_files"`
	Runnables   []*WorktreeRunnable `yaml:"runnables"`
}

type Worktree struct {
	config       *config.OzoneConfig
	ProjectName  string       `yaml:"project"`
	Context      string       `yaml:"context"`
	OzoneWorkDir string       `yaml:"work_dir"`
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

	workTreeRunnable := &WorktreeRunnable{
		Name:         configRunnable.Name,
		Ordinal:      ordinal,
		Service:      configRunnable.Service,
		SourceFiles:  wt.getConfigRunnableSourceFiles(configRunnable, buildScope),
		Dir:          configRunnable.Dir,
		BuildScope:   diffBuildScope,
		Conditionals: ConvertContextConditional(runnableBuildScope, configRunnable, wt.Context),
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

func (wt *Worktree) AddCallstacks(builds []*config.Runnable, config *config.OzoneConfig, context string) {
	ordinal := 0

	topLevelScope := CopyOrCreateNew(config.BuildVars)
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
		Hash:        "", // All source files + system environment variables + build vars
		Runnables:   worktreeRunnables,
		SourceFiles: allSourceFiles,
	}
	wt.CallStacks = append(wt.CallStacks, callStack)

	return nil
}

func mergeMapStringString(map1, map2 map[string]string) {
	for k, v := range map2 {
		map1[k] = v
	}
}

func copyMapStringString(m map[string]string) map[string]string {
	newMap := make(map[string]string)
	for k, v := range m {
		newMap[k] = v
	}
	return newMap
}

func (wt *Worktree) PrintWorktree() {
	log.SetFlags(log.Flags() &^ (log.Ldate | log.Ltime))

	yamlEncoder := yaml.NewEncoder(os.Stdout)
	yamlEncoder.SetIndent(2)
	yamlEncoder.Encode(&wt)
}
