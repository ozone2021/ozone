package env

import (
	"encoding/base64"
	"fmt"
	. "github.com/ozone2021/ozone/ozone-lib/config/config_variable"
	"github.com/ozone2021/ozone/ozone-lib/utils"
	"gopkg.in/src-d/go-git.v4"
	"log"
	"regexp"
	"strings"
)

func StaticFromGitDirBranchNameHash(varsParamMap VariableMap) (VariableMap, error) {
	dirPath, err := GenVarToString(varsParamMap, "ROOT_DIR")

	if dirPath == "" {
		dirPath = "./"
	}

	varsMap := make(VariableMap)

	repo, err := git.PlainOpen(dirPath)
	if err != nil {
		return nil, err
	}

	reference, err := repo.Head()
	if err != nil {
		return nil, err
	}

	branchName, err := GenVarToString(varsParamMap, "GIT_BRANCH")
	if err != nil || branchName == "{{GIT_BRANCH}}" {
		branchName = string(reference.Name())
	}
	branchName = strings.TrimPrefix(branchName, "refs/heads/")
	log.Printf("Branchname %s \n", branchName)
	reg, err := regexp.Compile("[^a-zA-Z0-9]+")
	if err != nil {
		log.Fatal(err)
	}
	namespace := reg.ReplaceAllString(branchName, "")

	varsMap["NAMESPACE"] = namespace
	varsMap["SUBDOMAIN"] = fmt.Sprintf("%s.", namespace)

	return varsMap, nil

}

func DynamicFromGitDirBranchNameHash(varsParamMap VariableMap) (VariableMap, error) {

	dirPath, err := GenVarToString(varsParamMap, "ROOT_DIR")
	if err != nil {
		dirPath = "./"
	}

	varsMap := make(VariableMap)

	repo, err := git.PlainOpen(dirPath)
	if err != nil {
		return nil, err
	}

	reference, err := repo.Head()
	if err != nil {
		return nil, err
	}

	branchName, err := GenVarToString(varsParamMap, "GIT_BRANCH")
	if err != nil || branchName == "{{GIT_BRANCH}}" {
		branchName = string(reference.Name())
	}
	log.Printf("Branchname %s \n", branchName)

	git64Hash := base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString([]byte(branchName))

	namespace := git64Hash
	if len([]byte(git64Hash)) > 12 {
		namespace = strings.ToLower(git64Hash)[:12]
	}
	varsMap["NAMESPACE"] = namespace
	varsMap["SUBDOMAIN"] = fmt.Sprintf("%s.", namespace)

	return varsMap, nil
}



func GitSubmoduleHash(varsParamMap VariableMap) (VariableMap, error) {
	for _, arg := range []string{
		"DOCKER_REGISTRY",
		"DIR",
		"SERVICE",
	} {
		if err := utils.ParamsOK("GitSubmoduleHash", arg, varsParamMap); err != nil {
			return nil, err
		}
	}
	dockerRegistry := varsParamMap["DOCKER_REGISTRY"]
	dirPath, err := GenVarToString(varsParamMap, "DIR")
	if err != nil {
		dirPath = "./"
	}
	serviceName := varsParamMap["SERVICE"]

	varsMap := make(VariableMap)

	repo, err := git.PlainOpen(dirPath)
	if err != nil {
		return nil, err
	}

	reference, err := repo.Head()
	if err != nil {
		return nil, err
	}
	submoduleHash := reference.Hash().String()

	varsMap["DOCKER_FULL_TAG"] = fmt.Sprintf("%s/%s:%s", dockerRegistry, serviceName, submoduleHash)
	return varsMap, nil
}
