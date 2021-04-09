package git

import (
	"encoding/base64"
	"gopkg.in/src-d/go-git.v4"
)

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

func Git64(dirPath string) string {
	repo, err := git.PlainOpen(dirPath)
	checkErr(err)

	reference, err := repo.Head()
	checkErr(err)
	branchName := string(reference.Name())

	return base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString([]byte(branchName))
}
