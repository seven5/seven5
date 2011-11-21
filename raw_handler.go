package seven5

import (
	"bytes"
	"fmt"
	"mongrel2"
	"runtime"
	"strings"
	"errors"
	"github.com/alecthomas/gozmq"
	"os"
)

const (
	//When doing route testing, this is the result value the framework will give in
	//response to a GET request on a raw handler.
	ROUTE_TEST_RESPONSE_CODE = 209
	//When doing route testing, requests should be marked with this header and the
	//header's value should be the name of the handler expected to get it.  When
	//generating a successful response to a test, seven5 also sets this header in the 
	//response.
	ROUTE_TEST_HEADER = "X-Seven5-Route-Test"
)
//RawHandler low-level interface to the raw mongrel2 communication that allows messages to
//be processed one at a time, at the mongrel2 level.  The parameter/return here are the mongrel2
//request and response.  RawHandler is a building block for other abstractions in the seven5
//toolkit.  
//Most developers shoud never need this interface.
type RawHandler interface {
	Name() string
	SetRawSockets(in gozmq.Socket, out gozmq.Socket)
	ProcessRequest(request *mongrel2.Request) (*mongrel2.Response)
}

//Called to process requests for a RawHandler in a never ending loop.  Should be called as a 
//go func.
func runloop(h RawHandler, config *ProjectConfig, in chan *mongrel2.Request, out chan *mongrel2.Response) {
	for {
		//block until we get a message from the server
		req := <-in

		if req == nil {
			config.Logger.Printf("[Raw Handler '%s' (%s)]: close of mongrel2 connection in raw handler!", h.Name(), config.Name)
			return
		} else {
			config.Logger.Printf("[Raw Handler '%s' (%s)]: serving %s", h.Name(), config.Name, req.Path)
		}
		
		//note: mongrel converts this to lower case!
		testHeader:=req.Header[strings.ToLower(ROUTE_TEST_HEADER)]
		if h.Name()==testHeader {
			testResp:=new (mongrel2.Response)
			config.Logger.Printf("[ROUTE TEST] '%s' (%s) : %s\n", h.Name(), config.Name, req.Path)
			testResp.ClientId=[]int{req.ClientId}
			testResp.ServerId=req.ServerId
			testResp.StatusCode=ROUTE_TEST_RESPONSE_CODE
			testResp.Header=map[string]string{ROUTE_TEST_HEADER:h.Name()}
			testResp.StatusMsg = "Thanks for testing with seven5"
			out <- testResp
			continue
		}

		resp := protectedProcessRequest(config, req, h)
		out <- resp
	}
}

//RunRaw starts a RawHandler (mongrel2 level handler) running based on the information in
//project description.
func RunRaw(h RawHandler,ctx gozmq.Context, config *ProjectConfig) (*mongrel2.Handler,error) {

	in := make(chan *mongrel2.Request)
	out := make(chan *mongrel2.Response)

	var addr *mongrel2.HandlerAddr
	
	for _,candidate := range config.Addr {
		if candidate.Name==h.Name() {
			addr=candidate
		}
	}
	
	if addr==nil {
		return nil,errors.New(fmt.Sprintf("Unable to find address assigned to handler named '%s'", h.Name()))
	}
	
	config.Logger.Printf("[Raw handler '%s' (%s)] : connecting to %s and %s",h.Name(), config.Name, addr.PullSpec, addr.PubSpec)
	
	//connect to service
	mongrel2Part, err := mongrel2.NewHandler(addr, in, out, ctx)

	if err != nil {
		return nil, errors.New(fmt.Sprintf("Raw handler %s: error initializing mongrel connection:%s\n", h.Name(), err))
	}
	
	//only using one routing for now, synchronous
	go runloop(h,config,in,out)
	
	return mongrel2Part, nil
}

//protectedProcessRequest is present to allow us to trap panic's that occur inside the web 
//application.  The web application should not really ever do this, it should generate 500
//pages instead but a nil pointer dereference or similar is possible.
func protectedProcessRequest(config *ProjectConfig, req *mongrel2.Request, h RawHandler) (resp *mongrel2.Response) {
	defer func() {
		if x := recover(); x != nil {
			config.Logger.Printf("[Raw Handler '%s' (%s)]: PANIC! sent error page for %s: %v\n", h.Name(), config.Name, req.Path, x)
			resp = new (mongrel2.Response)
			resp.StatusCode = 500
			resp.StatusMsg = "Internal Server Error"
			resp.Body = fmt.Sprintf("Panic: %v\n",x)
		}
	}()
	resp = h.ProcessRequest(req)
	config.Logger.Printf("[Raw Handler '%s' (%s)]: responded to %s with %d bytes of content\n", h.Name(), config.Name, req.Path, len(resp.Body))
	return
}

//Generate500Page returns an error page as a mongrel2.Response.  This includes a call stack of the point
//where the caller called this function.
func Generate500Page(err string, request *mongrel2.Request) *mongrel2.Response {
	fiveHundred := new(mongrel2.Response)
	
	fiveHundred.ServerId = request.ServerId
	fiveHundred.ClientId = []int{request.ClientId}
	
	fiveHundred.StatusCode = 500
	fiveHundred.StatusMsg = "Internal Server Error"
	fiveHundred.Body = generateStackTrace(fmt.Sprintf("%v", err))
	return fiveHundred
}

func generateStackTrace(err string) string {
	buffer := new(bytes.Buffer)
	buffer.WriteString(err)
	buffer.WriteString("\n----Stacktrace----\n")
	for i := 2; ; i++ {
		_, file, line, ok := runtime.Caller(i)
		if ok {
			f := strings.Split(file, "/")
			s := fmt.Sprintf("%s: line %d\n", f[len(f)-1], line)
			buffer.WriteString(s)
		} else {
			break
		}
	}
	return buffer.String()
}

//StartUp is what most web apps will want to use as an entry point. The return parameter 
//is the zmq context for this application and this should be closed on shutdown 
//(usually using defer). This functions runs all the RawHandler objects provided in a go
//routine using RunRaw().  If something went wrong this returns nil and most web
//apps will just want to exit since the error has already been printed.  If you are
//calling this from test code, you will want to set the second parameter to the proposed
//project directory; otherwise pass "" and it will be retreived from the command line args.
func StartUp(handlers []RawHandler, proposedDir string) gozmq.Context{
	var conf *ProjectConfig
	var ctx gozmq.Context
	
	if proposedDir!="" {
		conf, ctx=BootstrapFromDir(proposedDir)
	} else {
		conf, ctx=Bootstrap()
	}
	if conf==nil {
		return nil
	}
	
	for _,h := range handlers {
		mongrel2Level, err:= RunRaw(h,ctx,conf)

		if err != nil {
			fmt.Fprintf(os.Stderr, "error starting the mongrel connection for %s:%s\n",h.Name(), err)
			return nil
		}
		
		h.SetRawSockets(mongrel2Level.InSocket, mongrel2Level.OutSocket)
	}
	
	return ctx
}