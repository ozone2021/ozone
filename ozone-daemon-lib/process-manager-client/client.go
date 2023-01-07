package process_manager_client

import (
	"encoding/gob"
	"fmt"
	process_manager "github.com/ozone2021/ozone/ozone-daemon-lib/process-manager"
	"log"
	"net/rpc"
	"os"
)

func Halt(service string) error {
	ozoneWorkingDir, err := os.Getwd()
	if err != nil {
		log.Println(err)
	}

	query := process_manager.HaltQuery{
		ozoneWorkingDir,
		service,
	}
	var reply error

	return call("Halt", query, &reply)
}

func FetchTempDir() (error, string) {
	ozoneWorkingDir, err := os.Getwd()
	if err != nil {
		log.Println(err)
	}

	request := process_manager.DirQuery{
		OzoneWorkingDir: ozoneWorkingDir,
	}

	reply := process_manager.StringReply{
		Body: "",
	}
	//
	//client, err := rpc.DialHTTP("tcp", ":8000")
	//if err != nil {
	//	log.Fatal("dialing:", err)
	//}
	//defer client.Close()
	//err = client.Call("ProcessManager.TempDirRequest", request, &response)

	if err := call("TempDirRequest", &request, &reply); err != nil {
		return err, ""
	}

	return nil, reply.Body
}

func AddProcess(query *process_manager.ProcessCreateQuery) error {
	var errReply error

	if err := call("AddProcess", query, &errReply); err != nil {
		return err
	}

	return errReply
}

func SetContext(dir string, context string) error {
	query := process_manager.ContextSetQuery{
		OzoneWorkingDir: dir,
		Context:         context,
	}
	var reply process_manager.StringReply

	if err := call("SetContext", &query, &reply); err != nil {
		return err
	}

	return nil
}

func FetchContext(dir string) (string, error) {
	query := &process_manager.DirQuery{
		OzoneWorkingDir: dir,
	}
	var reply process_manager.StringReply

	err := call("FetchContext", &query, &reply)
	if err != nil {
		return "", err
	}

	return reply.Body, nil
}

func Ignore(ozoneWorkingDir, serviceName string) error {
	query := process_manager.IgnoreQuery{
		OzoneWorkingDir: ozoneWorkingDir,
		Service:         serviceName,
	}

	var errReply error
	err := call("Ignore", &query, &errReply)

	if err != nil {
		return err
	}
	if errReply != nil {
		return errReply
	}

	return nil
}

func CacheUpdate(ozoneWorkingDir string, runnableName string, ozoneFileAndDirHash string) bool {
	query := process_manager.CacheQuery{
		OzoneWorkingDir:     ozoneWorkingDir,
		RunnableName:        runnableName,
		OzoneFileAndDirHash: ozoneFileAndDirHash,
	}
	reply := process_manager.BoolReply{}

	if err := call("UpdateCache", &query, &reply); err != nil {
		log.Println(err)
	}
	return reply.Body
}

func CacheCheck(ozoneWorkingDir string, runnable string) string {
	query := process_manager.CacheQuery{
		OzoneWorkingDir: ozoneWorkingDir,
		RunnableName:    runnable,
	}
	reply := process_manager.StringReply{}

	call("CheckCache", &query, &reply)
	return reply.Body
}

func Status(ozoneWorkingDir string) (error, string) {
	query := process_manager.DirQuery{
		OzoneWorkingDir: ozoneWorkingDir,
	}
	reply := process_manager.StringReply{
		Body: "",
	}
	if err := call("Status", &query, &reply); err != nil {
		return err, ""
	}
	return nil, reply.Body
}

func call(name string, query interface{}, reply interface{}) error {
	gob.Register(&process_manager.DirQuery{})
	gob.Register(&process_manager.StringReply{})
	gob.Register(&process_manager.BoolReply{})
	client, err := rpc.DialHTTP("tcp", ":8000")
	if err != nil {
		log.Fatal("dialing:", err)
	}

	rpcName := fmt.Sprintf("ProcessManager.%s", name)
	defer client.Close()
	err = client.Call(rpcName, query, reply)
	if err != nil {
		log.Fatal("transport error:", err) // TODO daemon not running error?
		return err
	}
	return nil
}
