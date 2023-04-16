package debug

import (
	process_manager "github.com/ozone2021/ozone/ozone-daemon-lib/process-manager-queries"
	"log"
	"net/rpc"
	"os"
)

func Build() {
	ozoneWorkingDir, err := os.Getwd()
	if err != nil {
		log.Println(err)
	}

	query := &process_manager.DebugQuery{
		OzoneWorkingDir: ozoneWorkingDir,
	}

	client, err := rpc.DialHTTP("tcp", ":8000")
	if err != nil {
		log.Fatal("dialing:", err)
	}
	err = client.Call("ProcessManager.Debug", query, nil)
	if err != nil {
		log.Fatal("arith error:", err)
	}
}
