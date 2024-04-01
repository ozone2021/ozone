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
	connected           bool
	ozoneWorkingDir     string
	appIdUUID           uuid.UUID
	updatePipePath      string
	server              *LogServer
	heartbeatChan       chan int64
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
		connected:       false,
		server:          NewLogServer(updatePipePath, uiChan, reconnectChan),
		updatePipePath:  updatePipePath,
		appIdUUID:       appIdUUID,
		ozoneWorkingDir: ozoneWorkingDir,
		heartbeatChan:   reconnectChan,
		ui:              NewLogBubbleteaApp(appIdUUID.String(), uiChan),
		uiChan:          uiChan,
	}
}

func (c *LogAppController) RegisterLogApp() {
	client := log_registration_client_service.NewLogClient()

	client.Connect(c.ozoneWorkingDir)

	_, err := client.RegisterLogApp(c.appIdUUID.String())

	if err == nil {
		c.SetConnected(true)
	}
}

func (c *LogAppController) Start() {
	go c.server.Start()
	time.Sleep(2 * time.Second)
	c.RegisterLogApp()
	go c.CheckHeartbeat()
	if c.connected == false {
		go c.KeepConnecting()
	}
	c.ui.Run()
}

func (c *LogAppController) KeepConnecting() {
	for c.connected == false {
		c.RegisterLogApp()
		time.Sleep(2 * time.Second)
	}
}

func (c *LogAppController) SetConnected(connected bool) {
	c.connected = connected

	if c.ui.IsRunning() {
		c.uiChan <- ConnectedMessage{
			Connected: c.connected,
		}
	}
}

func (c *LogAppController) CheckHeartbeat() {
	for {
		select {
		case mostRecentHeartbeat := <-c.heartbeatChan:
			c.mostRecentHeartbeat = mostRecentHeartbeat
			c.SetConnected(true)
		default:
			if c.mostRecentHeartbeat == 0 {
				continue
			}
			currentTime := time.Now().Unix()
			if currentTime-c.mostRecentHeartbeat > 2 {
				c.SetConnected(false)
				c.RegisterLogApp()
			}
			time.Sleep(1 * time.Second)
		}
	}
}
