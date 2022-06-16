package env

import (
	"fmt"
	"gopkg.in/src-d/go-git.v4"
	"log"
	"regexp"
	"strings"
)

func DockerTagFromSha256(varsParamMap map[string]string) (map[string]string, error) {
	dirPath := varsParamMap["ROOT_DIR"]
	if dirPath == "" {
		dirPath = "./"
	}

	varsMap := make(map[string]string)

	repo, err := git.PlainOpen(dirPath)
	if err != nil {
		return nil, err
	}

	reference, err := repo.Head()
	if err != nil {
		return nil, err
	}

	branchName, ok := varsParamMap["GIT_BRANCH"]
	if !ok || branchName == "{{GIT_BRANCH}}" {
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

//func DynamicFromGitDirBranchNameHash(varsParamMap map[string]string) (map[string]string, error) {
//	dirPath := varsParamMap["ROOT_DIR"]
//	if dirPath == "" {
//		dirPath = "./"
//	}
//
//	varsMap := make(map[string]string)
//
//	repo, err := git.PlainOpen(dirPath)
//	if err != nil {
//		return nil, err
//	}
//
//	reference, err := repo.Head()
//	if err != nil {
//		return nil, err
//	}
//
//	branchName, ok := varsParamMap["GIT_BRANCH"]
//	if !ok || branchName == "{{GIT_BRANCH}}" {
//		branchName = string(reference.Name())
//	}
//	log.Printf("Branchname %s \n", branchName)
//
//	git64Hash := base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString([]byte(branchName))
//
//	namespace := git64Hash
//	if len([]byte(git64Hash)) > 12 {
//		namespace = strings.ToLower(git64Hash)[:12]
//	}
//	varsMap["NAMESPACE"] = namespace
//	varsMap["SUBDOMAIN"] = fmt.Sprintf("%s.", namespace)
//
//	return varsMap, nil
//}
//
//func GitSubmoduleHash(varsParamMap map[string]string) (map[string]string, error) {
//	for _, arg := range []string{
//		"DOCKER_REGISTRY",
//		"DIR",
//		"SERVICE",
//	} {
//		if err := utils.ParamsOK("GitSubmoduleHash", arg, varsParamMap); err != nil {
//			return nil, err
//		}
//	}
//	dockerRegistry := varsParamMap["DOCKER_REGISTRY"]
//	dirPath := varsParamMap["DIR"]
//	serviceName := varsParamMap["SERVICE"]
//
//	varsMap := make(map[string]string)
//
//	repo, err := git.PlainOpen(dirPath)
//	if err != nil {
//		return nil, err
//	}
//
//	reference, err := repo.Head()
//	if err != nil {
//		return nil, err
//	}
//	submoduleHash := reference.Hash().String()
//
//	varsMap["DOCKER_FULL_TAG"] = fmt.Sprintf("%s/%s:%s", dockerRegistry, serviceName, submoduleHash)
//	return varsMap, nil
//}