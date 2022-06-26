package config

import (
	"errors"
	"fmt"
	"github.com/ozone2021/ozone/ozone-lib/config/config_utils"
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
	Name     string      `yaml:"name"`
	WithVars VariableMap `yaml:"with_vars"`
	Type     string      `yaml:"type"`
}

type Environment struct {
	Name     string      `yaml:"name"`
	WithVars VariableMap `yaml:"with_vars"`
	Includes []*Include  `yaml:"include"`
}

type Step struct {
	Type     string      `yaml:"type"`
	Name     string      `yaml:"name"`
	WithVars VariableMap `yaml:"with_vars"`
}

type ContextStep struct {
	Context string   `yaml:"context"`
	Steps   []*Step  `yaml:"steps"`
	WithEnv []string `yaml:"with_env"`
}

type ContextEnv struct {
	Context string   `yaml:"context"`
	WithEnv []string `yaml:"with_env"`
}

type Runnable struct {
	Name        string   `yaml:"name"`
	Service     string   `yaml:"service"`
	Dir         string   `yaml:"dir"`
	SourceFiles []string `yaml:"source_files"`
	WhenChanged []string `yaml:"when_changed"`
	Depends     []*Step  `yaml:"depends_on"`
	//WithEnv     	[]string      	`yaml:"with_env"`
	ContextEnv   []*ContextEnv  `yaml:"context_envs"`
	ContextSteps []*ContextStep `yaml:"context_steps"`
	Type         RunnableType
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

func (config *OzoneConfig) fetchEnv(ordinal int, envName string, scopeMap VariableMap) (VariableMap, error) {
	nameFound := false
	varsMap := make(VariableMap)
	for _, e := range config.Environments {
		if e.Name == envName {
			nameFound = true
			if len(e.Includes) != 0 {
				for _, incl := range e.Includes {
					var inclVarsMap VariableMap
					var err error
					if incl.Type == "builtin" {
						inclParamVarsMap := config_utils.MergeMapsSelfRender(ordinal, incl.WithVars, scopeMap)
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

					inclVarsRendered := config_utils.RenderNoMerge(ordinal, inclVarsMap, scopeMap)
					varsMap = config_utils.MergeMapsSelfRender(ordinal, varsMap, inclVarsRendered)
				}
			}
			renderedEnvVars := config_utils.RenderNoMerge(ordinal, e.WithVars, scopeMap)

			varsMap = config_utils.MergeMapsSelfRender(ordinal, varsMap, renderedEnvVars)
		}
	}
	if nameFound == false {
		return nil, errors.New(fmt.Sprintf("Environment %s not found \n", envName))
	}

	return varsMap, nil
}

func (config *OzoneConfig) fetchBuiltinEnvFromInclude(ordinal int, envName string, varsMap VariableMap) (VariableMap, error) {
	var err error
	fromIncludeMap := make(VariableMap)

	// TODO
	switch envName {
	case "env/from_k8s_secret_file":
		fromIncludeMap, err = env.FromSecretFile(ordinal, varsMap)
	case "env/from_k8s_secret64":
		fromIncludeMap, err = env.FromSecret64(varsMap)
	case "env/from_env_file":
		fromIncludeMap, err = env.FromEnvFile(ordinal, varsMap)
	case "env/git_log_hash":
		fromIncludeMap, err = git_env.GitLogHash(varsMap)
	case "env/git_directory_branch_hash":
		fromIncludeMap, err = env.DynamicFromGitDirBranchNameHash(varsMap)
	case "env/git_directory_branch_static":
		fromIncludeMap, err = env.StaticFromGitDirBranchNameHash(varsMap)
	case "env/git_submodule_commit_hash":
		fromIncludeMap, err = env.GitSubmoduleHash(varsMap)
	}

	if err != nil {
		return nil, err
	}

	return fromIncludeMap, nil
}

func (config *OzoneConfig) FetchEnvs(ordinal int, envList []string, scope VariableMap) (VariableMap, error) {
	ordinal++
	varsMap := make(VariableMap)

	for _, env := range envList {
		renderedEnv, err := PongoRender(env, scope)
		if err != nil {
			return nil, err
		}

		fetchedMap, err := config.fetchEnv(ordinal, renderedEnv, scope)
		if err != nil {
			return nil, err
		}
		varsMap = config_utils.MergeMapsSelfRender(ordinal, varsMap, fetchedMap)
	}
	varsMap = config_utils.RenderNoMerge(ordinal, varsMap, scope)
	return varsMap, nil
}

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

	err = config_utils.RenderFilters(ozoneConfig.BuildVars)

	if err != nil {
		log.Fatalln(err)
	}

	return &ozoneConfig
}
