package process_manager_client

import (
    "encoding/gob"
    "fmt"
    "log"
    "net/rpc"
    "os"
    process_manager "ozone-daemon-lib/process-manager"
)

func Halt() error {
    ozoneWorkingDir, err := os.Getwd()
    if err != nil {
        log.Println(err)
    }

    query := process_manager.DirQuery{
        ozoneWorkingDir,
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

    if err = call("TempDirRequest", &request, &reply); err != nil {
        return err, ""
    }

    return nil, reply.Body
}

func Status() (error, string) {
    ozoneWorkingDir, err := os.Getwd()
    if err != nil {
        log.Println(err)
    }

    query := process_manager.DirQuery{
        OzoneWorkingDir: ozoneWorkingDir,
    }
    reply := process_manager.StringReply{
        Body: "",
    }
    if err = call("Status", &query, &reply); err != nil {
        return err, ""
    }
    return nil, reply.Body
}

func call(name string, query interface{}, reply interface{}) error {
    gob.Register(&process_manager.DirQuery{})
    gob.Register(&process_manager.StringReply{})
    client, err := rpc.DialHTTP("tcp", ":8000")
    if err != nil {
        log.Fatal("dialing:", err)
    }

    rpcName := fmt.Sprintf("ProcessManager.%s", name)
    err = client.Call(rpcName, query, reply)
    if err != nil {
        log.Fatal("transport error:", err) // TODO daemon not running error?
        return err
    }
    return nil
}