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
	"syscall"
	"time"
)

type OzoneProcess struct {
	Cmd			*exec.Cmd
	Name		string
	StartTime	int64
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

type ContextSetQuery struct {
	OzoneWorkingDir 	string
	Context				string
}

type DirQuery struct {
	OzoneWorkingDir   string
}

type StringReply struct {
	Body string
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

func (pm *ProcessManager) TempDirRequest(request *DirQuery, response *StringReply) error {
	log.Printf("Request %s \n", request.OzoneWorkingDir)
	tempDir := pm.createTempDirIfNotExists(request.OzoneWorkingDir)
	log.Println(tempDir)
	response.Body = tempDir
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

func (pm *ProcessManager) Halt(haltQuery *DirQuery, reply *error) error {
	dir := haltQuery.OzoneWorkingDir
	for _, process := range pm.processes[dir] {
		pgid, err := syscall.Getpgid(process.Cmd.Process.Pid)
		if err != nil {
			reply = &err
			return err
		}

		syscall.Kill(-pgid, 15)  // note the minus sign
		process.Cmd.Wait()
	}
	pm.processes[dir] = make(map[string]*OzoneProcess)

	return nil
}

func (pm *ProcessManager) FetchContext(dirQuery *DirQuery, reply *StringReply) error {
	context, ok := pm.contexts[dirQuery.OzoneWorkingDir]

	if ok {
		reply.Body = context
	}
	return nil
}

func (pm *ProcessManager) SetContext(contextSetQuery *ContextSetQuery, reply *StringReply) error {
	pm.contexts[contextSetQuery.OzoneWorkingDir] = contextSetQuery.Context

	return nil
}

func (pm *ProcessManager) Status(dirQuery *DirQuery, reply *StringReply) error {
	dir := dirQuery.OzoneWorkingDir

	if len(pm.processes[dir]) == 0 {
		reply.Body = fmt.Sprintf("No processes running for this workspace.")
		return nil
	}

	reply.Body = fmt.Sprintf("Service \tStatus \n\n")
	for name, process := range pm.processes[dir] {
		running := process.Cmd.ProcessState.ExitCode() == -1
		runningString := "running"
		if !running {
			runningString = fmt.Sprintf("exited code: %d", process.Cmd.ProcessState.ExitCode())
		}
		reply.Body = fmt.Sprintf("%s%s\t%s\n", reply.Body, name, runningString)
	}
	fmt.Println(reply.Body)
	return nil
}

func (pm *ProcessManager) createTempDirIfNotExists(ozoneWorkingDir string) string {
	dirName, ok := pm.directories[ozoneWorkingDir]
	if ok {
		return dirName
	}

	dirName, err := ioutil.TempDir("/tmp/", "ozone-")
	if err != nil {
		log.Fatal(err)
	}

	pm.directories[ozoneWorkingDir] = dirName
	return dirName
}

func (pm *ProcessManager) AddProcess(processQuery *ProcessCreateQuery, reply *error) error {
	tempDir := pm.createTempDirIfNotExists(processQuery.OzoneWorkingDir)
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
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Dir = processWorkingDirectory

	// Create log file
	logFileString := fmt.Sprintf("%s/%s-logs", tempDir, processQuery.Name)
	err, logFile := CreateLogFileIfNotExists(logFileString)
	if err != nil {
		reply = &err
		return err
	}
	logFile, err = os.OpenFile(logFileString, os.O_WRONLY, os.ModeAppend)
	if err != nil {
		reply = &err
		fmt.Println(err)
		return err
	}

	//if !processQuery.Synchronous {
	//	pm.handleAsynchronous(processQuery.Name, cmd, logFile, processWorkingDirectory)
	//}

	fmt.Printf("Logs at:\n")
	fmt.Printf("%s \n", logFileString)

	if processQuery.Synchronous {
		fmt.Println("sync")
		err = pm.handleSynchronous(
			cmd,
			logFile,
		)
	} else {
		fmt.Println("Async")
		err = pm.handleAsynchronous(
			processQuery.Name,
			cmd,
			logFile,
			processWorkingDirectory,
		)
	}
	if err != nil {
		fmt.Println("error")
		fmt.Println(err)
		reply = &err
		return err
	}

	fmt.Println("return nil")

	return nil
}

func CreateLogFileIfNotExists(logFileString string) (error, *os.File) {
	_, err := os.Stat(logFileString)

	var logFile *os.File
	if os.IsNotExist(err) {
		logFile, err = os.Create(logFileString)
		if err != nil {
			return err, nil
		}
	}
	fmt.Println("Created log file")
	return nil, logFile
}

func (pm *ProcessManager) handleSynchronous(cmd *exec.Cmd, logFile *os.File) error {
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	fmt.Printf("SYNCHRONOUS \n")
	err := cmd.Run()
	if err != nil {
		fmt.Println("errorhandleSynchronous")
		return err
	}

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

func handleLogs(in *bufio.Scanner, logFile *os.File) {
	for in.Scan() {
		fmt.Printf("SCANNED::  %s \n", in.Text())
		if len(in.Bytes()) != 0 {
			logLine := fmt.Sprintf("%s\n", in.Text())
			_, err := logFile.WriteString(logLine)
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
