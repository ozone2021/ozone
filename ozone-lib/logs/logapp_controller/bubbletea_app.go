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
	program         *tea.Program
}

type RunResultUpdate struct {
	*runspec.RunResult
}

func NewLogBubbleteaApp(appId string) *LogBubbleteaApp {
	app := &LogBubbleteaApp{
		appId:   appId,
		spinner: spinner.New(spinner.WithSpinner(spinner.Dot)),
	}
	app.program = tea.NewProgram(app)

	return app
}

func (m LogBubbleteaApp) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m LogBubbleteaApp) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	default:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m LogBubbleteaApp) View() string {
	return fmt.Sprintf("AppId, %s!\n\n%s", m.appId, m.spinner.View())
}

func (m LogBubbleteaApp) ChannelHandler() {
	for {
		select {
		case runResult := <-m.runResultUpdate:
			m.program.Send(RunResultUpdate{
				RunResult: runResult,
			})
		}
	}
}

func (m LogBubbleteaApp) Run() {

	if _, err := m.program.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
