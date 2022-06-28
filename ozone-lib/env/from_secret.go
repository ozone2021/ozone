package env

import (
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/mitchellh/go-homedir"
	. "github.com/ozone2021/ozone/ozone-lib/config/config_variable"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
)

type k8sSecret struct {
	stringData map[string]string `yaml:"stringData"`
}

func FromSecretFile(ordinal int, varsMap, fromIncludeMap *VariableMap) error {

	secretFile, ok := varsMap.GetVariable("SECRET_FILE")
	if !ok {
		return errors.New("Need SECRET_FILE")
	}

	if secretFile.String() != "" {
		expandedSecretFile, err := homedir.Expand(secretFile.String())
		if err != nil {
			return fmt.Errorf("Secret file error:  #%v \n", err)
		}

		yamlFile, err := ioutil.ReadFile(expandedSecretFile)
		if err != nil {
			return fmt.Errorf("Secret file error:  #%v ", err)
		}
		m := make(map[string]interface{})
		err = yaml.Unmarshal(yamlFile, &m)
		if err != nil {
			return fmt.Errorf("K8s/from_secret error: Unmarshal SECRET_FILE: %v", err)
		}
		stringDataMap, ok := m["stringData"].(map[interface{}]interface{})
		if !ok {
			return fmt.Errorf("Argument is not a map")
		}
		for key, value := range stringDataMap {
			strKey := fmt.Sprintf("%v", key)
			strValue := fmt.Sprintf("%v", value)

			fromIncludeMap.AddVariable(NewStringVariable(strKey, strValue), ordinal)
		}
	} else {
		log.Fatalln("K8s/from_secret needs env var SECRET_FILE")
	}
	return nil
}

func FromSecret64(ordinal int, varsMap, fromIncludeMap *VariableMap) error {

	secret64, ok := varsMap.GetVariable("SECRET_BASE64")
	if !ok {
		return errors.New("Need SECRET_BASE64")
	}

	if secret64.String() != "" {
		var decode64Bytes []byte
		decode64Bytes, err := base64.StdEncoding.DecodeString(secret64.String())
		if err != nil {
			return fmt.Errorf("SECRET_BASE64 decode error:  #%v ", err)
		}
		m := make(map[string]interface{})
		err = yaml.Unmarshal(decode64Bytes, &m)
		if err != nil {
			return fmt.Errorf("K8s/from_secret error: Unmarshal SECRET_FILE: %v", err)
		}
		stringDataMap, ok := m["stringData"].(map[interface{}]interface{})
		if !ok {
			return fmt.Errorf("Argument is not a map")
		}
		for key, value := range stringDataMap {
			strKey := fmt.Sprintf("%v", key)
			strValue := fmt.Sprintf("%v", value)

			fromIncludeMap.AddVariable(NewStringVariable(strKey, strValue), ordinal)
		}
	}
	return nil
}
