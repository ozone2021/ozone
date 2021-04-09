package process_manager

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"os/exec"
	"strings"
)

type OzoneProcess struct {
	Cmd  *exec.Cmd
	Name string
}

type ProcessCreateQuery struct {
	Name              string
	ProcessWorkingDir string
	OzoneWorkingDir   string
	Cmd               string
	Synchronous		  bool
	Env               map[string]string
}

type DebugQuery struct {
	OzoneWorkingDir   string
}

// map[string]process  to map directory Name to the processes
type ProcessManager struct {
	directories	map[string]string // maps working directory to temp dir
	processes 	map[string][]*OzoneProcess
}

func New() *ProcessManager {
	directories := make(map[string]string)
	processes := make(map[string][]*OzoneProcess)
	return &ProcessManager{
		processes: processes,
		directories: directories,
	}
}

func substituteOutput(input string, tempDir string) string {
	result := input
	//outputFolder := fmt.Sprintf("%s/%s", tempDir, query.Name)
	outputFolder := fmt.Sprintf("%s", tempDir)

	result = strings.ReplaceAll(result, "__TEMP__", tempDir)
	result = strings.ReplaceAll(result, "__OUTPUT__", outputFolder)

	return result
}

func (pm *ProcessManager) AddProcess(processQuery *ProcessCreateQuery, reply *int) error {
	tempDir, ok := pm.directories[processQuery.OzoneWorkingDir]

	if !ok {
		tempDir = createTempDir()
		pm.directories[processQuery.OzoneWorkingDir] =  tempDir
	}
	processWorkingDirectory := substituteOutput(processQuery.ProcessWorkingDir, tempDir)

	cmdString := substituteOutput(processQuery.Cmd, tempDir)

	cmdFields := strings.Fields(cmdString)
	var argFields []string
	if len(cmdFields) > 1 {
		argFields = cmdFields[1:]
	}
	cmd := exec.Command(cmdFields[0], argFields...)

	// Create log file
	logFileString := fmt.Sprintf("%s/logs", tempDir)
	var _, err = os.Stat(logFileString)

	var logFile *os.File
	if os.IsNotExist(err) {
		fmt.Println("log file doesn't exist")
		logFile, err = os.Create(logFileString)
		if err != nil {
			fmt.Println(err)
		}
	}
	logFile, err = os.OpenFile(logFileString, os.O_WRONLY, os.ModeAppend)
	if err != nil {
		fmt.Println(err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Println(err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		fmt.Println(err)
	}

	multi := io.MultiReader(stdout, stderr)
	in := bufio.NewScanner(multi)
	cmd.Dir = processWorkingDirectory

	go handleLogs(in, logFile)
	if processQuery.Synchronous {
		err = cmd.Run()
		fmt.Printf("SYNCHRONOUS \n")
	} else {
		err = cmd.Start()
		fmt.Printf("NONBLOCKING \n")
	}
	fmt.Printf("Logs at:\n")
	fmt.Printf("%s \n", logFileString)


	process := &OzoneProcess{
		Cmd:  cmd,
		Name: processQuery.Name,
	}

	if !processQuery.Synchronous {
		pm.processes[processWorkingDirectory] = append(pm.processes[processWorkingDirectory], process)
	}
	fmt.Printf("length: %d \n", len(pm.processes))
	return nil
}

func handleLogs(in *bufio.Scanner, logFile *os.File) {
	for in.Scan() {
		fmt.Printf("SCANNED::  %s \n", in.Text())
		_, err := logFile.WriteString(in.Text() + "\n")
		if err != nil {
			fmt.Println(err)
		}
	}
}

func createTempDir() string {
	dirName, err := ioutil.TempDir("/tmp/", "ozone-")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(dirName)
	return dirName
}

func (pm *ProcessManager) Debug(processQuery *ProcessCreateQuery, reply *int) error {
	for i, v := range pm.processes {
		fmt.Printf("dir %s process name is %s \n", i, v[0].Name)
		v[0].Cmd.Stdout.Write([]byte("please fucking work"))
	}
	return nil
}

func Run() {
	processManager := New()
	rpc.Register(processManager)
	rpc.HandleHTTP()
	l, e := net.Listen("tcp", ":8000")
	if e != nil {
		log.Fatal("listen error:", e)
	}
	fmt.Println("Running daemon...")
	http.Serve(l, nil)
}
