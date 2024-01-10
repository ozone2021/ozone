package cli

import (
	"fmt"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	process_manager_client "github.com/ozone2021/ozone/ozone-daemon-lib/process-manager-client"
	"github.com/ozone2021/ozone/ozone-lib/brpc_log_registration/log_registration_client"
	"github.com/spf13/cobra"
	"log"
	"os"
	"os/exec"
	"sync"
)

func init() {
	rootCmd.AddCommand(logsCmd)
}

func logs(service string) {
	err, tempDir := process_manager_client.FetchTempDir()

	if err != nil {
		fmt.Println(err)
		return
	}

	logsPath := fmt.Sprintf("%s/%s-logs", tempDir, service)

	cmd := exec.Command("tail", "-f", logsPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stdout
	cmd.Run()
}

type logAppModel struct {
	appId           string
	appUnixPipePath string
	spinner         spinner.Model
}

func initialLogModel(appId, appUnixPipePath string) logAppModel {
	return logAppModel{
		appId:   appId,
		spinner: spinner.New(spinner.WithSpinner(spinner.Dot)),
	}
}

func (m logAppModel) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m logAppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	default:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m logAppModel) View() string {
	return fmt.Sprintf("AppId, %s!\n\n%s", m.appId, m.spinner.View())
}

var logsCmd = &cobra.Command{
	Use:  "logs",
	Long: `Logs for given services`,
	Run: func(cmd *cobra.Command, args []string) {

		//service := args[0]
		//
		//if config.DeploysHasService(service) {
		//	//fmt.Printf("Logs for %s ...\n", service)
		//	//logs(service)
		//} else {
		//	log.Fatalf("No deploy with service name: %s \n", service)
		//}

		client := log_registration_client.NewLogClient()

		client.Connect(ozoneWorkingDir)

		registerAppResponse, err := client.RegisterLogApp()

		if err != nil {
			log.Fatalln(err)
		}
		p := tea.NewProgram(initialLogModel(registerAppResponse.AppId, registerAppResponse.AppUnixPipePath))

		var wg sync.WaitGroup
		go func() {
			defer wg.Done()
			wg.Add(1)
			if _, err := p.Run(); err != nil {
				fmt.Printf("Alas, there's been an error: %v", err)
				os.Exit(1)
			}
		}()

		wg.Wait()

		//_go.Build("microA", "micro-a", "main.go")
		//executable.Build("microA")
	},
}
