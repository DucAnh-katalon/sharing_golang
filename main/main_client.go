package main

import (
	"fmt"
	"net/rpc"
	"log"

	"duc.anh/core_server/rpc_fastdb"
)

func main() {
	fmt.Println("Starting client...")
	serverAddress := "localhost"
	client, err := rpc.Dial("tcp", serverAddress + ":1234")
	if err != nil {
		log.Fatal("dialing:", err)
	}
	fmt.Println("client connected to server at " + serverAddress + ":1234")
	
	var reply rpc_fastdb.Response
	key := 1
	value := []byte("value 1")
	args, err := rpc_fastdb.CreateArgs("Get", "bucket", key, value)
	client.Call("RPCDB.Get", args, &reply)
	if err != nil {
		log.Fatal("RPCDB error:", err)
	}
	client.Close()
	body,_ := rpc_fastdb.DecodeResponse(reply.Body)
	
	fmt.Println("status get: ", reply.Success)
	fmt.Println("RPCDB.Get response:", string (body["Data"]))

	sever_backup := "localhost:59549"
	args, err = rpc_fastdb.CreateArgs("Get", "bucket", key, value)
	client, err = rpc.Dial("tcp", sever_backup)
	if err != nil {
		log.Fatal("dialing:", err)
	}
	fmt.Println("client connected to server at " + sever_backup)
	client.Call("RPCDB.Get", args, &reply)
	if err != nil {
		log.Fatal("RPCDB error:", err)
	}
	client.Close()
	body,_ = rpc_fastdb.DecodeResponse(reply.Body)
	fmt.Println("status get: ", reply.Success)
	fmt.Println("RPCDB.Set response:", string (body["Data"]))
}