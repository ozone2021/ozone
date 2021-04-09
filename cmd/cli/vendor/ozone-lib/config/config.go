package config

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
)

type Var struct {
	Name 			string			`yaml:"name"`
	VarType			string			`yaml:"type"`
	Value 			string			`yaml:"value"`
}

type Include struct {
	Name 			string			`yaml:"name"`
	WithVars		[]Var			`yaml:"with_vars"`
}

type Environment struct {
	Name 			string			`yaml:"name"`
	RuntimeVars		[]string		`yaml:"runtime_vars"`
	Includes		[]Include		`yaml:"include"`
}

type Step struct {
	Type			string			`yaml:"type"`
	Value			string			`yaml:"value"`
	WithVars		[]Var			`yaml:"with_vars"`
}

type EnvStep struct {
	Env				string			`yaml:"env"`
	Steps			[]Step			`yaml:"steps"`
}

type Runnable struct {
	Name			string			`yaml:"name"`
	Dir				string			`yaml:"dir"`
	WithEnv			[]string		`yaml:"with_env"`
	EnvSteps		[]EnvStep		`yaml:"env_steps"`
}

type OzoneConfig struct {
	BuildVars 		[]Var 			`yaml:"build_vars"`
	Environments	[]Environment	`yaml:"environments"`
	Builds			[]Runnable		`yaml:"builds"`
	Deploys			[]Runnable		`yaml:"deploys"`
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