package log_registration_server

import (
	"context"
	"fmt"
	. "github.com/ozone2021/ozone/ozone-lib/run/brpc_log_registration/log_registration_pb"
	"github.com/ozone2021/ozone/ozone-lib/utils"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
	"log"
	"net"
	"os"
	"path/filepath"
	"sync"
)

type LogRegistrationServer struct {
	UnimplementedRegistrationServiceServer
	ozoneWorkingDir    string
	ozoneSocketDirPath string
	output             chan *LogAppDetails
}

type LogAppDetails struct {
	Id       string
	PipePath string
}

func NewLogRegistrationServer(ozoneWorkingDir string, output chan *LogAppDetails) *LogRegistrationServer {
	return &LogRegistrationServer{
		ozoneWorkingDir:    ozoneWorkingDir,
		ozoneSocketDirPath: filepath.Join(utils.GetTmpDir(ozoneWorkingDir), "socks"),
		output:             output,
	}
}

func mkfifo(path string, mode os.FileMode) error {
	dirPath := filepath.Dir(path)
	return os.MkdirAll(dirPath, os.ModePerm)
}

func (s *LogRegistrationServer) Start(wg *sync.WaitGroup) {
	defer wg.Done()
	pipePath := fmt.Sprintf("%s/log-registration.sock", s.ozoneSocketDirPath)

	err := os.Remove(pipePath) // Remove the pipe if it already exists
	if err != nil && !os.IsNotExist(err) {
		log.Fatalf("failed to remove existing pipe: %v", err)
	}
	err = mkfifo(pipePath, 0666) // Create the named pipe
	if err != nil {
		log.Fatalf("failed to create named pipe: %v", err)
	}
	defer os.Remove(pipePath) // Remove the named pipe when done

	listener, err := net.Listen("unix", pipePath)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	defer listener.Close()

	// Create the gRPC server
	grpcServer := grpc.NewServer()

	// Register the server with the gRPC server
	RegisterRegistrationServiceServer(grpcServer, s)

	// Start serving requests
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("failed to serve: %s", err)
	}
}

func (s *LogRegistrationServer) RegisterLogApp(_ context.Context, request *LogAppRegistrationRequest) (*emptypb.Empty, error) {
	pipePath := utils.GetLogPipePath(request.AppId, s.ozoneWorkingDir)

	s.output <- &LogAppDetails{
		Id:       request.AppId,
		PipePath: pipePath,
	}

	return &emptypb.Empty{}, nil
}

// TODO heartbeat from logApp to server to make sure the logApp is still alive
