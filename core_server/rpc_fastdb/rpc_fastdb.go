package rpc_fastdb

import (
	"fmt"
	"errors"
	"github.com/marcelloh/fastdb"
	"encoding/json"
)

type Transaction struct {
	// Transaction is a struct that is used to store transaction
	FuncName string
	Args *RequestArgs
}

type RPCDB struct {
	// DB is a fastdb.DB that is exposed to RPC
	DB *fastdb.DB
	Transaction []Transaction
}

type RequestArgs struct {
	// Args is a struct that is used to pass arguments to RPC
	Bucket string
	Key int
	Value []byte
}

type getResponse struct {
	// Response is a struct that is used to pass response to RPC for get method
	Data []byte 
}

type setResponse struct {
	// Response is a struct that is used to pass response to RPC for set method
	Data []byte
}

type Response struct {
	// Response is a struct that is used to pass response to RPC
	Success bool
	Body []byte
}

func CreateArgs(nameFunc string, bucket string, key int, value []byte) (*RequestArgs, error){
	// CreateArgs is a function that is used to create RequestArgs
	switch nameFunc {
	case "Set":
		if value == nil {
			return nil, errors.New("value is nil")
		}
		return &RequestArgs{bucket, key, value}, nil
	case "Get":
		return &RequestArgs{bucket, key, nil}, nil
	case "GetInfo":
		return &RequestArgs{bucket, key, nil}, nil
	default:
	
		return nil, errors.New("nameFunc not found must in [Set, Get]")
	}
	
}

func Open(path string, syncIime int) (*RPCDB, error){
	// Open is a function that is used to open fastdb.DB
	store, err := fastdb.Open(path, syncIime)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	return &RPCDB{DB: store}, nil
}



func (r *RPCDB) CopyDB(new_DB *RPCDB, reply *bool) error {
	// CopyDB is a function that is used to copy fastdb.DB
	fmt.Println("[RPCDB.CopyDB]:: Copying DB")
	r.Transaction = new_DB.Transaction
	for _, trans := range new_DB.Transaction {
		var reply  *Response
		if trans.FuncName == "Set"{
			r.Set(trans.Args, reply)
		}
	}
	*reply = true
	return nil
}

func (r *RPCDB) Set(args *RequestArgs, reply *Response) error {
	// Set is a function that is used to set value to fastdb.DB
	err := r.DB.Set(args.Bucket, args.Key, args.Value)
	if err != nil {
		*reply = Response{false,nil}
		return err
	}
	
	response := &setResponse{args.Value}
	body, err := json.Marshal(response)
	if err != nil {
		fmt.Println("encoding set response", err)
		return err
	}

	full_response := &Response{true, body}
	*reply = *full_response
	r.Transaction = append(r.Transaction, Transaction{"Set", args})
	fmt.Println("get new transaction ",r.Transaction)
	return nil
}

func (r *RPCDB) Get(args *RequestArgs, reply *Response) error {
	// Get is a function that is used to get value from fastdb.DB
	value, found := r.DB.Get(args.Bucket, args.Key)
	if !found {
		*reply = Response{false,nil}
		return errors.New("key not found")
	}
	response := &getResponse{value}
	body, err := json.Marshal(response)
	if err != nil {
		fmt.Println("encoding get response", err)
		return err
	}
	full_response := &Response{true, body}
	*reply = *full_response
	r.Transaction = append(r.Transaction, Transaction{"Get", args})
	return nil
}



func DecodeResponse(data []byte) (map[string] []byte, error) {
	var m map[string] []byte
	if data == nil {
		return nil, nil
	}
	err := json.Unmarshal(data, &m)
	if err != nil {
		fmt.Println("deccode error:", err)
		return nil, err
	}
	return m, nil
}