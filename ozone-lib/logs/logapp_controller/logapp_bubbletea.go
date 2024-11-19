package logapp_controller

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/fatih/color"
	"github.com/muesli/reflow/wordwrap"
	"github.com/ozone2021/ozone/ozone-lib/config/runspec"
	. "github.com/ozone2021/ozone/ozone-lib/logs/brpc_log_server/log_server"
	"io"
	"log"
	"os"
	"strings"
	"sync"
)

const useHighPerformanceRenderer = false

var (
	titleStyle = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Right = "├"
		return lipgloss.NewStyle().BorderStyle(b).Padding(0, 1)
	}()

	infoStyle = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Left = "┤"
		return titleStyle.Copy().BorderStyle(b)
	}()
)

type FollowMode int

const (
	OFF FollowMode = iota
	FOLLOW_CURRENT
	FOLLOW_ALL
)

type LogBubbleteaApp struct {
	appId                  string
	isRunning, logsShowing bool
	ready                  bool
	connected              bool

	spinner    spinner.Model
	program    *tea.Program
	followMode FollowMode
	viewport   viewport.Model
	keyMap     KeyMap

	input UiChan

	logsShownAtLeastOnce        bool
	runId                       string
	logOutput                   string
	logStopChan                 chan struct{}
	logMutex                    sync.Mutex
	runResult                   *runspec.RunResult
	runResultMutex              sync.Mutex
	selectedCallstackResultNode *runspec.CallstackResultNode
}

type UiMsg interface{}

type ConnectedMessage struct {
	Connected bool
}

type LogLineUpdate struct {
	ClearOutput bool
	Line        string
}

func NewLogBubbleteaApp(appId string, uiChan UiChan) *LogBubbleteaApp {
	app := &LogBubbleteaApp{
		connected:   false,
		appId:       appId,
		spinner:     spinner.New(spinner.WithSpinner(spinner.Dot)),
		input:       uiChan,
		followMode:  FOLLOW_ALL,
		logStopChan: make(chan struct{}, 1),
		keyMap:      LogKeyMap(),
	}
	app.program = tea.NewProgram(app, tea.WithMouseCellMotion(), tea.WithAltScreen())

	return app
}

func (m *LogBubbleteaApp) ResetLogBubbleteaApp() {
	m.selectedCallstackResultNode = nil
}

func (m *LogBubbleteaApp) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m *LogBubbleteaApp) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)
	switch msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg.(tea.KeyMsg), m.keyMap.Follow):
			m.followMode = FOLLOW_ALL
		case key.Matches(msg.(tea.KeyMsg), viewport.DefaultKeyMap().Up):
			m.followMode = OFF
		case key.Matches(msg.(tea.KeyMsg), m.keyMap.NextLog):
			ok := m.moveToNextSelection()
			if !ok {
				return m, nil
			}
		case key.Matches(msg.(tea.KeyMsg), m.keyMap.PageDown):
			m.viewport.GotoBottom()
		}
	case tea.WindowSizeMsg:
		headerHeight := lipgloss.Height(m.headerView())
		footerHeight := lipgloss.Height(m.footerView())
		verticalMarginHeight := headerHeight + footerHeight

		if !m.ready {
			// Since this program is using the full size of the viewport we
			// need to wait until we've received the window dimensions before
			// we can initialize the viewport. The initial dimensions come in
			// quickly, though asynchronously, which is why we wait for them
			// here.
			m.viewport = viewport.New(msg.(tea.WindowSizeMsg).Width, msg.(tea.WindowSizeMsg).Height-verticalMarginHeight)
			m.viewport.YPosition = headerHeight
			m.viewport.HighPerformanceRendering = useHighPerformanceRenderer

			m.ready = true

			// This is only necessary for high performance rendering, which in
			// most cases you won't need.
			//
			// Render the viewport one line below the header.
			m.viewport.YPosition = headerHeight + 1
		} else {
			m.viewport.Width = msg.(tea.WindowSizeMsg).Width
			m.viewport.Height = msg.(tea.WindowSizeMsg).Height - verticalMarginHeight
		}

		m.setContent()

		if useHighPerformanceRenderer {
			// Render (or re-render) the whole viewport. Necessary both to
			// initialize the viewport and when the window is resized.
			//
			// This is needed for high-performance rendering only.
			cmds = append(cmds, viewport.Sync(m.viewport))
		}
	case *RunResultUpdate:
		m.runResultMutex.Lock()
		defer m.runResultMutex.Unlock()

		runResultUpdate := msg.(*RunResultUpdate)
		runResult := runResultUpdate.RunResult
		m.runResult = runResult
		// This is for when the log app is left running and then a new run is started from CLI, not by pressing "r"
		// in run app.
		if m.runResult == nil || runResultUpdate.RunId != m.runResult.RunId {
			m.selectedCallstackResultNode = nil
			ok := m.moveToNextSelection()
			if !ok {
				return m, nil
			}
		} else {
			// This handles resets triggered by the run app
			if runResultUpdate.ShouldReset == true {
				m.ResetLogBubbleteaApp()
				go m.ShowLogs()
			} else {
				diffNode, ok := diff(m.runResult, runResult)
				if ok == false {
					break
				}

				if diffNode.Status == runspec.NotStarted {
					break
				}

				if m.selectedCallstackResultNode == nil || m.logsShownAtLeastOnce == false {
					m.selectedCallstackResultNode = diffNode
					go m.ShowLogs()
				}
			}
		}
		return m, nil
	case ConnectedMessage:
		m.connected = msg.(ConnectedMessage).Connected
		if m.connected && m.logsShownAtLeastOnce == false {
			m.moveToNextSelection()
		}
		return m, nil
	case LogLineUpdate:
		updateMsg := msg.(LogLineUpdate)
		logLine := updateMsg.Line
		if updateMsg.ClearOutput == true {
			m.logOutput = ""
		}

		m.logOutput = m.logOutput + logLine
		m.setContent()

		m.logsShownAtLeastOnce = true
		if m.followMode != OFF {
			m.viewport.GotoBottom()
		}
		return m, nil
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m *LogBubbleteaApp) setContent() {
	m.viewport.SetContent(wordwrap.String(m.logOutput, m.viewport.Width))
}

func (m *LogBubbleteaApp) moveToNextSelection() bool {
	defer func() { go m.ShowLogs() }()
	next, err := m.NextSelection()
	if err != nil {
		return false
	}
	m.selectedCallstackResultNode = next

	return true
}

func (m *LogBubbleteaApp) headerView() string {
	titleString := "Ozone logs: "
	if m.connected {
		statusColor := color.New(color.FgGreen).SprintFunc()
		titleString += statusColor("Connected")
	} else {
		statusColor := color.New(color.FgRed).SprintFunc()
		titleString += statusColor("Disconnected")
	}
	title := titleStyle.Render(titleString)

	line := strings.Repeat("─", max(0, m.viewport.Width-lipgloss.Width(title)))
	return lipgloss.JoinHorizontal(lipgloss.Center, title, line)
}

func (m *LogBubbleteaApp) footerView() string {
	info := infoStyle.Render(fmt.Sprintf("%3.f%%", m.viewport.ScrollPercent()*100))
	line := strings.Repeat("─", max(0, m.viewport.Width-lipgloss.Width(info)))
	line = lipgloss.JoinHorizontal(lipgloss.Center, line, info)

	if m.selectedCallstackResultNode != nil {
		line += "\n" + m.selectedCallstackResultNode.Name
		line += "\n logFile: " + m.selectedCallstackResultNode.LogFile
	}
	return line
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

type PredicateFunc func(*runspec.CallstackResultNode) bool

func (m *LogBubbleteaApp) Next(startId *string) (*runspec.CallstackResultNode, error) {
	if m.runResult == nil {
		return nil, errors.New("No run result")
	}
	if m.runResult.Index.Len() == 0 {
		return nil, errors.New("No nodes")
	}
	if startId == nil {
		return m.runResult.Index.Front().Value, nil
	}
	if m.runResult.Index.Back().Value.Id == *startId {
		return m.runResult.Index.Front().Value, nil
	}
	el := m.runResult.Index.GetElement(*startId)
	if el != nil {
		return el.Next().Value, nil
	}
	return nil, errors.New("Couldn't find next node")
}

func (m *LogBubbleteaApp) Previous() (*runspec.CallstackResultNode, error) {
	if m.runResult.Index.Front().Value.Id == m.selectedCallstackResultNode.Id {
		return m.runResult.Index.Back().Value, nil
	}
	el := m.runResult.Index.GetElement(m.selectedCallstackResultNode.Id)
	if el != nil {
		return el.Prev().Value, nil
	}
	return nil, errors.New("Couldn't find previous node")
}

func (m *LogBubbleteaApp) nextLogPredicate(node *runspec.CallstackResultNode) bool {
	return node.IsCallstack
}

func (m *LogBubbleteaApp) NextSelection() (*runspec.CallstackResultNode, error) {
	if m.selectedCallstackResultNode == nil {
		return m.Next(nil)
	}
	next, err := m.Next(&m.selectedCallstackResultNode.Id)
	if err != nil {
		return nil, err
	}
	count := 0
	for count < m.runResult.Index.Len() {
		if m.selectedCallstackResultNode.LogFile != next.LogFile {
			return next, nil
		}
		next, err = m.Next(&next.Id)
		if err != nil {
			return nil, err
		}
		count++
	}

	return nil, errors.New("Couldn't find next selection for predicate")
}

func (m *LogBubbleteaApp) PreviousSelection() (*runspec.CallstackResultNode, error) {
	previous, err := m.Previous()
	if err != nil {
		return nil, err
	}
	count := 0
	for count < m.runResult.Index.Len() {
		if m.selectedCallstackResultNode.LogFile != previous.LogFile {
			return previous, nil
		}
		previous, err = m.Previous()
		if err != nil {
			return nil, err
		}
		count++
	}

	return nil, errors.New("Couldn't find previous selection for predicate")
}

func (m *LogBubbleteaApp) GetSelectedCallstackResultNode() *runspec.CallstackResultNode {
	return m.selectedCallstackResultNode
}

//func (m *LogBubbleteaApp) FollowIfEnabled() {
//	if m.followMode {
//		for _, current := range m.runResult.Index {
//			if current.Status == runspec.Running && current.LogFile != m.selectedCallstackResultNode.LogFile {
//				go m.ShowLogs(current)
//				return
//			}
//		}
//	}
//}

func (m *LogBubbleteaApp) closeLogs() {
	if m.logsShowing {
		m.logStopChan <- struct{}{}
	}
}

func (m *LogBubbleteaApp) ShowLogs() error {
	if m.selectedCallstackResultNode == nil {
		return nil
	}

	file, err := os.Open(m.selectedCallstackResultNode.LogFile)
	if err != nil {
		return err
	}
	defer file.Close()

	if m.logsShownAtLeastOnce {
		m.closeLogs()
		m.logMutex.Lock()
		defer m.logMutex.Unlock()
	}

	reader := bufio.NewReader(file)

	m.program.Send(LogLineUpdate{
		ClearOutput: true,
		Line:        "",
	})
	m.logsShowing = true
	for {
		select {
		case <-m.logStopChan:
			return nil
		default:
			line, err := reader.ReadString('\n')
			if err == io.EOF {
				// Handle end of file
				continue
			} else if err != nil {
				return nil
			}
			m.program.Send(LogLineUpdate{
				ClearOutput: false,
				Line:        line,
			})
		}
	}
	return nil
}

func (m *LogBubbleteaApp) FollowMode() FollowMode {
	return m.followMode
}

func (m *LogBubbleteaApp) View() string {
	s := ""
	s += fmt.Sprintf("AppId, %s!\n\n%s", m.appId, m.spinner.View())

	if m.runResult == nil {
		s += fmt.Sprintf("\n\n   %s Loading... \n\n", m.spinner.View())
	} else {
		return fmt.Sprintf("%s\n%s\n%s", m.headerView(), m.viewport.View(), m.footerView())
	}

	// The footer
	s += "\nPress q to quit.\n"

	// Send the UI for rendering
	return s
}

// node1 old... node2 new
func diff(root1 *runspec.RunResult, root2 *runspec.RunResult) (*runspec.CallstackResultNode, bool) {
	if root1 == nil {
		return root2.Roots[0], true
	}
	for _, key1 := range root1.Index.Keys() {
		node1, _ := root1.Index.Get(key1)
		for _, key2 := range root2.Index.Keys() {
			node2, _ := root2.Index.Get(key2)
			if node1.Id == node2.Id {
				if node1.Status != node2.Status {
					return node2, true
				}
			}
		}
	}

	return nil, false
}

func (m *LogBubbleteaApp) ChannelHandler() {
	for {
		select {
		case message := <-m.input:
			switch message.(type) {
			case *RunResultUpdate, ConnectedMessage:
				m.program.Send(message)
			default:
				log.Fatalf("Unknown message type: %T", message)
			}
		}
	}
}

func (m *LogBubbleteaApp) IsRunning() bool {
	return m.isRunning
}

func (m *LogBubbleteaApp) Run() {
	go m.ChannelHandler()

	m.isRunning = true
	if _, err := m.program.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
