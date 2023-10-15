package executable

import (
	"fmt"
	process_manager "github.com/ozone2021/ozone/ozone-daemon-lib/process-manager-queries"
	"github.com/ozone2021/ozone/ozone-lib/config/config_variable"
	"github.com/ozone2021/ozone/ozone-lib/logger_lib"
	"net/rpc"
	"os"
)

func Build(env *config_variable.VariableMap, logger *logger_lib.Logger) error {
	ozoneWorkingDir, err := os.Getwd()
	if err != nil {
		return err
	}
	cmdString := fmt.Sprintf("__OUTPUT__/executable")

	query := &process_manager.ProcessCreateQuery{
		"serviceName-TODO not all runnables have a service",
		"__OUTPUT__/",
		ozoneWorkingDir,
		cmdString,
		false,
		false,
		env,
	}

	client, err := rpc.DialHTTP("tcp", ":8000")
	if err != nil {
		logger.Fatal("dialing:", err)
		return err
	}
	err = client.Call("ProcessManager.AddProcess", query, nil)
	if err != nil {
		logger.Fatal("arith error:", err)
		return err
	}
	return nil
}
