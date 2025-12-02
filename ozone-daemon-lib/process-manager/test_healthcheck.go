package main

import (
	"fmt"
	"log"
	"net/rpc"
)

func main() {
	client, err := rpc.DialHTTP("tcp", "localhost:8000")
	if err != nil {
		log.Fatalf("Error connecting to daemon: %v", err)
	}

	var reply struct{ Body string }
	err = client.Call("ProcessManager.HealthCheck", struct{}{}, &reply)
	if err != nil {
		log.Fatalf("RPC call failed: %v", err)
	}

	fmt.Println("HealthCheck reply:", reply.Body)
}
