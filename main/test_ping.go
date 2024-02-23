package main 
import (
	"net/rpc"
	"fmt"
	"duc.anh/core_server/server_config"
)

func main() {
	address := fmt.Sprintf("localhost:%d", 59199)
	// config := new(server_config.ServerConfig)
	conn, err := rpc.Dial("tcp", address)
	fmt.Println("Connecting to ", address,err)
	if err != nil {
		fmt.Println("Error: ", err)
		return
	}
	var reply *bool
	input := new(server_config.ServerConfig)
	fmt.Println("Registering backup node ",input)
	err = conn.Call("ServerConfig.RegisterBackUpNode", 125, &reply)
	if err != nil {
		fmt.Println("Error: ", err)
		return
	}
	fmt.Println("Connected to ", address, "  ", *reply)

}