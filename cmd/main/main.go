package main

import (
	"cli"
)

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

func createLocalAuxFolder() {

}

func main() {

	//dirPath := "../../"

	//ozone.Test()
	//config.ReadConfig()

	//cmd := exec.Command("./main")
	//cmd.Stdout = os.Stdout
	//cmd.Stderr = os.Stderr
	//err := cmd.Run()

	cli.Execute()

	//switch run {
	//case 0:

	//	return
	//case 1:
	//	debug.Build()
	//}

}