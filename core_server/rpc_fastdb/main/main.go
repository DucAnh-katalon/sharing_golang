package main 

import (
	"fmt"
	"detrue/core_rpc"	
)

func main() {
	store, err := core_rpc.Open(":memory:", 100)
	if err != nil {
		fmt.Println(err)
		return
	}
	args := &core_rpc.SetArgs{"test", 2, []byte("test1")}
	var reply bool
	store.Set(args,&reply)
	fmt.Println(reply)

	var reply2 []byte
	store.Get(args,&reply2)
	fmt.Println(reply2)

}