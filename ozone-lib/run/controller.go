package run

import (
	"github.com/ozone2021/ozone/ozone-lib/config"
	"github.com/ozone2021/ozone/ozone-lib/runspec"
)

type RunController struct {
	ozoneContext    string
	ozoneWorkingDir string
	ozoneConfig     *config.OzoneConfig
}

func NewRunController(ozoneContext, ozoneWorkingDir string, ozoneConfig *config.OzoneConfig) *RunController {
	return &RunController{
		ozoneContext:    ozoneContext,
		ozoneWorkingDir: ozoneWorkingDir,
		ozoneConfig:     ozoneConfig,
	}
}

func (c *RunController) Run(runnables []*config.Runnable) {
	spec := runspec.NewRunspec(c.ozoneContext, c.ozoneWorkingDir, c.ozoneConfig)
	spec.AddCallstacks(runnables, c.ozoneConfig, c.ozoneContext)
}
