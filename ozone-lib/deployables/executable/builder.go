package executable

import (
    "fmt"
	"log"
	"net/rpc"
	"os"
	process_manager "github.com/ozone2021/ozone/ozone-daemon-lib/process-manager"
)

func Build(serviceName string, env map[string]string) error {
	ozoneWorkingDir, err := os.Getwd()
	if err != nil {
		return err
	}
	cmdString := fmt.Sprintf("__OUTPUT__/executable")

	query := &process_manager.ProcessCreateQuery{
		serviceName,
		"__OUTPUT__/",
		ozoneWorkingDir,
		cmdString,
		false,
		false,
		env,
	}

	client, err := rpc.DialHTTP("tcp", ":8000")
	if err != nil {
		log.Fatal("dialing:", err)
		return err
	}
	err = client.Call("ProcessManager.AddProcess", query, nil)
	if err != nil {
		log.Fatal("arith error:", err)
		return err
	}
	return nil
}
