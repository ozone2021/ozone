package _go

import (
	"fmt"
	"log"
	"net/rpc"
	"os"
	"ozone-daemon-lib/process-manager"
)



func Build(serviceName string, relativeDir string, file string, varsMap map[string]string) {
	ozoneWorkingDir, err := os.Getwd()
	if err != nil {
		log.Println(err)
	}
	cmdString := fmt.Sprintf("go build -o __OUTPUT__/executable %s/%s",
		relativeDir,
		file)

	query := &process_manager.ProcessCreateQuery{
		serviceName,
		ozoneWorkingDir,
		ozoneWorkingDir,
		cmdString,
		true,
		varsMap,
	}

	client, err := rpc.DialHTTP("tcp", ":8000")
	if err != nil {
		log.Fatal("dialing:", err)
	}
	err = client.Call("ProcessManager.AddProcess", query, nil)
	if err != nil {
		log.Fatal("arith error:", err)
	}
}