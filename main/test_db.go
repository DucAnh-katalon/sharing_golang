package main

import (
	"duc.anh/core_server/rpc_fastdb"
)

func main(){
	store, _ := rpc_fastdb.Open("replicate_db:", 100)

}