package log_registration_client_service

import (
	"context"
	"fmt"
	. "github.com/ozone2021/ozone/ozone-lib/run/brpc_log_registration/log_registration_pb"
	"github.com/ozone2021/ozone/ozone-lib/utils"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
	"log"
	"path/filepath"
)

type LogRegistrationService struct {
	registrationServiceClient RegistrationServiceClient
}

func NewLogClient() *LogRegistrationService {
	return &LogRegistrationService{}
}

func (c *LogRegistrationService) Connect(ozoneWorkingDir string) {
	registrationSock := filepath.Join(utils.GetTmpDir(ozoneWorkingDir), "socks/log-registration.sock")

	pipePath := fmt.Sprintf("unix://%s", registrationSock)

	conn, err := grpc.Dial(pipePath, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}

	c.registrationServiceClient = NewRegistrationServiceClient(conn)
}

func (c *LogRegistrationService) RegisterLogApp(appUuid string) (*emptypb.Empty, error) {
	return c.registrationServiceClient.RegisterLogApp(context.Background(), &LogAppRegistrationRequest{
		AppId: appUuid,
	})
}
