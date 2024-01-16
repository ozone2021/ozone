package logapp_controller

import (
	"fmt"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/ozone2021/ozone/ozone-lib/config/runspec"
	"os"
)

type LogBubbleteaApp struct {
	appId           string
	spinner         spinner.Model
	runResultUpdate chan *runspec.RunResult
	runResult       *runspec.RunResult
	program         *tea.Program
}

type RunResultUpdate struct {
	*runspec.RunResult
}

func NewLogBubbleteaApp(appId string, runResultUpdate chan *runspec.RunResult) *LogBubbleteaApp {
	app := &LogBubbleteaApp{
		appId:           appId,
		spinner:         spinner.New(spinner.WithSpinner(spinner.Dot)),
		runResultUpdate: runResultUpdate,
	}
	app.program = tea.NewProgram(app)

	return app
}

func (m LogBubbleteaApp) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m LogBubbleteaApp) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case RunResultUpdate:
		newData := msg.(RunResultUpdate).RunResult
		m.runResult = newData
		return m, nil
	default:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m LogBubbleteaApp) View() string {
	s := ""
	s += fmt.Sprintf("AppId, %s!\n\n%s", m.appId, m.spinner.View())

	if m.runResult == nil {
		s += fmt.Sprintf("\n\n   %s Loading... \n\n", m.spinner.View())
	} else {
		s += m.runResult.PrintRunResult(false)
	}

	// The footer
	s += "\nPress q to quit.\n"

	// Send the UI for rendering
	return s
}

func (m LogBubbleteaApp) ChannelHandler() {
	for {
		select {
		case runResult := <-m.runResultUpdate:
			m.program.Send(RunResultUpdate{
				RunResult: runResult,
			})
			// TODO shutdown handle
		}
	}
}

func (m LogBubbleteaApp) Run() {
	go m.ChannelHandler()

	if _, err := m.program.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
