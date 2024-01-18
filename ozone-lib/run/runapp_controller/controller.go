package runapp_controller

import (
	"github.com/ozone2021/ozone/ozone-lib/config"
	"github.com/ozone2021/ozone/ozone-lib/config/runspec"
	. "github.com/ozone2021/ozone/ozone-lib/run/brpc_log_registration/log_registration_server"
	"github.com/ozone2021/ozone/ozone-lib/run/logapp_update_controller"
	"sync"
)

type RunController struct {
	ozoneContext           string
	ozoneWorkingDir        string
	ozoneConfig            *config.OzoneConfig
	ui                     *RunCmdBubbleteaApp
	server                 *LogRegistrationServer
	inputLogAppDetailsChan chan *LogAppDetails
	channelHandlerShutdown chan struct{}
	connectedLogApps       map[string]*LogAppDetails
	logUpdateController    *logapp_update_controller.LogappUpdateController
	runResult              *runspec.RunResult
}

func NewRunController(ozoneContext, ozoneWorkingDir, combinedArgs string, ozoneConfig *config.OzoneConfig) *RunController {
	inputLogAppDetailsChan := make(chan *LogAppDetails)

	runResult := runspec.NewRunResult()
	logUpdateController := logapp_update_controller.NewLogappUpdateController(ozoneWorkingDir, inputLogAppDetailsChan, runResult.UpdateListeners)

	runResult.AddListener(logUpdateController.UpdateLogApps)

	return &RunController{
		ozoneContext:        ozoneContext,
		ozoneWorkingDir:     ozoneWorkingDir,
		ozoneConfig:         ozoneConfig,
		ui:                  NewRunCmdBubbleteaApp(combinedArgs, runResult),
		server:              NewLogRegistrationServer(ozoneWorkingDir, inputLogAppDetailsChan),
		logUpdateController: logUpdateController,
		runResult:           runResult,
	}
}

func (c *RunController) Start(wg *sync.WaitGroup) {
	go c.logUpdateController.Start()
	wg.Add(2)
	go c.server.Start(wg)
	go c.ui.Run(wg)
}

func (c *RunController) Run(runnables []*config.Runnable) {
	spec := runspec.NewRunspec(c.ozoneContext, c.ozoneWorkingDir, c.ozoneConfig)
	spec.AddCallstacks(runnables, c.ozoneConfig, c.ozoneContext)

	c.ui.FinishedAddingCallstacks()

	spec.ExecuteCallstacks(c.runResult)
}
