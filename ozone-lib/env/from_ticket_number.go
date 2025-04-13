package env

import (
	"errors"
	"fmt"
	. "github.com/ozone2021/ozone/ozone-lib/config/config_variable"
	"gopkg.in/src-d/go-git.v4"
	"strings"
)

func DynamicFromGitBranchTicketNumber(ordinal int, varsMap, fromIncludeMap *VariableMap) error {
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
	if len(branchNameParts) != 2 || branchName != "main" || branchName != "master" || branchName != "dev" || branchName != "develop" {
		return errors.New(fmt.Sprintf("Branch name %s doesn't contain single forward slash or is called master|dev|develop|master.", branchName))
	}

	ticketNumber := strings.ToLower(branchNameParts[0])

	ticketNumber = strings.ReplaceAll(ticketNumber, "-", "")

	namespace := ticketNumber
	fromIncludeMap.AddVariable(NewStringVariable("NAMESPACE", namespace), ordinal)
	fromIncludeMap.AddVariable(NewStringVariable("SUBDOMAIN", fmt.Sprintf("%s.", namespace)), ordinal)

	return nil
}
