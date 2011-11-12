package main

import (
	"fmt"
	"mongrel2"
	"os"
	"github.com/alecthomas/gozmq"
)

const (
	//easier to see with fewer choices
	MAX_WORKERS = 3
)

//simple demo of the raw mogrel2 level with a number of goroutines reading 
//from the same input channel (all write to the same output channel)
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

	//we spawn some number of worker go routines to read from the channel
	//round-robin delivery as you will see from the print messages
	for i := 0; i < MAX_WORKERS-1; i++ {
		go handle(in, i, out) //this a closure, so vars are ok
	}
	//we spawn the last one in the "main" goroutine so we don't exit
	handle(in,MAX_WORKERS-1,out)
}

// simple worker routine: read a request and respond to it.  the interesting
// part is the order of the printouts to screen in the browser.  the id
// of the handler should be changing in round-robin format
func handle(in chan *mongrel2.Request, id int, out chan *mongrel2.Response) {

	for {
		req,ok := <-in
		if req==nil {
			if ok {	
				panic("nil message received! who is sending that?")
			} else {
				fmt.Printf("channel has been closed, exiting (%d)\n",id)
				return
			}
		}
		
		resp := new(mongrel2.Response)
		resp.ServerId = req.ServerId
		resp.ClientId = []int{req.ClientId}
		resp.Body = fmt.Sprintf("thanks for talking to handler %2d, client %2d", id, req.ClientId)
		resp.Header = map[string]string{"Content-Type": "text/plain"}
		out <- resp
	}
}
