package runspec

import (
	"github.com/ozone2021/ozone/ozone-lib/config"
	"github.com/ozone2021/ozone/ozone-lib/config/config_utils"
	. "github.com/ozone2021/ozone/ozone-lib/config/config_variable"
	"github.com/ozone2021/ozone/ozone-lib/logger_lib"
	"github.com/ozone2021/ozone/ozone-lib/utilities"
)

type RunspecConditionals struct {
	empty         bool
	Satisfied     bool            `yaml:"satisfied"`
	WhenScript    map[string]bool `yaml:"when_script"`
	WhenNotScript map[string]bool `yaml:"when_not_script"`
}

func NewRunspecConditionals() *RunspecConditionals {
	return &RunspecConditionals{
		empty:         true,
		Satisfied:     true,
		WhenScript:    make(map[string]bool),
		WhenNotScript: make(map[string]bool),
	}
}

func (wtc *RunspecConditionals) AddWhenScriptResult(script string, outcome bool) {
	wtc.empty = false
	if outcome == false {
		wtc.Satisfied = false
	}
	wtc.WhenScript[script] = outcome
}

func (wtc *RunspecConditionals) AddWhenNotScriptResult(script string, outcome bool) {
	wtc.empty = false
	if outcome == false {
		wtc.Satisfied = false
	}
	wtc.WhenNotScript[script] = outcome
}

func ConvertContextConditional(buildScope *VariableMap, configRunnable *config.Runnable, context string, logger *logger_lib.Logger) *RunspecConditionals {
	wtc := NewRunspecConditionals()

	for _, contextConditional := range configRunnable.ContextConditionals {
		inPattern, err := config_utils.ContextInPattern(context, contextConditional.Context, buildScope)
		if err != nil {
			return nil
		}
		if inPattern {
			// WhenScript
			for _, script := range contextConditional.WhenScript {
				exitCode, err := utilities.RunBashScript(script, buildScope, logger)
				switch exitCode {
				case 0:
					wtc.AddWhenScriptResult(script, true)
				case 3: // TODO document special ozone exit code to force error
					logger.Fatalln("Something went wrong with conditional script: %script, err: %s \n", script, err)
				default:
					wtc.AddWhenScriptResult(script, false)
					logger.Printf("Not running, contextConditional whenScript not satisfied: %s", script)
				}
			}
			// When Not script
			for _, script := range contextConditional.WhenNotScript {
				exitCode, err := utilities.RunBashScript(script, buildScope, logger)
				switch exitCode {
				case 0:
					wtc.AddWhenNotScriptResult(script, false)
				case 3: // TODO document special ozone exit code
					wtc.AddWhenNotScriptResult(script, false)
					logger.Fatalln("Something went wrong with conditional script: %script, err: %s \n", script, err)
				default:
					wtc.AddWhenNotScriptResult(script, true)
				}
			}
		}
	}

	return wtc
}
