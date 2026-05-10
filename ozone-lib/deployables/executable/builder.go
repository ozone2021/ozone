package executable

import (
	"errors"
	"fmt"
	"net/rpc"
	"os"

	process_manager "github.com/ozone2021/ozone/ozone-daemon-lib/process-manager-queries"
	"github.com/ozone2021/ozone/ozone-lib/config/config_variable"
	"github.com/ozone2021/ozone/ozone-lib/logger_lib"
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
		errorMessage := fmt.Sprintf("Error dialing process manager: %s", err.Error())
		logger.Println(errorMessage)
		return errors.New(errorMessage)
	}
	err = client.Call("ProcessManager.AddProcess", query, nil)
	if err != nil {
		errorMessage := fmt.Sprintf("Error dialing process manager: %s", err.Error())
		logger.Println(errorMessage)
		return errors.New(errorMessage)
	}
	return nil
}
