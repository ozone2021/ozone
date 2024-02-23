package runapp_controller

import (
	"context"
	"github.com/ozone2021/ozone/ozone-lib/config"
	"github.com/ozone2021/ozone/ozone-lib/config/runspec"
	. "github.com/ozone2021/ozone/ozone-lib/run/brpc_log_registration/log_registration_server"
	"github.com/ozone2021/ozone/ozone-lib/run/logapp_update_controller"
	"os"
	"sync"
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

	runnables []*config.Runnable
	cancel    context.CancelFunc
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
	}
}

func (c *RunController) HandleReRunMessages() {
	for {
		select {
		case <-c.reRunChan:
			c.cancel()
			time.Sleep(2 * time.Second)
			c.runResult.ResetRunResult()
			c.Run(c.runnables)
		case <-c.shutdownChan:
			os.Exit(0)
		}
	}
}

func (c *RunController) Start(wg *sync.WaitGroup) {
	go c.logUpdateController.Start()
	go c.HandleReRunMessages()

	wg.Add(2)
	go c.server.Start(wg)
	go c.ui.Run(wg)
}

func (c *RunController) Run(runnables []*config.Runnable) {
	c.runnables = runnables

	ctx, cancelContext := context.WithCancel(context.Background())
	c.cancel = cancelContext
	spec := runspec.NewRunspec(c.ozoneContext, c.ozoneWorkingDir, c.ozoneConfig)
	spec.AddCallstacks(runnables, c.ozoneConfig, c.ozoneContext)

	c.ui.FinishedAddingCallstacks()

	spec.ExecuteCallstacks(ctx, c.runResult)

}
