package worktree

import (
	"github.com/ozone2021/ozone/ozone-lib/config"
	"github.com/ozone2021/ozone/ozone-lib/config/config_utils"
	. "github.com/ozone2021/ozone/ozone-lib/config/config_variable"
	"github.com/ozone2021/ozone/ozone-lib/utilities"
	"log"
)

type WorktreeConditionals struct {
	empty         bool
	Satisfied     bool            `yaml:"satisfied"`
	WhenScript    map[string]bool `yaml:"when_script"`
	WhenNotScript map[string]bool `yaml:"when_not_script"`
}

func NewWorktreeConditionals() *WorktreeConditionals {
	return &WorktreeConditionals{
		empty:         true,
		Satisfied:     true,
		WhenScript:    make(map[string]bool),
		WhenNotScript: make(map[string]bool),
	}
}

func (wtc *WorktreeConditionals) AddWhenScriptResult(script string, outcome bool) {
	wtc.empty = false
	if outcome == false {
		wtc.Satisfied = false
	}
	wtc.WhenScript[script] = outcome
}

func (wtc *WorktreeConditionals) AddWhenNotScriptResult(script string, outcome bool) {
	wtc.empty = false
	if outcome == false {
		wtc.Satisfied = false
	}
	wtc.WhenNotScript[script] = outcome
}

func ConvertContextConditional(buildScope *VariableMap, configRunnable *config.Runnable, context string) *WorktreeConditionals {
	wtc := NewWorktreeConditionals()

	for _, contextConditional := range configRunnable.ContextConditionals {
		inPattern, err := config_utils.ContextInPattern(context, contextConditional.Context, buildScope)
		if err != nil {
			return nil
		}
		if inPattern {
			// WhenScript
			for _, script := range contextConditional.WhenScript {
				exitCode, err := utilities.RunBashScript(script, buildScope)
				switch exitCode {
				case 0:
					wtc.AddWhenScriptResult(script, true)
				case 3: // TODO document special ozone exit code to force error
					log.Printf("TODO catch these errors inside the conditionals struct: %s \n", err)
				default:
					wtc.AddWhenScriptResult(script, false)
					log.Printf("Not running, contextConditional whenScript not satisfied: %s", script)
				}
			}
			// When Not script
			for _, script := range contextConditional.WhenNotScript {
				exitCode, err := utilities.RunBashScript(script, buildScope)
				switch exitCode {
				case 0:
					wtc.AddWhenNotScriptResult(script, false)
				case 3: // TODO document special ozone exit code
					log.Printf("TODO catch these errors inside the conditionals struct: %s \n", err)
				default:
					wtc.AddWhenNotScriptResult(script, true)
				}
			}
		}
	}

	if wtc.empty {
		return nil
	}

	return wtc
}
