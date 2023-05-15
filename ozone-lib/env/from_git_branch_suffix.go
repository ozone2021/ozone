package env

import (
	"errors"
	"fmt"
	. "github.com/ozone2021/ozone/ozone-lib/config/config_variable"
	"gopkg.in/src-d/go-git.v4"
	"strings"
)

// eg for release/1.0, gets the output 1.0
func DynamicFromGitReleaseBranch(ordinal int, varsMap, fromIncludeMap *VariableMap) error {
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
	if err != nil || branchName == "{{GIT_BRANCH}}" { // TODO prune unset vars before env or build
		branchName = string(reference.Name())
	}

	branchName = strings.TrimPrefix(branchName, "refs/heads/")

	branchNameParts := strings.Split(branchName, "/")
	if len(branchNameParts) != 2 {
		return errors.New(fmt.Sprintf("Branch name %s doesn't contain single forward slash.", branchName))
	}

	releaseVersion := strings.ToLower(branchNameParts[1])

	releaseVersionRFC1123 := strings.ReplaceAll(releaseVersion, ".", "-")

	namespace := fmt.Sprintf("release-%s", releaseVersionRFC1123)
	fromIncludeMap.AddVariable(NewStringVariable("NAMESPACE", namespace), ordinal)
	fromIncludeMap.AddVariable(NewStringVariable("SUBDOMAIN", fmt.Sprintf("%s.", namespace)), ordinal)

	return nil
}
