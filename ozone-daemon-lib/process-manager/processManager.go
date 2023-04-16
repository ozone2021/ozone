package process_manager

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"github.com/TwiN/go-color"
	"github.com/ozone2021/ozone/ozone-daemon-lib/cache"
	"github.com/ozone2021/ozone/ozone-daemon-lib/process-manager-queries"
	"github.com/ozone2021/ozone/ozone-lib/config/config_variable"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"
)

type OzoneProcess struct {
	Cmd       *exec.Cmd
	Name      string
	StartTime int64
	Port      int64
}

// map[string]process  to map directory Name to the processes
type ProcessManager struct {
	cache       *cache.Cache
	ignores     map[string][]string
	contexts    map[string]string
	directories map[string]string // maps working directory to temp dir
	processes   map[string]map[string]*OzoneProcess
}

func New() *ProcessManager {
	cache := cache.New()
	contexts := make(map[string]string)
	ignores := make(map[string][]string)
	directories := make(map[string]string)
	processes := make(map[string]map[string]*OzoneProcess)
	return &ProcessManager{
		cache:       cache,
		contexts:    contexts,
		ignores:     ignores,
		processes:   processes,
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

func (pm *ProcessManager) CheckCache(request *process_manager_queries.CacheQuery, response *process_manager_queries.StringReply) error {
	response.Body = pm.cache.Check(request.OzoneWorkingDir, request.RunnableName)
	return nil
}

func (pm *ProcessManager) UpdateCache(request *process_manager_queries.CacheQuery, response *process_manager_queries.BoolReply) error {
	didUpdate := pm.cache.Update(request.OzoneWorkingDir, request.RunnableName, request.OzoneFileAndDirHash)
	response.Body = didUpdate
	return nil
}

func (pm *ProcessManager) TempDirRequest(request *process_manager_queries.DirQuery, response *process_manager_queries.StringReply) error {
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

func deleteEmpty(s []string) []string {
	var r []string
	for _, str := range s {
		if str != "" {
			r = append(r, str)
		}
	}
	return r
}

func (pm *ProcessManager) Halt(haltQuery *process_manager_queries.HaltQuery, reply *error) error {
	dir := haltQuery.OzoneWorkingDir
	for _, process := range pm.processes[dir] {
		if haltQuery.Service == "" || haltQuery.Service == process.Name {
			cmdString := fmt.Sprintf("docker rm -f %s",
				process.Name,
			)
			cmdFields, argFields := CommandFromFields(cmdString)
			cmd := exec.Command(cmdFields[0], argFields...)
			logFile, _, err := pm.setUpLogging(haltQuery.OzoneWorkingDir, process.Name)
			if err != nil {
				reply = &err
				return err
			}
			pm.handleSynchronous(
				cmd,
				logFile,
				true,
			)
		}
	}
	if haltQuery.Service == "" {
		pm.processes[dir] = make(map[string]*OzoneProcess)
	} else {
		delete(pm.processes[dir], haltQuery.Service)
	}

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

func (pm *ProcessManager) Ignore(query *process_manager_queries.IgnoreQuery, reply *error) error {
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

func (pm *ProcessManager) FetchContext(dirQuery *process_manager_queries.DirQuery, reply *process_manager_queries.StringReply) error {
	context, ok := pm.contexts[dirQuery.OzoneWorkingDir]

	if ok {
		reply.Body = context
	}
	return nil
}

func (pm *ProcessManager) SetContext(contextSetQuery *process_manager_queries.ContextSetQuery, reply *process_manager_queries.StringReply) error {
	pm.contexts[contextSetQuery.OzoneWorkingDir] = contextSetQuery.Context

	return nil
}

func (pm *ProcessManager) Status(dirQuery *process_manager_queries.DirQuery, reply *process_manager_queries.StringReply) error {
	dir := dirQuery.OzoneWorkingDir
	fmt.Printf("Status for %s \n", dir)

	if len(pm.processes[dir]) == 0 && len(pm.ignores[dir]) == 0 {
		reply.Body = fmt.Sprintf("No processes running for this workspace.")
		return nil
	}

	buf := new(bytes.Buffer)
	writer := tabwriter.NewWriter(buf, 10, 8, 2, '\t', tabwriter.AlignRight)
	fmt.Fprintln(writer, "service\tstatus\tport")
	for name, process := range pm.processes[dir] {
		running := process.Cmd.ProcessState.ExitCode() == -1
		runningString := color.Ize(color.Green, "running")
		if !running {
			runningString = fmt.Sprintf("exited code: %d", process.Cmd.ProcessState.ExitCode())
		}
		portString := strconv.FormatInt(process.Port, 10)
		fmt.Fprintf(writer, "%s\t%s\t:%s\n", name, runningString, portString)
	}
	writer.Flush()
	reply.Body = buf.String()

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

func (pm *ProcessManager) AddProcess(processQuery *process_manager_queries.ProcessCreateQuery, reply *error) error {
	if pm.serviceIsIgnored(processQuery.Name, processQuery.OzoneWorkingDir) {
		return nil
	}
	cmdString := processQuery.Cmd

	logFile, tempDir, err := pm.setUpLogging(processQuery.OzoneWorkingDir, processQuery.Name)
	if err != nil {
		reply = &err
		return err
	}
	cmdString = substituteOutput(cmdString, tempDir)
	log.Println("cmd is:")
	log.Println(cmdString)

	cmdFields, argFields := CommandFromFields(cmdString)
	cmd := exec.Command(cmdFields[0], argFields...)

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
			processQuery.Env,
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

	return nil
}

func (pm *ProcessManager) setUpLogging(ozoneWorkingDir string, serviceName string) (*os.File, string, error) {
	tempDir := pm.createTempDirIfNotExists(ozoneWorkingDir)
	logFileString := fmt.Sprintf("%s/%s-logs", tempDir, serviceName)
	fmt.Printf("Logs at:\n")
	fmt.Printf("%s \n", logFileString)
	file, err := CreateLogFileTempDirIfNotExists(logFileString)
	if err != nil {
		return nil, "", err
	}
	return file, tempDir, nil
}

func CreateLogFileTempDirIfNotExists(logFileString string) (*os.File, error) {
	_, err := os.Stat(logFileString)

	var logFile *os.File
	if os.IsNotExist(err) {
		logFile, err = os.Create(logFileString)
		if err != nil {
			return nil, err
		}
	}
	logFile, err = os.OpenFile(logFileString, os.O_WRONLY, os.ModeAppend)
	if err != nil {
		return nil, err
	}
	fmt.Println("Created log file")
	return logFile, err
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
	env *config_variable.VariableMap,
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

	portStringVar, ok := env.GetVariable("PORT")
	if !ok {
		return errors.New("PORT needed")
	}

	port, err := strconv.ParseInt(portStringVar.String(), 10, 64)
	if err != nil {
		return err
	}
	process := &OzoneProcess{
		Cmd:       cmd,
		Name:      name,
		StartTime: time.Now().Unix(),
		Port:      port,
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

func (pm *ProcessManager) Debug(processQuery *process_manager_queries.ProcessCreateQuery, reply *int) error {
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
