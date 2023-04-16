package worktree

import (
	"github.com/ozone2021/ozone/ozone-lib/config"
	. "github.com/ozone2021/ozone/ozone-lib/config/config_variable"
	"gopkg.in/yaml.v3"
	"log"
	"os"
)

type WorktreeStep struct {
	Steps   []*config.Step `yaml:"steps"`
	WithEnv []string       `yaml:"with_env"`
}

type WorktreeConditionals struct {
	Satisfied     bool     `yaml:"satisfied"`
	WhenScript    []string `yaml:"when_script"`
	WhenNotScript []string `yaml:"when_not_script"`
}

//	ContextConditionals []*ContextConditional `yaml:"context_conditionals"` # TODO save whether satisified
//  Steps is the depends and contextSteps merged
type WorktreeRunnable struct {
	Name         string               `yaml:"name"`
	Service      string               `yaml:"service"`
	Dir          string               `yaml:"dir"`
	SourceFiles  []string             `yaml:"source_files"`
	Env          VariableMap          `yaml:"context_envs"`
	Conditionals WorktreeConditionals `yaml:"conditionals"`
	Steps        []*WorktreeRunnable  `yaml:"steps"`
	Type         config.RunnableType  `yaml:"RunnableType"`
}

type CallStack struct {
	Hash         string           `yaml:"hash"`
	RootRunnable WorktreeRunnable `yaml:"root_runnable"`
}

type Worktree struct {
	ProjectName string       `yaml:"project"`
	Context     string       `yaml:"context"`
	WorkDir     string       `yaml:"work_dir"`
	BuildVars   *VariableMap `yaml:"build_vars"`
	CallStacks  []*CallStack `yaml:"call_stack"`
}

func NewWorktree(context, ozoneWorkingDir string, config *config.OzoneConfig) *Worktree {
	systemEnvVars := OSEnvToVarsMap()

	renderedBuildVars := config.BuildVars
	renderedBuildVars.RenderNoMerge(systemEnvVars)
	renderedBuildVars.SelfRender()

	worktree := &Worktree{
		ProjectName: config.ProjectName,
		Context:     context,
		WorkDir:     ozoneWorkingDir,
		BuildVars:   renderedBuildVars,
	}

	return worktree
}

func (wt *Worktree) AddRunnable(name string) {

}

func (wt *Worktree) PrintWorktree() {
	log.SetFlags(log.Flags() &^ (log.Ldate | log.Ltime))

	//indent := 0
	//
	//PrintWithIndent(fmt.Sprintf("Project: %s", wt.ProjectName), indent)
	//PrintWithIndent(fmt.Sprintf("Context: %s", wt.Context), indent)
	//PrintWithIndent(fmt.Sprintf("Buildvars: "), indent)
	//wt.BuildVars.Print(indent)
	//fmt.Sprintf("Project: %s\n", wt.ProjectName)
	//fmt.Sprintf("Context: %s\n", wt.ProjectName)
	//b, _ := yaml.Marshal(wt)

	yamlEncoder := yaml.NewEncoder(os.Stdout)
	yamlEncoder.SetIndent(2) // this is what you're looking for
	yamlEncoder.Encode(&wt)
}
