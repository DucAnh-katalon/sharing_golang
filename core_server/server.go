package core_server
import (
	
	"time"
	"fmt"
	"net"
	"net/rpc"
	"duc.anh/core_server/server_config"
	"duc.anh/core_server/rpc_fastdb"
	"github.com/phayes/freeport"
	
)

func pingToAddress(address string) bool {
	conn, err := rpc.Dial("tcp", address)
	if err != nil {
		fmt.Println("Error: ", err)
		return false
	}
	defer conn.Close()
	return true
}

func isPrimaryServer(current_port int) (bool, bool) {
	cluster_config_address := fmt.Sprintf("localhost:%d", server_config.CONFIG_PORT)
	caller, err := rpc.Dial("tcp", cluster_config_address)
	current_port_is_primary := false
	sever_config_started := false
	if err != nil {
		cfg,err := server_config.ReadConfig()
		if err != nil {
			return true, false
		}
		fmt.Println("[Server.isPrimaryServer]:: Error when connecting to cluster config: ", err)
		for idx, port := range cfg.BackUpNode {
			if port == current_port {
				cfg.PrimaryNode = current_port
				cfg.BackUpNode = cfg.BackUpNode[idx+1:]
				fmt.Println("[Server.isPrimaryServer]:: This server is primary",cfg)
				cfg.WriteConfig()

				return true, false
			}
			address := fmt.Sprintf("localhost:%d", port)
			if pingToAddress(address){
				current_port_is_primary = false
				sever_config_started = true
				break
			}
		}
		
	}else {
		var reply *bool
		caller.Call("ServerConfig.IsPrimary", current_port, &reply)
		caller.Close()
		current_port_is_primary = *reply
		fmt.Println("[Server.isPrimaryServer]:: Current port is primary: ", current_port_is_primary)
		sever_config_started = true
	}
	return current_port_is_primary, sever_config_started
}

func CurrentPortListen(current_port int){
	cfg := new(server_config.ServerConfig)
	cfg.NodePort = current_port
	current_node_address := fmt.Sprintf(":%d", cfg.NodePort)
	new_server := rpc.NewServer()
	register_err := new_server.Register(cfg)
	store, _ := rpc_fastdb.Open("replicate_db:", 100)
	register_err = new_server.RegisterName("RPCDB",store)
	
	if register_err != nil {
		fmt.Println("[Server.CurrentPortListen]:: Registering server config error: ", register_err)
	}
	listener, err := net.Listen("tcp", current_node_address)
	if err != nil {
		fmt.Println("[Server.CurrentPortListen]:: Creating listener error: ", err)
		return
	}	
	// defer listener.Close()
	fmt.Println("[Server.CurrentPortListen]:: CurrentPortListen on port: ", cfg.NodePort)
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("[Server.CurrentPortListen]:: Listener accpet call errror ", err)
			continue
		}
		fmt.Println("[Server.CurrentPortListen]:: New connection accepted",conn)

		go new_server.ServeConn(conn)
		fmt.Println("[Server.CurrentPortListen]:: New connection served",cfg, store.Transaction)
	}
}

func ServiceListen(){
	services_port := server_config.SERVICE_PORT
	fmt.Println("[Server.ServiceListen]:: Service_listen on port: ", services_port)
	new_server := rpc.NewServer()
	store, _ := rpc_fastdb.Open("replicate_db:", 100) 
	new_server.RegisterName("RPCDB", store)

	services_listener, _ := net.Listen("tcp", fmt.Sprintf("localhost:%d", services_port))
	for {
		conn, err := services_listener.Accept()
		if err != nil {
			fmt.Println("[Server.ServiceListen]:: Listener accpet call errror ", err)
			continue
		}
		new_server.ServeConn(conn)
		transaction := store.Transaction[len(store.Transaction)-1]
		func_name := transaction.FuncName
		args := transaction.Args
		fmt.Println("[Server.ServiceListen]:: New transaction: ", func_name, "  ", args)
		cfg, err := server_config.ReadConfig()
		if err != nil {
			fmt.Println("[Server.ServiceListen]:: Error when reading config")
			continue
		}
		for _, port := range cfg.BackUpNode {
			fmt.Println("[Server.ServiceListen]:: Syncing DB to backup node: ", port)
			address := fmt.Sprintf("localhost:%d", port)
			backup_node_caller ,err := rpc.Dial("tcp", address)
			if err != nil {
				fmt.Println("[Server.ServiceListen]:: Error when connecting to backup node: ", err)
				continue
				backup_node_caller.Close()
			}
			var reply *rpc_fastdb.Response
			
			err = backup_node_caller.Call("RPCDB."+func_name, args, &reply)
			if err != nil {
				fmt.Println("[Server.ServiceListen]:: Error when syncing to backup node: ", err)
			}
			backup_node_caller.Close()
		}
		}
}


func SyncConfig(cfg *server_config.ServerConfig){
	fmt.Println("[SyncConfig]:: Syncing config")	
	fmt.Println("[Server.ConfigPortListen.SyncConfig]:: Syncing config to backup nodes",cfg)
	var err error
	var reply *bool
	var alive_backup_node []int
	//init alive backup node list is the same with backup node list
	alive_backup_node = cfg.BackUpNode
	for idx, port := range cfg.BackUpNode {
		backup_config_address := fmt.Sprintf("localhost:%d", port)
		fmt.Println("[Server.ConfigPortListen.SyncConfig]:: Syncing config to backup node: ", backup_config_address)
		
		backup_config_listener ,err_backup := rpc.Dial("tcp", backup_config_address)
		if err_backup != nil {
			fmt.Println("Error: ", err_backup, " when connecting to ", backup_config_address)
			// pop this port from backup node list
			copy(alive_backup_node[idx:], alive_backup_node[idx+1:]) // Shift a[i+1:] left one index.
			alive_backup_node[len(alive_backup_node)-1] = 0  
			alive_backup_node = alive_backup_node[:len(alive_backup_node)-1]  
			fmt.Println("Alive backup node: ", alive_backup_node)
		}
		err = backup_config_listener.Call("ServerConfig.UpdateAndWriteConfig", *cfg, &reply)
		if err != nil {
			fmt.Println("Error when sync to : ",port, "  ", err)
			// pop this port from backup node list
		}
		backup_config_listener.Close()
	}
	// check if there is any down backup node
	if len(alive_backup_node) != len(cfg.BackUpNode) {
		fmt.Println("[Server.ConfigPortListen.SyncConfig]:: Some backup node is down")
		cfg.BackUpNode = alive_backup_node
		go SyncConfig(cfg)
	}
	fmt.Println("[Server.ConfigPortListen.SyncConfig]:: Syncing config done")
}

func RegisterBackUpNode(current_port int){
	fmt.Println("[Server.RegisterBackUpNode]:: Registering backup node")
	cluster_config_address := fmt.Sprintf("localhost:%d", server_config.CONFIG_PORT)
	cluster_config_listener, err := rpc.Dial("tcp", cluster_config_address)
	if err != nil {
		fmt.Println("[Server.RegisterBackUpNode] connect to cluster config errir", err)
		return 
	}
	
	var reply *bool
	err = cluster_config_listener.Call("ServerConfig.RegisterBackUpNode",current_port, &reply)	
	if err != nil {
		fmt.Println("Error: ", err)
		return 
	}
	fmt.Println("[Server.RegisterBackUpNode]:: Registering backup node done")
	cluster_config_listener.Close()

	// service_node_address := fmt.Sprintf("localhost:%d", server_config.SERVICE_PORT)
	// service_node_listener, err := rpc.Dial("tcp", service_node_address)
	// if err != nil {
	// 	fmt.Println("[Server.RegisterBackUpNode] connect to service node error", err)
	// }
	// var reply_service *bool
	// fmt.Println("syncdata")
	// arg,_ :=  rpc_fastdb.CreateArgs("GetInfo", "bucket", 0, nil)
	// err = service_node_listener.Call("RPCDB.CopyDB", store, &reply_service)
	// if err != nil {
	// 	fmt.Println("sync Data Error: ", err)
	// }
	// service_node_listener.Close()
	return
	
}	

func ConfigPortListen(current_port int){
	cfg, err := server_config.ReadConfig()
	need_sync := false
	if err != nil {
		fmt.Println("[Server.ConfigPortListen]:: Error when reading config.\nCreated one",)
		cfg = new(server_config.ServerConfig)
		cfg.NodePort = current_port
		cfg.WriteConfig()
	} else {
		fmt.Println("[Server.ConfigPortListen]:: Config: ", cfg)
		need_sync = true
	}
	cfg.PrimaryNode = cfg.NodePort
	cluster_config_address := fmt.Sprintf("localhost:%d", server_config.CONFIG_PORT)
	cfg_watcher := new(server_config.ServerConfig)
	*cfg_watcher = *cfg
	new_server := rpc.NewServer()
	register_err := new_server.Register(cfg_watcher)
	if register_err != nil {
		fmt.Println("[Server.ConfigPortListen]:: Registering server config error: ", register_err)
	}
	cluster_config_listener, _ := net.Listen("tcp", cluster_config_address)
	defer cluster_config_listener.Close()
	fmt.Println("[Server.ConfigPortListen]:: ConfigPortListen on port: ", server_config.CONFIG_PORT)
	if need_sync {
		SyncConfig(cfg_watcher)
	}
	for {
		conn, err := cluster_config_listener.Accept()
		if err != nil {
			fmt.Println("[Server.ConfigPortListen]:: Listener accpet call errror ", err)
			continue
		}
		new_server.ServeConn(conn)
		if cfg_watcher.NeedSync{
			fmt.Println("[Server.ConfigPortListen]:: Syncing config ",cfg_watcher.BackUpNode)
			go SyncConfig(cfg_watcher)
			cfg_watcher.NeedSync = false
		}
	}
}


func DoPrimaryNode(current_port int){	
	// Listen on service port
	go ServiceListen()
	// Listen on config port
	go ConfigPortListen(current_port)	
}


func StartServer() {
	current_port, _ := freeport.GetFreePort()
	fmt.Println("[StartServer]:: Starting server on port: ", current_port)
	go CurrentPortListen(current_port)
	Register_back_up_node := false
	for {
		// check if this server is primary or not
		
		fmt.Println("[StartServer]:: Checking if this server is primary or not")
		is_primary, is_primary_node_start := isPrimaryServer(current_port)
		if is_primary {
			if !is_primary_node_start {
				go DoPrimaryNode(current_port)
			}
			fmt.Println("[StartServer]:: This server is primary")
		} else {
			fmt.Println("[StartServer]:: This server is backup")
			if !Register_back_up_node {
				RegisterBackUpNode(current_port)
				Register_back_up_node = true
			}
		}
			
		time.Sleep(20 * time.Second)
	}
}
