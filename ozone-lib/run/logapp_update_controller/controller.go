package logapp_update_controller

import (
	"context"
	"github.com/jinzhu/copier"
	. "github.com/ozone2021/ozone/ozone-lib/brpc_log_registration/log_registration_server"
	log_server "github.com/ozone2021/ozone/ozone-lib/brpc_log_server/log_server_pb"
	"github.com/ozone2021/ozone/ozone-lib/runspec"
	"google.golang.org/grpc"
	"log"
)

type LogappUpdateController struct {
	registrationServer *LogRegistrationServer
	registeredLogApps  map[string]*LogAppDetails
	logAppGrpcClients  map[string]log_server.LogUpdateServiceClient
	incomingLogApps    <-chan *LogAppDetails // TODO check arrow
	resultUpdate       chan *runspec.RunResult
}

func NewLogappUpdateController(ozoneWorkingDir string) *LogappUpdateController {
	incomingLogApps := make(chan *LogAppDetails)

	return &LogappUpdateController{
		registrationServer: NewLogRegistrationServer(ozoneWorkingDir, incomingLogApps),
		registeredLogApps:  make(map[string]*LogAppDetails),
		logAppGrpcClients:  make(map[string]log_server.LogUpdateServiceClient),
		incomingLogApps:    incomingLogApps,
	}
}

func (c *LogappUpdateController) Start() {
	go c.registrationServer.Start()

	for {
		select {
		case logAppDetails := <-c.incomingLogApps:
			c.registeredLogApps[logAppDetails.Id] = logAppDetails
		case runResult := <-c.resultUpdate:
			c.updateLogApps(runResult)
		}
	}
}

func (c *LogappUpdateController) connectToLogApp(details *LogAppDetails) {
	conn, err := grpc.Dial(details.PipePath, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}
	c.logAppGrpcClients[details.Id] = log_server.NewLogUpdateServiceClient(conn)

}

func (c *LogappUpdateController) updateLogApps(runResult *runspec.RunResult) {
	for _, logAppClient := range c.logAppGrpcClients {
		var runResultPb *log_server.RunResult
		err := copier.Copy(&runResultPb, &runResult)
		if err != nil {
			log.Fatalf("failed to copy: %v", err)
		}
		logAppClient.UpdateRunResult(context.Background(), runResultPb)
	}
}
