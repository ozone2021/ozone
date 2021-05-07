package config

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
)

type RunnableType int

const (
	BuildType 	RunnableType = iota
	DeployType 	RunnableType = iota
	TestType	RunnableType = iota
)

type ContextInfo struct {
	Default			string			`yaml:"default"`
	List			[]string		`yaml:"list"`
}

type Var struct {
	Name 			string			`yaml:"name"`
	VarType			string			`yaml:"type"`
	Value 			string			`yaml:"value"`
}

type Include struct {
	Name 			string			`yaml:"name"`
	WithVars		[]*Var			`yaml:"with_vars"`
	Type			string			`yaml:"type"`
}

type Environment struct {
	Name 			string			`yaml:"name"`
	WithVars		[]*Var			`yaml:"with_vars"`
	Includes		[]*Include		`yaml:"include"`
}

type Step struct {
	Type			string					`yaml:"type"`
	Value			string					`yaml:"value"`
	WithVars		map[string]string		`yaml:"with_vars"`
}

type ContextStep struct {
	Context string		`yaml:"context"`
	Steps   []*Step		`yaml:"steps"`
	WithEnv	[]string	`yaml:"with_env"`
}

type Runnable struct {
	Name        	string			`yaml:"name"`
	Service			string			`yaml:"service"`
	Dir         	string			`yaml:"dir"`
	WithEnv     	[]string      	`yaml:"with_env"`
	ContextSteps	[]*ContextStep 	`yaml:"context_steps"`
	Type			RunnableType
}

type OzoneConfig struct {
	ContextInfo		ContextInfo		`yaml:"context"`
	BuildVars 		[]*Var 			`yaml:"build_vars"`
	Environments	[]*Environment	`yaml:"environments"`
	Builds			[]*Runnable		`yaml:"builds"`
	Deploys			[]*Runnable		`yaml:"deploys"`
	Tests			[]*Runnable		`yaml:"tests"`
}

func(config *OzoneConfig) FetchRunnable(name string) (bool, *Runnable) {
	if has, runnable := config.HasBuild(name); has == true {
		return true, runnable
	}
	if has, runnable := config.HasDeploy(name); has == true {
		return true, runnable
	}
	if has, runnable := config.HasTest(name); has == true {
		return true, runnable
	}

	return false, nil
}

func(config *OzoneConfig) HasBuild(name string) (bool, *Runnable) {
	return config.ListHasRunnableOfType(name, config.Builds, BuildType)
}

func(config *OzoneConfig) HasDeploy(name string) (bool, *Runnable) {
	return config.ListHasRunnableOfType(name, config.Deploys, DeployType)
}

func(config *OzoneConfig) HasTest(name string) (bool, *Runnable) {
	return config.ListHasRunnableOfType(name, config.Tests, TestType)
}

func(config *OzoneConfig) DeploysHasService(service string) bool {
	for _, runnable := range config.Deploys {
		if runnable.Service == service {
			return true
		}
	}
	return false
}

func(config *OzoneConfig) ListHasRunnableOfType(name string, runnables []*Runnable, givenType RunnableType) (bool, *Runnable) {
	for _, runnable := range runnables {
		if runnable.Name == name && runnable.Type == givenType {
			return true, runnable
		}
	}
	return false, nil
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

	for _, b := range ozoneConfig.Builds {
		b.Type = BuildType
	}
	for _, d := range ozoneConfig.Deploys {
		d.Type = DeployType
	}
	for _, b := range ozoneConfig.Tests {
		b.Type = TestType
	}


	return &ozoneConfig
}

