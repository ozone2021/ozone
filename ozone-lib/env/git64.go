package env

import (
	"encoding/base64"
	"fmt"
	"gopkg.in/src-d/go-git.v4"
	"github.com/JamesArthurHolland/ozone/ozone-lib/utils"
)

func FromGitDirBranchNameHash(varsParamMap map[string]string) (map[string]string, error) {
	for _, arg := range []string{
		"DIR",
	} {
		if err := utils.ParamsOK("FromGitDirBranchNameHash", arg, varsParamMap); err != nil {
			return nil, err
		}
	}
	dirPath := varsParamMap["DIR"]

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

	varsMap["NAMESPACE"] = git64Hash

	return varsMap, nil
}

func GitSubmoduleHash(varsParamMap map[string]string) (map[string]string, error) {
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
	dirPath := varsParamMap["DIR"]
	serviceName := varsParamMap["SERVICE"]

	varsMap := make(map[string]string)

	repo, err := git.PlainOpen(dirPath)
	reference, err := repo.Head()
	if err != nil {
		return nil, err
	}
	submoduleHash := reference.Hash().String()

	varsMap["FULL_TAG"] = fmt.Sprintf("%s/%s:%s", dockerRegistry, serviceName, submoduleHash)
	return varsMap, nil
}