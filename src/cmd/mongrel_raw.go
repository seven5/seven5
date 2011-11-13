package main

import (
	"fmt"
	"github.com/alecthomas/gozmq"
	"mongrel2"
	"os"
	"time"
)

//Simple demo program that processes one request and returns one response to
//the server and client that sent the request.
func main() {

	// do a version check
	x, y, z := gozmq.Version()
	if x != 2 && y != 1 {
		fmt.Printf("version of zmq is %d.%d.%d and this code was tested primarily on 2.1.10\n", x, y, z)
	}
	// we need to pick a name and register it
	addr, err := mongrel2.GetHandlerAddress("some_name") //we dont really care

	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to get an address for our handler:%s\n", err)
		return
	}

	//allocate channels so we can talk to the mongrel2 system with go
	// abstractions
	in := make(chan *mongrel2.Request)
	out := make(chan *mongrel2.Response)

	// this allocates the "raw" abstraction for talking to a mongrel server
	// mongrel doc refers to this as a "handler"
	handler, err := mongrel2.NewHandler(addr, in, out)

	if err != nil {
		fmt.Fprintf(os.Stderr, "error initializing mongrel connection:%s\n", err)
		return
	}
		
	// don't forget to clean up various resources when done
	defer handler.Shutdown()

	//block until we get a message from the server
	req := <- in 
			
	// there are many interesting fields in req, but we just print out a couple
	fmt.Printf("server %s sent %s from client %d\n", req.ServerId, req.Path, req.ClientId)

	//create a response to go back to the client
	response := new(mongrel2.Response)
	response.ServerId = req.ServerId
	response.ClientId = []int{req.ClientId}
	response.Body = fmt.Sprintf("<pre>hello there, %s with client %d!</pre>", req.ServerId, req.ClientId)

	//send it via the other channel
	fmt.Printf("Responding to server with %d bytes of content\n",len(response.Body))
	out <- response
	
	//this is what we have to do to make sure the sent message gets delivered
	//before we shut down.  we tried to use ZMQ_LINGER(-1) at init time but
	//this is generating an error right now.
	time.Sleep(1000000000)
}
