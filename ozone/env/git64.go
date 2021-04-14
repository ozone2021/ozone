package env

import (
	"encoding/base64"
	"fmt"
	"gopkg.in/src-d/go-git.v4"
	"ozone-lib/config"
)

func getParams() []string {
	return []string{
		"DOCKER_REGISTRY",
		"DIR",
		"SERVICE",
	}
}

func FromGitSubmoduleBranchHash(varsParamMap map[string]string) (map[string]string, error) {
	for _, arg := range getParams() {
		if err := config.ParamsOK(arg, varsParamMap); err != nil {
			return nil, err
		}
	}
	dockerRegistry := varsParamMap["DOCKER_REGISTRY"]
	dirPath := varsParamMap["DIR"]
	serviceName := varsParamMap["SERVICE"]

	varsMap := make(map[string]string)

	repo, err := git.PlainOpen(dirPath)
	if err != nil {
		return nil, err
	}

	reference, err := repo.Head()
	if err != nil {
		return nil, err
	}
	branchName := string(reference.Name())

	git64Hash := base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString([]byte(branchName))

	varsMap["FULL_TAG"] = fmt.Sprintf("%s/%s:%s", dockerRegistry, serviceName, git64Hash)

	return varsMap, nil
}