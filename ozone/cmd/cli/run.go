package cli

import (
	"fmt"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	. "github.com/ozone2021/ozone/ozone-lib/brpc_log_registration/log_registration_server"
	ozoneConfig "github.com/ozone2021/ozone/ozone-lib/config"
	"github.com/ozone2021/ozone/ozone-lib/runspec"
	"github.com/spf13/cobra"
	"log"
	"os"
	"sync"
)

type model struct {
	runList          string
	callStacksLoaded bool
	runResult        *runspec.RunResult // items on the to-do list
	spinner          spinner.Model
}

type RunResultUpdate struct {
	*runspec.RunResult
}

type FinishedAddingCallstacks struct{}

func initialModel(runList string, result *runspec.RunResult) model {
	return model{
		runList:          runList,
		callStacksLoaded: false,
		spinner:          spinner.New(spinner.WithSpinner(spinner.Dot)),
		runResult:        result,
	}
}

func (m model) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {

	// Is it a key press?
	case tea.KeyMsg:
		msg := msg.(tea.KeyMsg).String()
		// Cool, what was the actual key pressed?
		switch msg {
		// These keys should exit the program.
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	case RunResultUpdate:
		newData := msg.(RunResultUpdate).RunResult
		m.runResult = newData
		return m, nil
	case FinishedAddingCallstacks:
		m.callStacksLoaded = true
	default:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	// Return the updated model to the Bubble Tea runtime for processing.
	// Note that we're not returning a command.
	return m, nil
}

func (m model) View() string {
	// The header
	s := fmt.Sprintf("Ozone run build: %s \n\n", m.runList)

	if m.runResult == nil {
		s += fmt.Sprintf("\n\n   %s Loading... \n\n", m.spinner.View())
	} else {
		s += m.runResult.PrintRunResult(false)

		if m.callStacksLoaded == false {
			s += fmt.Sprintf("\n\n   %s Initialising callstacks... \n\n", m.spinner.View())
		}
	}

	// The footer
	s += "\nPress q to quit.\n"

	// Send the UI for rendering
	return s
}

var runCmd = &cobra.Command{
	Use:  "run",
	Long: `Shows a dry run of what is going to be ran.`,
	Run: func(cmd *cobra.Command, args []string) {

		spec := runspec.NewRunspec(ozoneContext, ozoneWorkingDir, config)

		var runnables []*ozoneConfig.Runnable

		combinedArgs := ""
		for _, arg := range args {
			combinedArgs += fmt.Sprintf("%s ", arg)
			if has, runnable := config.FetchRunnable(arg); has == true {
				runnables = append(runnables, runnable)
				continue
			} else {
				log.Fatalf("Config doesn't have runnable: %s \n", arg)
			}
		}

		spec.AddCallstacks(runnables, config, ozoneContext)

		runResult := runspec.NewRunResult()
		p := tea.NewProgram(initialModel(combinedArgs, nil))
		sendRunResultUpdate := func(rr *runspec.RunResult) {
			p.Send(RunResultUpdate{
				RunResult: rr,
			})
		}
		//runResult := spec.RunSpecRootNodeToRunResult(spec.CallStacks[ozoneConfig.BuildType][0])

		var wg sync.WaitGroup
		go func() {
			defer wg.Done()
			wg.Add(1)
			if _, err := p.Run(); err != nil {
				fmt.Printf("Alas, there's been an error: %v", err)
				os.Exit(1)
			}
		}()

		runResult.AddListener(sendRunResultUpdate)

		p.Send(FinishedAddingCallstacks{})

		server := NewLogRegistrationServer(ozoneWorkingDir)

		go func() {
			defer wg.Done()
			wg.Add(1)
			server.Start()
		}()

		spec.ExecuteCallstacks(runResult)

		//runResult.PrintErrorLog()
		//
		//fmt.Println("=================================================")
		//fmt.Println("====================  Run result  ===============")
		//fmt.Println("=================================================")

		wg.Wait()

		//runResult.PrintRunResult(true)
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}
