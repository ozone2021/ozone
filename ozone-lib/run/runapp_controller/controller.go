package runapp_controller

import (
	"github.com/ozone2021/ozone/ozone-lib/config"
	"github.com/ozone2021/ozone/ozone-lib/config/runspec"
	. "github.com/ozone2021/ozone/ozone-lib/run/brpc_log_registration/log_registration_server"
	"github.com/ozone2021/ozone/ozone-lib/run/logapp_update_controller"
	"os"
	"time"
)

type RunController struct {
	ozoneContext           string
	ozoneWorkingDir        string
	ozoneConfig            *config.OzoneConfig
	ui                     *RunCmdBubbleteaApp
	server                 *LogRegistrationServer
	inputLogAppDetailsChan chan *LogAppDetails
	reRunChan              chan struct{}
	shutdownChan           chan struct{}
	logUpdateController    *logapp_update_controller.LogappUpdateController
	runResult              *runspec.RunResult

	waitChan chan struct{}

	runnables []*config.Runnable
}

func NewRunController(ozoneContext, ozoneWorkingDir, combinedArgs string, ozoneConfig *config.OzoneConfig) *RunController {
	inputLogAppDetailsChan := make(chan *LogAppDetails)

	runResult := runspec.NewRunResult()
	logUpdateController := logapp_update_controller.NewLogappUpdateController(ozoneWorkingDir, inputLogAppDetailsChan, runResult.UpdateListeners)

	runResult.AddListener(logUpdateController.UpdateLogApps)

	shutdownChan := make(chan struct{})
	reRunChan := make(chan struct{})

	return &RunController{
		ozoneContext:        ozoneContext,
		ozoneWorkingDir:     ozoneWorkingDir,
		ozoneConfig:         ozoneConfig,
		ui:                  NewRunCmdBubbleteaApp(combinedArgs, runResult, shutdownChan, reRunChan),
		server:              NewLogRegistrationServer(ozoneWorkingDir, inputLogAppDetailsChan),
		logUpdateController: logUpdateController,
		reRunChan:           reRunChan,
		shutdownChan:        shutdownChan,
		runResult:           runResult,
		waitChan:            make(chan struct{}, 1),
	}
}

func (c *RunController) HandleReRunMessages() {
	for {
		select {
		case <-c.reRunChan:
			c.waitChan <- struct{}{}
			time.Sleep(2 * time.Second)
			c.runResult.ResetRunResult()
			go c.Run(c.runnables)
		case <-c.shutdownChan:
			c.waitChan <- struct{}{}
			os.Exit(0)
		}
	}
}

func (c *RunController) Start() {
	go c.logUpdateController.Start()
	go c.HandleReRunMessages()

	go c.server.Start()
	c.ui.Run()
}

func (c *RunController) Run(runnables []*config.Runnable) {
	//logger_lib.ClearLogs(c.ozoneWorkingDir)
	c.runResult.ResetRunResult()

	// Required for first run
	c.runnables = runnables

	spec := runspec.NewRunspec(c.ozoneContext, c.ozoneWorkingDir, c.ozoneConfig)
	c.runResult.RunId = spec.GetRunID()
	spec.AddCallstacks(runnables, c.ozoneConfig, c.ozoneContext)

	c.ui.FinishedAddingCallstacks()

	go spec.ExecuteCallstacks(c.runResult)

	<-c.waitChan
}
