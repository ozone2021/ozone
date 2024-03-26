package logapp_update_controller

import (
	"context"
	"fmt"
	"github.com/jinzhu/copier"
	"github.com/ozone2021/ozone/ozone-lib/config/runspec"
	"github.com/ozone2021/ozone/ozone-lib/logs/brpc_log_server/log_server_pb"
	. "github.com/ozone2021/ozone/ozone-lib/run/brpc_log_registration/log_registration_server"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
	"log"
	"time"
)

type LogappUpdateController struct {
	registrationServer     *LogRegistrationServer
	registeredLogApps      map[string]*ConnectedLogApp
	incomingLogApps        <-chan *LogAppDetails // TODO check arrow
	channelHandlerShutdown chan struct{}
	updateAllFunction      runspec.UpdateAllListenersFunc
}

type ConnectedLogApp struct {
	*LogAppDetails
	log_server_pb.LogUpdateServiceClient
}

func NewLogappUpdateController(ozoneWorkingDir string, incomingLogAppDetails chan *LogAppDetails, updateFunction runspec.UpdateAllListenersFunc) *LogappUpdateController {
	return &LogappUpdateController{
		registrationServer: NewLogRegistrationServer(ozoneWorkingDir, incomingLogAppDetails),
		registeredLogApps:  make(map[string]*ConnectedLogApp),
		incomingLogApps:    incomingLogAppDetails,
		updateAllFunction:  updateFunction,
	}
}

func (c *LogappUpdateController) Start() {
	go c.SendHeartbeats()

	for {
		select {
		case logAppDetails := <-c.incomingLogApps:
			c.registeredLogApps[logAppDetails.Id] = &ConnectedLogApp{
				LogAppDetails:          logAppDetails,
				LogUpdateServiceClient: c.connectToLogApp(logAppDetails),
			}
			c.updateAllFunction()
		case <-c.channelHandlerShutdown:
			return
		}
	}
}

func (c *LogappUpdateController) SendHeartbeats() {
	for {
		for _, connectedApp := range c.registeredLogApps {
			_, err := connectedApp.ReceiveMainAppHeartbeat(context.Background(), &emptypb.Empty{})
			if err != nil {
				// TODO remove log app
			}
		}
		time.Sleep(1 * time.Second)
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
		err := copier.CopyWithOption(&runResultPb, &runResult, copier.Option{IgnoreEmpty: true, DeepCopy: true})
		if err != nil {
			log.Fatalf("failed to copy: %v", err)
		}
		for _, key := range runResult.Index.Keys() {
			node, _ := runResult.Index.Get(key)
			logNode := &log_server_pb.CallstackLogNode{}
			err := copier.CopyWithOption(&logNode, &node, copier.Option{IgnoreEmpty: true, DeepCopy: true})
			if err != nil {
				log.Fatalf("failed to copy: %v", err)
			}
			runResultPb.IndexList = append(runResultPb.IndexList, logNode)
		}
		_, err = connectedApp.UpdateRunResult(context.Background(), runResultPb)
		if err != nil {
			log.Println("failed to update log app: ", err)
		}
	}
}
