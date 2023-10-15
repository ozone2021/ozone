package main

import (
	"github.com/ozone2021/ozone/ozone/cmd/cli"
	"log"
	"os"
)

func main() {
	tmpOzone := "/tmp/ozone"
	_, err := os.Stat(tmpOzone)
	if os.IsNotExist(err) {
		err := os.Mkdir(tmpOzone, 0755)
		if err != nil {
			log.Fatalln(err)
		}
	}

	cli.Execute()
}
