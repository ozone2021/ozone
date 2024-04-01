package runapp_controller

import (
	"fmt"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	runspec2 "github.com/ozone2021/ozone/ozone-lib/config/runspec"
	"os"
)

type RunCmdBubbleteaApp struct {
	runList          string
	callStacksLoaded bool
	runResult        *runspec2.RunResult // items on the to-do list
	spinner          spinner.Model
	program          *tea.Program
	shutdownChan     chan struct{}
	reRunChan        chan struct{}
}

type RunResultUpdate struct {
	*runspec2.RunResult
}

type FinishedAddingCallstacks struct{}

func NewRunCmdBubbleteaApp(runList string, result *runspec2.RunResult, shutdownChan chan struct{}, reRunChan chan struct{}) *RunCmdBubbleteaApp {
	app := RunCmdBubbleteaApp{
		runList:          runList,
		callStacksLoaded: false,
		spinner:          spinner.New(spinner.WithSpinner(spinner.Dot)),
		runResult:        result,
		shutdownChan:     shutdownChan,
		reRunChan:        reRunChan,
	}
	app.program = tea.NewProgram(&app)

	result.AddListener(app.UpdateRunResult)

	return &app
}

func (m *RunCmdBubbleteaApp) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m *RunCmdBubbleteaApp) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {

	// Is it a key press?
	case tea.KeyMsg:
		msg := msg.(tea.KeyMsg).String()
		switch msg {
		case "ctrl+c", "q":
			m.shutdownChan <- struct{}{}
			return m, tea.Quit
		case "r":
			m.reRunChan <- struct{}{}
		}
	case RunResultUpdate:
		newData := msg.(RunResultUpdate).RunResult
		m.runResult = newData
	case FinishedAddingCallstacks:
		m.callStacksLoaded = true
	default:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	// Return the updated RunCmdBubbleteaApp to the Bubble Tea runtime for processing.
	// Note that we're not returning a command.
	return m, nil
}

func (m *RunCmdBubbleteaApp) View() string {
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
	s += "\n -------------------------------------------------- \n"
	s += "       q: quit      r: rerun\n"

	// Send the UI for rendering
	return s
}

func (m *RunCmdBubbleteaApp) Run() {
	if _, err := m.program.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}

func (m *RunCmdBubbleteaApp) UpdateRunResult(runResult *runspec2.RunResult, _ bool) {
	go m.program.Send(RunResultUpdate{
		RunResult: runResult,
	})
}

func (m *RunCmdBubbleteaApp) FinishedAddingCallstacks() {
	m.program.Send(FinishedAddingCallstacks{})
}
