package config

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
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
	WithVars		[]Var			`yaml:"with_vars"`
	Type			string			`yaml:"type"`
}

type Environment struct {
	Name 			string			`yaml:"name"`
	WithVars		[]Var			`yaml:"with_vars"`
	Includes		[]Include		`yaml:"include"`
}

type Step struct {
	Type			string			`yaml:"type"`
	Value			string			`yaml:"value"`
	WithVars		[]Var			`yaml:"with_vars"`
}

type ContextStep struct {
	Context string `yaml:"context"`
	Steps   []Step `yaml:"steps"`
	WithEnv	[]string	`yaml:"with_env"`
}

type Runnable struct {
	Name        	string			`yaml:"name"`
	Service			string			`yaml:"service"`
	Dir         	string			`yaml:"dir"`
	WithEnv     	[]string      	`yaml:"with_env"`
	ContextSteps	[]ContextStep 	`yaml:"context_steps"`
}

type OzoneConfig struct {
	ContextInfo		ContextInfo		`yaml:"context"`
	BuildVars 		[]Var 			`yaml:"build_vars"`
	Environments	[]Environment	`yaml:"environments"`
	Builds			[]Runnable		`yaml:"builds"`
	Deploys			[]Runnable		`yaml:"deploys"`
}

func(config *OzoneConfig) HasRunnable(name string) bool {
	return config.ListHasRunnable(name, config.Builds) || config.HasDeploy(name)
}

func(config *OzoneConfig) HasDeploy(name string) bool {
	return config.ListHasRunnable(name, config.Deploys)
}
func(config *OzoneConfig) DeploysHasService(service string) bool {
	for _, runnable := range config.Deploys {
		if runnable.Service == service {
			return true
		}
	}
	return false
}

func(config *OzoneConfig) ListHasRunnable(name string, runnables []Runnable) bool {
	for _, runnable := range runnables {
		if runnable.Name == name {
			return true
		}
	}
	return false
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

	return &ozoneConfig
}

