package log_registration_server

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	. "github.com/ozone2021/ozone/ozone-lib/brpc_log_registration/log_registration_pb"
	"github.com/ozone2021/ozone/ozone-lib/utils"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
	"log"
	"net"
	"os"
	"path/filepath"
)

type LogRegistrationServer struct {
	UnimplementedRegistrationServiceServer
	ozoneSocketDirPath string
	registeredLogApps  map[string]*LogAppDetails
}

type LogAppDetails struct {
	Id       string
	PipePath string
}

func NewLogRegistrationServer(ozoneWorkingDir string) *LogRegistrationServer {
	return &LogRegistrationServer{
		ozoneSocketDirPath: filepath.Join(utils.GetTmpDir(ozoneWorkingDir), "socks"),
		registeredLogApps:  make(map[string]*LogAppDetails),
	}
}

func mkfifo(path string, mode os.FileMode) error {
	dirPath := filepath.Dir(path)
	return os.MkdirAll(dirPath, os.ModePerm)
	//if err != nil {
	//	fmt.Println("Error creating directory:", err)
	//} else {
	//	fmt.Println("Directories created successfully!")
	//}
	//return syscall.Mkfifo(path, uint32(mode))
}

func (s *LogRegistrationServer) Start() {
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

func (s *LogRegistrationServer) RegisterLogApp(context context.Context, _ *emptypb.Empty) (*LogAppRegistrationResponse, error) {
	appIdUUID, err := uuid.NewUUID()
	if err != nil {
		return nil, err
	}
	appId := appIdUUID.String()
	pipePath := fmt.Sprintf("%s/log-app-%s.sock", s.ozoneSocketDirPath, appId)

	s.registeredLogApps[appId] = &LogAppDetails{
		Id:       appId,
		PipePath: pipePath,
	}

	return &LogAppRegistrationResponse{
		AppId:           appId,
		AppUnixPipePath: pipePath,
	}, nil
}

// TODO heartbeat from logApp to server to make sure the logApp is still alive
