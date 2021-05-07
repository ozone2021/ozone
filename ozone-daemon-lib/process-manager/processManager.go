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
	"time"
)

type OzoneProcess struct {
	Cmd			*exec.Cmd
	Name		string
	StartTime	int64
}

type ContextQueryRequest struct {
	Directory string
	Context string
	DefaultContext string
}
type ContextQueryResponse struct {
	Context string
}

type TempDirRequest struct {
	WorkingDirectory string
}
type TempDirResponse struct {
	TempDir		string
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
	contexts	map[string]string
	directories	map[string]string // maps working directory to temp dir
	processes 	map[string]map[string]*OzoneProcess
}

func New() *ProcessManager {
	contexts := make(map[string]string)
	directories := make(map[string]string)
	processes := make(map[string]map[string]*OzoneProcess)
	return &ProcessManager{
		contexts: contexts,
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

func (pm *ProcessManager) TempDirRequest(request *TempDirRequest, response *TempDirResponse) error {
	log.Printf("Request %s \n", request.WorkingDirectory)
	tempDir, ok := pm.directories[request.WorkingDirectory]
	if ok {
		response.TempDir = tempDir
	}
	return nil
}
func (pm *ProcessManager) Test(request *int, reply *int) error {
	log.Println("here")
	return nil
}

func deleteEmpty (s []string) []string {
	var r []string
	for _, str := range s {
		if str != "" {
			r = append(r, str)
		}
	}
	return r
}

func (pm *ProcessManager) AddProcess(processQuery *ProcessCreateQuery, reply *error) error {
	tempDir, ok := pm.directories[processQuery.OzoneWorkingDir]

	if !ok {
		tempDir = createTempDir()
		pm.directories[processQuery.OzoneWorkingDir] =  tempDir
	}
	processWorkingDirectory := substituteOutput(processQuery.ProcessWorkingDir, tempDir)

	cmdString := substituteOutput(processQuery.Cmd, tempDir)
	log.Println("cmd is:")
	log.Println(processQuery.Cmd)

	cmdFields := strings.Fields(cmdString)
	var argFields []string
	if len(cmdFields) > 1 {
		argFields = cmdFields[1:]
	}
	argFields = deleteEmpty(argFields)
	cmd := exec.Command(cmdFields[0], argFields...)
	cmd.Dir = processWorkingDirectory

	// Create log file
	logFileString := fmt.Sprintf("%s/%s-logs", tempDir, processQuery.Name)
	var _, err = os.Stat(logFileString)

	var logFile *os.File
	if os.IsNotExist(err) {
		fmt.Println("log file doesn't exist")
		logFile, err = os.Create(logFileString)
		if err != nil {
			reply = &err
			fmt.Println(err)
			return err
		}
	}
	logFile, err = os.OpenFile(logFileString, os.O_WRONLY, os.ModeAppend)
	if err != nil {
		reply = &err
		fmt.Println(err)
		return err
	}

	if !processQuery.Synchronous {
		pm.handleAsynchronous(processQuery.Name, cmd, logFile, processWorkingDirectory)
	}

	fmt.Printf("Logs at:\n")
	fmt.Printf("%s \n", logFileString)

	if processQuery.Synchronous {
		err = pm.handleAsynchronous(
			processQuery.Name,
			cmd,
			logFile,
			processWorkingDirectory,
		)
	} else {
		err = pm.handleSynchronous(
			cmd,
			logFile,
		)
	}
	if err != nil {
		fmt.Println(err)
		reply = &err
		return err
	}

	return nil
}

func (pm *ProcessManager) handleSynchronous(cmd *exec.Cmd, logFile *os.File) error {
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	err := cmd.Run()
	if err != nil {
		return err
	}
	fmt.Printf("SYNCHRONOUS \n")

	return nil
}

func (pm *ProcessManager) handleAsynchronous(
	name string,
	cmd *exec.Cmd,
	logFile *os.File,
	processWorkingDirectory string) error {

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil
	}

	multi := io.MultiReader(stdout, stderr)
	in := bufio.NewScanner(multi)

	go handleLogs(in, logFile)

	processMap, ok := pm.processes[processWorkingDirectory]
	if ok {
		process, ok := processMap[name]
		if ok {
			process.Cmd.Process.Kill()
		}
	}

	err = cmd.Start()
	if err != nil {
		return err
	}
	fmt.Printf("NONBLOCKING \n")

	process := &OzoneProcess{
		Cmd:       cmd,
		Name:      name,
		StartTime: time.Now().Unix(),
	}

	_, ok = pm.processes[processWorkingDirectory]
	if !ok {
		pm.processes[processWorkingDirectory] = make(map[string]*OzoneProcess)
	}
	pm.processes[processWorkingDirectory][name] = process

	return nil
}



func (pm *ProcessManager) ContextQuery(contextQuery *ContextQueryRequest, response *ContextQueryResponse) error {
	if contextQuery.Context != "" {
		pm.contexts[contextQuery.Directory] = contextQuery.Context
	} else {
		pm.contexts[contextQuery.Directory] = contextQuery.DefaultContext
	}
	response.Context = pm.contexts[contextQuery.Directory]

	return nil
}

func handleLogs(in *bufio.Scanner, logFile *os.File) {
	for in.Scan() {
		fmt.Printf("SCANNED::  %s \n", in.Text())
		if len(in.Bytes()) != 0 {
			_, err := logFile.WriteString(in.Text())
			if err != nil {
				fmt.Println(err)
			}
		} else {
			_, err := logFile.WriteString("\n")
			if err != nil {
				fmt.Println(err)
			}
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
	for i, pMap := range pm.processes {
		for name, _ := range pMap {
			fmt.Printf("dir %s process name is %s \n", i, name)
		}
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
