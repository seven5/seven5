package main

import (
	"fmt"
	"github.com/alecthomas/gozmq"
	"os"
	"seven5"
)

func main() {

	// do a version check
	x, y, z := gozmq.Version()
	if x!=2 && y!=1 {
		fmt.Printf("version of zmq is %d.%d.%d and this code was tested primarily on 2.1.10\n", x, y, z)
	}
	// we need to register our version with the ZMQ infrastructure
	me, err := seven5.Type4UUID()
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to get new UUID:%s\n", err)
	}

	// this is the "raw" abstraction for talking to a mongrel server
	mongrel2 := seven5.NewMongrel2("tcp://127.0.0.1:9997", "tcp://127.0.0.1:9996", me)
	// don't forget to clean up various resources when done
	defer mongrel2.Shutdown()

	//read a single request
	req, err := mongrel2.ReadMessage()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading message:%s\n", err)
		return
	}
	// there are many interesting fields in req, but we just print out a couple
	fmt.Printf("server %s sent %s from client %d\n", req.ServerId, req.Path, req.ClientId)

	//create a response to go back to the client
	response := new(seven5.Response)
	response.UUID = req.ServerId
	response.ClientId = []int{req.ClientId}
	response.Body = fmt.Sprintf("<pre>hello there, %s with client %d!</pre>", req.ServerId, req.ClientId)

	//send it via the mongrel server
	err = mongrel2.WriteMessage(response)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading message:%s\n", err)
		return
	}
}
