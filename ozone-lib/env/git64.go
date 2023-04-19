package env

import (
	"encoding/base64"
	"errors"
	"fmt"
	. "github.com/ozone2021/ozone/ozone-lib/config/config_variable"
	"github.com/ozone2021/ozone/ozone-lib/utils"
	"gopkg.in/src-d/go-git.v4"
	"log"
	"regexp"
	"strings"
)

func StaticFromGitDirBranchNameHash(ordinal int, varsMap, fromIncludeMap *VariableMap) error {
	dirPathVar, ok := varsMap.GetVariable("GIT_DIR")
	dirPath := dirPathVar.String()

	if !ok {
		dirPath = "./"
	}

	repo, err := git.PlainOpen(dirPath)
	if err != nil {
		return err
	}

	reference, err := repo.Head()
	if err != nil {
		return err
	}

	branchNameVar, ok := varsMap.GetVariable("GIT_BRANCH")
	branchName := branchNameVar.String()
	if err != nil || branchName == "{{GIT_BRANCH}}" {
		branchName = string(reference.Name())
	}
	branchName = strings.TrimPrefix(branchName, "refs/heads/")
	reg, err := regexp.Compile("[^a-zA-Z0-9]+")
	if err != nil {
		log.Fatal(err)
	}
	namespace := reg.ReplaceAllString(branchName, "")

	fromIncludeMap.AddVariable(NewStringVariable("NAMESPACE", namespace), ordinal)
	fromIncludeMap.AddVariable(NewStringVariable("SUBDOMAIN", fmt.Sprintf("%s.", namespace)), ordinal)

	return nil

}

func DynamicFromGitDirBranchNameHash(ordinal int, varsMap, fromIncludeMap *VariableMap) error {
	dirPath := "./"
	dirPathVar, ok := varsMap.GetVariable("GIT_DIR")
	if ok {
		dirPath = dirPathVar.String()
	}

	repo, err := git.PlainOpen(dirPath)
	if err != nil {
		return err
	}

	reference, err := repo.Head()
	if err != nil {
		return err
	}

	branchNameVar, ok := varsMap.GetVariable("GIT_BRANCH")
	if !ok {
		return errors.New("GIT_BRANCH not set")
	}
	branchName := branchNameVar.String()
	if err != nil || branchName == "{{GIT_BRANCH}}" {
		branchName = string(reference.Name())
	}

	git64Hash := base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString([]byte(branchName))

	namespace := git64Hash
	if len([]byte(git64Hash)) > 12 {
		namespace = strings.ToLower(git64Hash)[:12]
	}
	fromIncludeMap.AddVariable(NewStringVariable("NAMESPACE", namespace), ordinal)
	fromIncludeMap.AddVariable(NewStringVariable("SUBDOMAIN", fmt.Sprintf("%s.", namespace)), ordinal)

	return nil
}

func GitSubmoduleHash(ordinal int, varsMap, fromIncludeMap *VariableMap) error {
	for _, arg := range []string{
		"DOCKER_REGISTRY",
		"DIR",
		"SERVICE",
	} {
		if err := utils.ParamsOK("GitSubmoduleHash", arg, varsMap); err != nil {
			return err
		}
	}
	dockerRegistry, _ := varsMap.GetVariable("DOCKER_REGISTRY")
	dirPathVar, ok := varsMap.GetVariable("GIT_DIR")
	dirPath := dirPathVar.String()

	if !ok {
		dirPath = "./"
	}
	serviceName, _ := varsMap.GetVariable("SERVICE")

	repo, err := git.PlainOpen(dirPath)
	if err != nil {
		return err
	}

	reference, err := repo.Head()
	if err != nil {
		return err
	}
	submoduleHash := reference.Hash().String()

	fromIncludeMap.AddVariable(NewStringVariable("DOCKER_FULL_TAG", fmt.Sprintf("%s/%s:%s", dockerRegistry, serviceName, submoduleHash)), ordinal)
	return nil
}
