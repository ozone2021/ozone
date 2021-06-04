package process_manager

import (
	"bufio"
	"fmt"
	"github.com/JamesArthurHolland/ozone/ozone-daemon-lib/cache"
	"github.com/TwinProduction/go-color"
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
	IgnoreError		  bool
	Env               map[string]string
}

type DebugQuery struct {
	OzoneWorkingDir   string
}

type ContextSetQuery struct {
	OzoneWorkingDir 	string
	Context				string
}

type IgnoreQuery struct {
	OzoneWorkingDir 	string
	Service			 	string
}

type DirQuery struct {
	OzoneWorkingDir   string
}

type CacheUpdateQuery struct {
	OzoneWorkingDir     string
	Runnable            string
	OzoneFileAndDirHash string
}

type StringReply struct {
	Body string
}

type BoolReply struct {
	Body bool
}

// map[string]process  to map directory Name to the processes
type ProcessManager struct {
	cache		*cache.Cache
	ignores		map[string][]string
	contexts	map[string]string
	directories	map[string]string // maps working directory to temp dir
	processes 	map[string]map[string]*OzoneProcess
}

func New() *ProcessManager {
	cache := cache.New()
	contexts := make(map[string]string)
	ignores := make(map[string][]string)
	directories := make(map[string]string)
	processes := make(map[string]map[string]*OzoneProcess)
	return &ProcessManager{
		cache: cache,
		contexts: contexts,
		ignores: ignores,
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

func (pm *ProcessManager) UpdateCache(request *CacheUpdateQuery, response *BoolReply) error {
	didUpdate := pm.cache.Update(request.OzoneWorkingDir, request.Runnable, request.OzoneFileAndDirHash)
	response.Body = didUpdate
	return nil
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

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func remove(s []string, i int) []string {
	s[len(s)-1], s[i] = s[i], s[len(s)-1]
	return s[:len(s)-1]
}

func (pm *ProcessManager) serviceIsIgnored(service, ozoneWorkingDirectory string) bool {
	_, ok := pm.ignores[ozoneWorkingDirectory]
	if ok && contains(pm.ignores[ozoneWorkingDirectory], service) {
		return true
	}
	return false
}

func (pm *ProcessManager) Ignore(query *IgnoreQuery, reply *error) error {
	ozoneWorkingDirectory := query.OzoneWorkingDir
	_, ok := pm.ignores[ozoneWorkingDirectory]
	if !ok {
		pm.ignores[ozoneWorkingDirectory] = []string{query.Service}
	} else if contains(pm.ignores[ozoneWorkingDirectory], query.Service) {
		for k, v := range pm.ignores[ozoneWorkingDirectory] {
			if v == query.Service {
				pm.ignores[ozoneWorkingDirectory] = remove(pm.ignores[ozoneWorkingDirectory], k)
			} else {
				pm.ignores[ozoneWorkingDirectory] = append(pm.ignores[ozoneWorkingDirectory], query.Service)
			}
		}
	}
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
	fmt.Printf("Status for %s \n", dir)

	if len(pm.processes[dir]) == 0 && len(pm.ignores[dir]) == 0 {
		reply.Body = fmt.Sprintf("No processes running for this workspace.")
		return nil
	}

	reply.Body = fmt.Sprintf("Runnable \tStatus \n\n")
	for name, process := range pm.processes[dir] {
		running := process.Cmd.ProcessState.ExitCode() == -1
		runningString := color.Ize(color.Green,"running")
		if !running {
			runningString = fmt.Sprintf("exited code: %d", process.Cmd.ProcessState.ExitCode())
		}
		reply.Body = fmt.Sprintf("%s%s\t\t%s\n", reply.Body, name, runningString)
	}

	for _, name := range pm.ignores[dir] {
		colourOutput := color.Ize(color.Red, "ignored")
		reply.Body = fmt.Sprintf("%s%s\t\t%s", reply.Body, name, colourOutput)
	}
	fmt.Println(reply.Body)
	return nil
}

func (pm *ProcessManager) createTempDirIfNotExists(ozoneWorkingDir string) string {
	dirName, ok := pm.directories[ozoneWorkingDir]
	if ok {
		return dirName
	}

	dirName, err := ioutil.TempDir("/tmp/ozone/", "ozone-")
	if err != nil {
		log.Fatal(err)
	}

	pm.directories[ozoneWorkingDir] = dirName
	return dirName
}

func CommandFromFields(cmdString string) ([]string, []string) {
	cmdFields := strings.Fields(cmdString)
	var argFields []string
	if len(cmdFields) > 1 {
		argFields = cmdFields[1:]
	}
	argFields = deleteEmpty(argFields)
	return cmdFields, argFields
}

func (pm *ProcessManager) AddProcess(processQuery *ProcessCreateQuery, reply *error) error {
	if pm.serviceIsIgnored(processQuery.Name, processQuery.OzoneWorkingDir) {
		return nil
	}
	tempDir := pm.createTempDirIfNotExists(processQuery.OzoneWorkingDir)
	processWorkingDirectory := substituteOutput(processQuery.ProcessWorkingDir, tempDir)

	cmdString := substituteOutput(processQuery.Cmd, tempDir)
	log.Println("cmd is:")
	log.Println(cmdString)

	cmdFields, argFields := CommandFromFields(cmdString)
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
			processQuery.IgnoreError,
		)
	} else {
		fmt.Println("Async")
		err = pm.handleAsynchronous(
			processQuery.Name,
			cmd,
			logFile,
			processQuery.OzoneWorkingDir,
			processQuery.IgnoreError,
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

func (pm *ProcessManager) handleSynchronous(cmd *exec.Cmd, logFile *os.File, ignoreErr bool) error {
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	fmt.Printf("SYNCHRONOUS \n")
	err := cmd.Run()
	if err != nil && !ignoreErr {
		fmt.Println("errorhandleSynchronous")
		return err
	}
	cmd.Wait()

	return nil
}

func (pm *ProcessManager) handleAsynchronous(
	name string,
	cmd *exec.Cmd,
	logFile *os.File,
	ozoneWorkingDirectory string,
	ignoreErr bool) error {

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

	processMap, ok := pm.processes[ozoneWorkingDirectory]
	if ok {
		process, ok := processMap[name]
		if ok {
			process.Cmd.Process.Kill()
		}
	}

	err = cmd.Start()
	if err != nil && !ignoreErr {
		return err
	}
	fmt.Printf("NONBLOCKING \n")

	process := &OzoneProcess{
		Cmd:       cmd,
		Name:      name,
		StartTime: time.Now().Unix(),
	}

	_, ok = pm.processes[ozoneWorkingDirectory]
	if !ok {
		pm.processes[ozoneWorkingDirectory] = make(map[string]*OzoneProcess)
	}
	pm.processes[ozoneWorkingDirectory][name] = process

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
