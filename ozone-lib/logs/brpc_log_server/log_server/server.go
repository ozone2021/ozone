package log_server

import (
	"context"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/elliotchance/orderedmap/v2"
	"github.com/jinzhu/copier"
	"github.com/ozone2021/ozone/ozone-lib/config/runspec"
	. "github.com/ozone2021/ozone/ozone-lib/logs/brpc_log_server/log_server_pb"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
	"log"
	"net"
	"os"
	"path/filepath"
	"time"
)

type UiChan chan tea.Msg

type LogServer struct {
	UnimplementedLogUpdateServiceServer
	pipePath            string
	grpcServer          *grpc.Server
	uiChan              UiChan
	mostRecentHeartbeat int64
	reconnectChan       chan int64
}

type Heartbeat struct {
	HeartbeatTime   int64
	ShouldReconnect bool
}

type RunResultUpdate struct {
	RunResult   *runspec.RunResult
	ShouldReset bool
	RunId       string
}

type LogAppDetails struct {
	Id string
}

func NewLogServer(pipePath string, uiChan UiChan, reconnectChan chan int64) *LogServer {
	return &LogServer{
		pipePath:      pipePath,
		grpcServer:    grpc.NewServer(),
		uiChan:        uiChan,
		reconnectChan: reconnectChan,
	}
}

func mkfifo(path string, mode os.FileMode) error {
	dirPath := filepath.Dir(path)
	return os.MkdirAll(dirPath, os.ModePerm)
}

func (s *LogServer) Start() {
	err := os.Remove(s.pipePath) // Remove the pipe if it already exists
	if err != nil && !os.IsNotExist(err) {
		log.Fatalf("failed to remove existing pipe: %v", err)
	}

	err = mkfifo(s.pipePath, 0666) // Create the named pipe
	if err != nil {
		log.Fatalf("failed to create named pipe: %v", err)
	}

	listener, err := net.Listen("unix", s.pipePath)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	defer listener.Close()

	// Register the server with the gRPC server
	RegisterLogUpdateServiceServer(s.grpcServer, s)

	// Start serving requests
	if err := s.grpcServer.Serve(listener); err != nil {
		log.Fatalf("failed to serve: %s", err)
	}
}

func (s *LogServer) UpdateRunResult(ctx context.Context, in *RunResult) (*emptypb.Empty, error) {
	runspecRunresult := &runspec.RunResult{}

	err := copier.CopyWithOption(&runspecRunresult, &in, copier.Option{IgnoreEmpty: true, DeepCopy: true})
	if err != nil {
		log.Printf("Error unmarshalling runResult %s", err)
		return nil, err
	}

	runspecRunresult.Index = orderedmap.NewOrderedMap[string, *runspec.CallstackResultNode]()
	for _, key := range in.IndexList {
		node := &runspec.CallstackResultNode{}
		err := copier.CopyWithOption(&node, &key, copier.Option{IgnoreEmpty: true, DeepCopy: true})
		if err != nil {
			log.Printf("Error unmarshalling runResult %s", err)
			return nil, err
		}
		runspecRunresult.Index.Set(node.Id, node)
	}

	s.uiChan <- &RunResultUpdate{
		RunResult:   runspecRunresult,
		ShouldReset: in.Reset_,
		RunId:       in.RunId,
	}

	return &emptypb.Empty{}, nil
}

// TODO heartbeat from logApp to server to make sure the logApp is still alive
func (s *LogServer) ReceiveMainAppHeartbeat(context.Context, *emptypb.Empty) (*emptypb.Empty, error) {
	s.mostRecentHeartbeat = time.Now().Unix()

	s.reconnectChan <- s.mostRecentHeartbeat

	return &emptypb.Empty{}, nil
}
