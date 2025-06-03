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
	"sync"
	"time"
)

type LogappUpdateController struct {
	registrationServer     *LogRegistrationServer
	registeredLogApps      map[string]*ConnectedLogApp
	registeredLogAppsMutex sync.Mutex
	incomingLogApps        <-chan *LogAppDetails // TODO check arrow
	channelHandlerShutdown chan struct{}
	resetLogApps           chan struct{}
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
		resetLogApps:       make(chan struct{}),
	}
}

func (c *LogappUpdateController) ResetLogApps() {
	c.resetLogApps <- struct{}{}
}

func (c *LogappUpdateController) Start() {
	go c.SendHeartbeats()

	for {
		select {
		case logAppDetails := <-c.incomingLogApps:
			c.registeredLogAppsMutex.Lock()
			c.registeredLogApps[logAppDetails.Id] = &ConnectedLogApp{
				LogAppDetails:          logAppDetails,
				LogUpdateServiceClient: c.connectToLogApp(logAppDetails),
			}
			c.registeredLogAppsMutex.Unlock()
			c.updateAllFunction(false)
		case <-c.channelHandlerShutdown:
			return
		}
	}
}

func (c *LogappUpdateController) SendHeartbeats() {
	for {
		select {
		case <-c.resetLogApps:
			for id, _ := range c.registeredLogApps {
				c.updateAllFunction(true)
				c.deleteRegisteredApp(id)
			}
		default:
			for id, connectedApp := range c.registeredLogApps {
				_, err := connectedApp.ReceiveMainAppHeartbeat(context.Background(), &emptypb.Empty{})
				if err != nil {
					c.deleteRegisteredApp(id)
				}
			}
			time.Sleep(1 * time.Second)
		}
	}
}

func (c *LogappUpdateController) deleteRegisteredApp(id string) {
	c.registeredLogAppsMutex.Lock()
	delete(c.registeredLogApps, id)
	c.registeredLogAppsMutex.Unlock()
}

func (c *LogappUpdateController) connectToLogApp(details *LogAppDetails) log_server_pb.LogUpdateServiceClient {
	conn, err := grpc.Dial(fmt.Sprintf("unix://"+details.PipePath), grpc.WithInsecure())
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}
	return log_server_pb.NewLogUpdateServiceClient(conn)
}

func (c *LogappUpdateController) UpdateLogApps(runResult *runspec.RunResult, reset bool) {
	for id, connectedApp := range c.registeredLogApps {
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
		runResultPb.Reset_ = reset
		timeoutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, err = connectedApp.UpdateRunResult(timeoutCtx, runResultPb)
		if err != nil {
			c.deleteRegisteredApp(id)
			log.Println(err)
		}
	}
}
