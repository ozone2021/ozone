package config

import (
	"errors"
	"fmt"
	. "github.com/ozone2021/ozone/ozone-lib/config/config_variable"
	"github.com/ozone2021/ozone/ozone-lib/env"
	"github.com/ozone2021/ozone/ozone-lib/env/git_env"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
)

type RunnableType int

const (
	PreUtilityType  RunnableType = iota
	BuildType       RunnableType = iota
	DeployType      RunnableType = iota
	TestType        RunnableType = iota
	PostUtilityType RunnableType = iota
)

type ContextInfo struct {
	Default string   `yaml:"default"`
	List    []string `yaml:"list"`
}

type Include struct {
	Name     string       `yaml:"name"`
	WithVars *VariableMap `yaml:"with_vars"`
	Type     string       `yaml:"type"`
}

type Environment struct {
	Name     string       `yaml:"name"`
	WithVars *VariableMap `yaml:"with_vars"`
	Includes []*Include   `yaml:"include"`
}

type Step struct {
	Type        string            `yaml:"type"`
	Name        string            `yaml:"name"`
	WithVars    *VariableMap      `yaml:"with_vars"`
	VarOutputAs map[string]string `yaml:"var_output_as"`
}

type ContextStep struct {
	Context string   `yaml:"context"`
	Steps   []*Step  `yaml:"steps"`
	WithEnv []string `yaml:"with_env"`
}

type ContextEnv struct {
	Context  string       `yaml:"context"`
	WithEnv  []string     `yaml:"with_env"`
	WithVars *VariableMap `yaml:"with_vars"`
}

type ContextConditional struct {
	Context string `yaml:"context"`
	//WhenChanged
	WhenScript    []string `yaml:"when_script"`
	WhenNotScript []string `yaml:"when_not_script"`
}

type Runnable struct {
	Name        string   `yaml:"name"`
	Service     string   `yaml:"service"`
	Dir         string   `yaml:"dir"`
	SourceFiles []string `yaml:"source_files"`
	Depends     []*Step  `yaml:"depends_on"`
	//WithEnv     	[]string      	`yaml:"with_env"`
	ContextEnv          []*ContextEnv         `yaml:"context_envs"`
	ContextConditionals []*ContextConditional `yaml:"context_conditionals"`
	ContextSteps        []*ContextStep        `yaml:"context_steps"`
	Type                RunnableType
}

type OzoneConfig struct {
	ProjectName   string         `yaml:"project"`
	ContextInfo   ContextInfo    `yaml:"context"`
	BuildVars     *VariableMap   `yaml:"build_vars"`
	Environments  []*Environment `yaml:"environments"`
	PreUtilities  []*Runnable    `yaml:"pre_utilities"`
	Builds        []*Runnable    `yaml:"builds"`
	Deploys       []*Runnable    `yaml:"deploys"`
	Tests         []*Runnable    `yaml:"tests"`
	PostUtilities []*Runnable    `yaml:"post_utilities"`
}

func (config *OzoneConfig) FetchRunnable(name string) (bool, *Runnable) {
	if has, runnable := config.HasPreUtility(name); has == true {
		return true, runnable
	}
	if has, runnable := config.HasBuild(name); has == true {
		return true, runnable
	}
	if has, runnable := config.HasDeploy(name); has == true {
		return true, runnable
	}
	if has, runnable := config.HasTest(name); has == true {
		return true, runnable
	}
	if has, runnable := config.HasPostUtility(name); has == true {
		return true, runnable
	}

	return false, nil
}

func (config *OzoneConfig) HasContext(name string) bool {
	for _, c := range config.ContextInfo.List {
		if name == c {
			return true
		}
	}
	return false
}

func (config *OzoneConfig) HasPreUtility(name string) (bool, *Runnable) {
	return config.ListHasRunnableOfType(name, config.PreUtilities, PreUtilityType)
}

func (config *OzoneConfig) HasBuild(name string) (bool, *Runnable) {
	return config.ListHasRunnableOfType(name, config.Builds, BuildType)
}

func (config *OzoneConfig) HasDeploy(name string) (bool, *Runnable) {
	return config.ListHasRunnableOfType(name, config.Deploys, DeployType)
}

func (config *OzoneConfig) HasTest(name string) (bool, *Runnable) {
	return config.ListHasRunnableOfType(name, config.Tests, TestType)
}

func (config *OzoneConfig) HasPostUtility(name string) (bool, *Runnable) {
	return config.ListHasRunnableOfType(name, config.PostUtilities, PostUtilityType)
}

func (config *OzoneConfig) DeploysHasService(service string) bool {
	for _, runnable := range config.Deploys {
		if runnable.Service == service {
			return true
		}
	}
	return false
}

func (config *OzoneConfig) ListHasRunnableOfType(name string, runnables []*Runnable, givenType RunnableType) (bool, *Runnable) {
	for _, runnable := range runnables {
		if runnable.Name == name && runnable.Type == givenType {
			return true, runnable
		}
	}
	return false, nil
}

func (config *OzoneConfig) fetchEnv(ordinal int, envName string, scopeMap *VariableMap) (*VariableMap, error) {
	// TODO ordinal
	nameFound := false
	varsMap := NewVariableMap()
	for _, e := range config.Environments {
		if e.Name == envName {
			nameFound = true
			if len(e.Includes) != 0 {
				for _, incl := range e.Includes {
					var inclVarsMap *VariableMap
					var err error
					if incl.Type == "builtin" {
						inclParamVarsMap := CopyOrCreateNew(incl.WithVars)
						inclParamVarsMap.MergeVariableMaps(scopeMap)
						inclVarsMap, err = config.fetchBuiltinEnvFromInclude(ordinal, incl.Name, inclParamVarsMap)
						if err != nil {
							return nil, err
						}
					} else {
						inclVarsMap, err = config.fetchEnv(ordinal, incl.Name, scopeMap)
						if err != nil {
							return nil, err
						}
					}

					inclVarsMap.RenderNoMerge(ordinal, scopeMap)
					varsMap.MergeVariableMaps(inclVarsMap)
				}
			}
			renderedEnvVars := CopyOrCreateNew(e.WithVars)
			renderedEnvVars.RenderNoMerge(ordinal, scopeMap)

			varsMap.MergeVariableMaps(renderedEnvVars)
		}
	}
	if nameFound == false {
		return nil, errors.New(fmt.Sprintf("Environment %s not found \n", envName))
	}

	return varsMap, nil
}

func (config *OzoneConfig) fetchBuiltinEnvFromInclude(ordinal int, envName string, varsMap *VariableMap) (*VariableMap, error) {
	var err error
	fromIncludeMap := NewVariableMap()

	// TODO
	switch envName {
	case "env/from_k8s_secret_file":
		err = env.FromSecretFile(ordinal, varsMap, fromIncludeMap)
	case "env/from_k8s_secret64":
		err = env.FromSecret64(ordinal, varsMap, fromIncludeMap)
	case "env/from_env_file":
		err = env.FromEnvFile(ordinal, varsMap, fromIncludeMap)
	case "env/from_version_file":
		err = env.FromVersionFile(ordinal, varsMap, fromIncludeMap)
	case "env/git_log_hash":
		err = git_env.GitLogHash(ordinal, varsMap, fromIncludeMap)
	case "env/git_directory_branch_hash":
		err = env.DynamicFromGitDirBranchNameHash(ordinal, varsMap, fromIncludeMap)
	case "env/git_directory_branch_static":
		err = env.StaticFromGitDirBranchNameHash(ordinal, varsMap, fromIncludeMap)
	case "env/git_submodule_commit_hash":
		err = env.GitSubmoduleHash(ordinal, varsMap, fromIncludeMap)
	}

	if err != nil {
		return nil, err
	}

	return fromIncludeMap, nil
}

func (config *OzoneConfig) FetchEnvs(ordinal int, envList []string, scope *VariableMap) (*VariableMap, error) {
	ordinal++
	varsMap := NewVariableMap()

	for _, env := range envList {
		renderedEnv, err := PongoRender(env, scope.ConvertMapPongo())
		if err != nil {
			return nil, err
		}

		fetchedMap, err := config.fetchEnv(ordinal, renderedEnv, scope)
		if err != nil {
			return nil, err
		}
		err = varsMap.MergeVariableMaps(fetchedMap)
		if err != nil {
			return nil, err
		}
		varsMap.MergeVariableMaps(fetchedMap)
	}
	varsMap.RenderNoMerge(ordinal, scope)
	return varsMap, nil
}

//func (r *Runnable) UnmarshalYAML(unmarshal func(interface{}) error) error {
//	var yamlObj map[string]interface{}
//	if err := unmarshal(&yamlObj); err != nil {
//		return err
//	}
//
//	r.Name = yamlObj["name"].(string)
//	service, ok := yamlObj["service"].(string)
//	if ok {
//		r.Service = service
//	}
//
//	dir, ok := yamlObj["dir"].(string)
//	if ok {
//		r.Dir = dir
//	}
//
//	sourceFiles, ok := yamlObj["source_files"].([]string)
//	if ok {
//		r.SourceFiles = NewSliceVariable(config_keys.SOURCE_FILES_KEY, sourceFiles)
//	}
//
//	dependencyBytes, ok := yamlObj["depends_on"].([]byte)
//	var steps []*Step
//	if ok {
//		if err := json.Unmarshal(dependencyBytes, &steps); err != nil {
//			return err
//		}
//		r.Depends = steps
//	}
//
//	contextEnvBytes, ok := yamlObj["context_envs"].([]byte)
//	var envs []*ContextEnv
//	if ok {
//		if err := json.Unmarshal(contextEnvBytes, &envs); err != nil {
//			return err
//		}
//		r.ContextEnv = envs
//	}
//
//	contextSteps, ok := yamlObj["context_steps"].([]byte)
//	var csteps []*ContextStep
//	if ok {
//		if err := json.Unmarshal(contextSteps, &csteps); err != nil {
//			return err
//		}
//		r.ContextEnv = envs
//	}
//
//	var contextInfo ContextInfo
//	b, ok := yamlObj["context"].([]byte)
//	if ok {
//		if err := json.Unmarshal(b, &contextInfo); err != nil {
//			return err
//		}
//	}
//
//	return nil
//}

func ReadConfig() *OzoneConfig {
	ozoneConfig := OzoneConfig{}

	dat, err := ioutil.ReadFile("./Ozonefile")
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	err = yaml.Unmarshal(dat, &ozoneConfig)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	for _, b := range ozoneConfig.PreUtilities {
		b.Type = PreUtilityType
	}
	for _, b := range ozoneConfig.Builds {
		b.Type = BuildType
	}
	for _, d := range ozoneConfig.Deploys {
		d.Type = DeployType
	}
	for _, b := range ozoneConfig.Tests {
		b.Type = TestType
	}
	for _, b := range ozoneConfig.PostUtilities {
		b.Type = PostUtilityType
	}

	//ozoneConfig.BuildVars["PROJECT"] = ozoneConfig.ProjectName
	//osEnv := OSEnvToVarsMap()
	//ozoneConfig.BuildVars = RenderNoMerge(ozoneConfig.BuildVars, osEnv)
	//ozoneConfig.BuildVars = RenderNoMerge(ozoneConfig.BuildVars, ozoneConfig.BuildVars)

	err = ozoneConfig.BuildVars.RenderFilters()

	if err != nil {
		log.Fatalln(err)
	}

	return &ozoneConfig
}
