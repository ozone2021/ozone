package logapp_controller

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/ozone2021/ozone/ozone-lib/config/runspec"
	"github.com/ozone2021/ozone/ozone-lib/logs/brpc_log_server/log_server"
	"github.com/ozone2021/ozone/ozone-lib/run/brpc_log_registration/log_registration_client_service"
	"github.com/ozone2021/ozone/ozone-lib/utils"
	"log"
	"path/filepath"
)

type LogAppController struct {
	ozoneWorkingDir string
	appIdUUID       uuid.UUID
	updatePipePath  string
	server          *log_server.LogServer
	ui              *LogBubbleteaApp
}

func NewLogAppController(ozoneWorkingDir string) *LogAppController {
	updateChan := make(chan *runspec.RunResult)

	appIdUUID, err := uuid.NewUUID()
	if err != nil {
		log.Fatalln(err)
	}
	trimmedUUID := appIdUUID.String()[:6]

	updatePipePath := filepath.Join(utils.GetTmpDir(ozoneWorkingDir), "socks", fmt.Sprintf("log-%s.sock", trimmedUUID))

	return &LogAppController{
		server:          log_server.NewLogServer(updatePipePath, updateChan),
		updatePipePath:  updatePipePath,
		appIdUUID:       appIdUUID,
		ozoneWorkingDir: ozoneWorkingDir,
		ui:              NewLogBubbleteaApp(appIdUUID.String()),
	}
}

func (c *LogAppController) registerLogApp() {
	client := log_registration_client_service.NewLogClient()

	client.Connect(c.ozoneWorkingDir)

	_, err := client.RegisterLogApp(c.appIdUUID.String())

	if err != nil {
		log.Fatalln(err)
	}
}

func (c *LogAppController) Start() {
	go c.server.Start()
	c.ui.Run()
}
