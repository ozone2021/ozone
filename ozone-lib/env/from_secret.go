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

func FromSecretFile(ordinal int, varsParamMap VariableMap) (VariableMap, error) {
	varsMap := make(VariableMap)

	secretFile, ok := varsParamMap["SECRET_FILE"]
	if !ok {
		return nil, errors.New("Need SECRET_FILE")
	}

	if secretFile.ToString() != "" {
		expandedSecretFile, err := homedir.Expand(secretFile.ToString())
		if err != nil {
			return nil, fmt.Errorf("Secret file error:  #%v \n", err)
		}

		yamlFile, err := ioutil.ReadFile(expandedSecretFile)
		if err != nil {
			return nil, fmt.Errorf("Secret file error:  #%v ", err)
		}
		m := make(map[string]interface{})
		err = yaml.Unmarshal(yamlFile, &m)
		if err != nil {
			return nil, fmt.Errorf("K8s/from_secret error: Unmarshal SECRET_FILE: %v", err)
		}
		stringDataMap, ok := m["stringData"].(map[interface{}]interface{})
		if !ok {
			return nil, fmt.Errorf("Argument is not a map")
		}
		for key, value := range stringDataMap {
			strKey := fmt.Sprintf("%v", key)
			strValue := fmt.Sprintf("%v", value)

			varsMap[strKey] = NewStringVariable(strValue, ordinal)
		}
	} else {
		log.Fatalln("K8s/from_secret needs env var SECRET_FILE")
	}
	return varsMap, nil
}

func FromSecret64(ordinal int, varsParamMap VariableMap) (VariableMap, error) {
	varsMap := make(VariableMap)

	secret64, ok := varsParamMap["SECRET_BASE64"]
	if !ok {
		return nil, errors.New("Need SECRET_BASE64")
	}

	if secret64.ToString() != "" {
		var decode64Bytes []byte
		decode64Bytes, err := base64.StdEncoding.DecodeString(secret64.ToString())
		if err != nil {
			return nil, fmt.Errorf("SECRET_BASE64 decode error:  #%v ", err)
		}
		m := make(map[string]interface{})
		err = yaml.Unmarshal(decode64Bytes, &m)
		if err != nil {
			return nil, fmt.Errorf("K8s/from_secret error: Unmarshal SECRET_FILE: %v", err)
		}
		stringDataMap, ok := m["stringData"].(map[interface{}]interface{})
		if !ok {
			return nil, fmt.Errorf("Argument is not a map")
		}
		for key, value := range stringDataMap {
			strKey := fmt.Sprintf("%v", key)
			strValue := fmt.Sprintf("%v", value)

			varsMap[strKey] = NewStringVariable(strValue, ordinal)
		}
	}
	return varsMap, nil
}
