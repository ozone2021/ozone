package env

import (
	"fmt"
	"github.com/mitchellh/go-homedir"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
)

type k8sSecret struct {
	stringData map[string]string `yaml:"stringData"`
}



func FromSecret(varsMap map[string]string) (map[string]string, error) {

	secretFile, ok := varsMap["SECRET_FILE"]
	if ok && secretFile != "" {
		expandedSecretFile, err := homedir.Expand(secretFile)
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

			varsMap[strKey] = strValue
		}
	} else {
		log.Fatalln("K8s/from_secret needs env var SECRET_FILE")
	}
	return varsMap, nil
}