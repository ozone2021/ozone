package logapp_update_controller

import (
	"context"
	"fmt"
	"github.com/jinzhu/copier"
	"github.com/ozone2021/ozone/ozone-lib/config/runspec"
	"github.com/ozone2021/ozone/ozone-lib/logs/brpc_log_server/log_server_pb"
	. "github.com/ozone2021/ozone/ozone-lib/run/brpc_log_registration/log_registration_server"
	"google.golang.org/grpc"
	"log"
)

type LogappUpdateController struct {
	registrationServer     *LogRegistrationServer
	registeredLogApps      map[string]*ConnectedLogApp
	incomingLogApps        <-chan *LogAppDetails // TODO check arrow
	channelHandlerShutdown chan struct{}
}

type ConnectedLogApp struct {
	*LogAppDetails
	log_server_pb.LogUpdateServiceClient
}

func NewLogappUpdateController(ozoneWorkingDir string, incomingLogAppDetails chan *LogAppDetails) *LogappUpdateController {
	return &LogappUpdateController{
		registrationServer: NewLogRegistrationServer(ozoneWorkingDir, incomingLogAppDetails),
		registeredLogApps:  make(map[string]*ConnectedLogApp),
		incomingLogApps:    incomingLogAppDetails,
	}
}

func (c *LogappUpdateController) Start() {
	for {
		select {
		case logAppDetails := <-c.incomingLogApps:
			c.registeredLogApps[logAppDetails.Id] = &ConnectedLogApp{
				LogAppDetails:          logAppDetails,
				LogUpdateServiceClient: c.connectToLogApp(logAppDetails),
			}
		case <-c.channelHandlerShutdown:
			return
		}
	}
}

func (c *LogappUpdateController) connectToLogApp(details *LogAppDetails) log_server_pb.LogUpdateServiceClient {
	conn, err := grpc.Dial(fmt.Sprintf("unix://"+details.PipePath), grpc.WithInsecure())
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}
	return log_server_pb.NewLogUpdateServiceClient(conn)
}

func (c *LogappUpdateController) UpdateLogApps(runResult *runspec.RunResult) {
	for _, connectedApp := range c.registeredLogApps {
		runResultPb := &log_server_pb.RunResult{}
		err := copier.Copy(&runResultPb, &runResult)
		if err != nil {
			log.Fatalf("failed to copy: %v", err)
		}
		_, err = connectedApp.UpdateRunResult(context.Background(), runResultPb)
		if err != nil {
			log.Println("failed to update log app: ", err)
		}
	}
}
