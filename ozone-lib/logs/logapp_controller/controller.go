package logapp_controller

import (
	"github.com/google/uuid"
	. "github.com/ozone2021/ozone/ozone-lib/logs/brpc_log_server/log_server"
	"github.com/ozone2021/ozone/ozone-lib/run/brpc_log_registration/log_registration_client_service"
	"github.com/ozone2021/ozone/ozone-lib/utils"
	"log"
	"time"
)

type LogAppController struct {
	ozoneWorkingDir     string
	appIdUUID           uuid.UUID
	updatePipePath      string
	server              *LogServer
	reconnectChan       chan int64
	mostRecentHeartbeat int64
	ui                  *LogBubbleteaApp
	uiChan              UiChan
}

func NewLogAppController(ozoneWorkingDir string) *LogAppController {
	uiChan := make(UiChan)

	appIdUUID, err := uuid.NewUUID()
	if err != nil {
		log.Fatalln(err)
	}

	updatePipePath := utils.GetLogPipePath(appIdUUID.String(), ozoneWorkingDir)

	reconnectChan := make(chan int64)

	return &LogAppController{
		server:          NewLogServer(updatePipePath, uiChan, reconnectChan),
		updatePipePath:  updatePipePath,
		appIdUUID:       appIdUUID,
		ozoneWorkingDir: ozoneWorkingDir,
		reconnectChan:   reconnectChan,
		ui:              NewLogBubbleteaApp(appIdUUID.String(), uiChan),
		uiChan:          uiChan,
	}
}

func (c *LogAppController) RegisterLogApp() bool {
	client := log_registration_client_service.NewLogClient()

	client.Connect(c.ozoneWorkingDir)

	_, err := client.RegisterLogApp(c.appIdUUID.String())

	connected := true
	if err != nil {
		connected = false
	}

	return connected
}

func (c *LogAppController) Start() {
	go c.server.Start()
	time.Sleep(2 * time.Second)
	connected := c.RegisterLogApp()
	go c.CheckHeartbeat()
	c.ui.Run(connected)
}

func (c *LogAppController) CheckHeartbeat() {
	for {
		select {
		case mostRecentHeartbeat := <-c.reconnectChan:
			c.mostRecentHeartbeat = mostRecentHeartbeat
		default:
			if c.mostRecentHeartbeat == 0 {
				continue
			}
			currentTime := time.Now().Unix()
			if currentTime-c.mostRecentHeartbeat > 5 {
				connected := c.RegisterLogApp()
				c.uiChan <- ConnectedMessage{
					Connected: connected,
				}
			}
		}
	}
}
