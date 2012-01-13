package seven5

import (
	"bytes"
	"fmt"
	"mongrel2"
	"runtime"
	"strings"
)

//HttpRunner is an interface that is used by the infrastructure to indicate that the object can
//"pump" http through it. 
type HttpRunner interface {
	mongrel2.HttpHandler
	//RunHttp is called by the infrastructure to start the flow of the HTTP requests and responses.
	//The second parameter is where these HTTP requests/responses are processed.  The "target" does
	//not know about or deal with the particulars of the way the requests and response are
	//gathered or squirted back to the client (it doesn't deal with any communications).
	RunHttp(config *projectConfig, target Httpified)
}

//Httpified is an interface indicating that the object in question can process mongrel2
//messages as expected for a mongrel2 handler.  There may be implementations of Httpifid later 
//that are not tied to mongrel2 (and this probably means ProcessRequest will change).
type Httpified interface {
	Named
	//ProcessRequest is the critical method.  All the HttpRequests are pushed through this method
	//and each one should generate a response object.  Implementors of this interface are handling
	//HTTP "raw". 
	ProcessRequest(request *mongrel2.HttpRequest) *mongrel2.HttpResponse
}

//HttpRunnerDefault is a default implementation of the HttpRunner interface for a mongrel2 based
//service.  It uses two Go channels: one for reading requests, the other for writing responses. It
//does not know how to deal with sockets, just channels.  The socket level is handled by the mongrel2
//package.
type HttpRunnerDefault struct {
	*mongrel2.HttpHandlerDefault
	In  chan *mongrel2.HttpRequest
	Out chan *mongrel2.HttpResponse
}

//RunHttp launches the processing of the HTTP protocol via goroutines.  This call will 
//never return.  It will repeatedly call ProcessRequest() as messages arrive and need
//to be sent to the mongrel2 server.  It is careful to protect itself from code that
//might call panic() even though this is _not_ advised for implementors; it is preferred
//to have implementations that detect a problem use the HTTP error code set.
func (self *HttpRunnerDefault) RunHttp(config *projectConfig, target Httpified) {

	i := make(chan *mongrel2.HttpRequest)
	o := make(chan *mongrel2.HttpResponse)
	self.In = i
	self.Out = o

	go self.ReadLoop(self.In)
	go self.WriteLoop(self.Out)

	for {
		//block until we get a message from the server
		req := <-self.In

		if req == nil {
			config.Logger.Printf("[%s]: close of mongrel2 connection in raw handler!", target.Name())
			return
		} else {
			config.Logger.Printf("[%s]: serving %s", target.Name(), req.Path)
		}

		/*for k,v:=range req.Header {
			fmt.Fprintf(os.Stderr,"--->>> header: '%s'='%s'\n",k,v)
		}*/

		//note: mongrel converts this to lower case!
		testHeader := req.Header[strings.ToLower(route_test_header)]
		if target.Name() == testHeader {
			testResp := new(mongrel2.HttpResponse)
			config.Logger.Printf("[ROUTE TEST] %s : %s\n", target.Name(), req.Path)
			testResp.ClientId = []int{req.ClientId}
			testResp.ServerId = req.ServerId
			testResp.StatusCode = route_test_response_code
			testResp.Header = map[string]string{route_test_header: target.Name()}
			testResp.StatusMsg = "Thanks for testing with seven5"
			self.Out <- testResp
			continue
		}

		resp := protectedProcessRequest(config, req, target)
		self.Out <- resp
	}
}

//protectedProcessRequest is present to allow us to trap panic's that occur inside the web 
//application.  The web application should not really ever do this, it should generate 500
//pages instead but a nil pointer dereference or similar is possible.
func protectedProcessRequest(config *projectConfig, req *mongrel2.HttpRequest, target Httpified) (resp *mongrel2.HttpResponse) {
	defer func() {
		if x := recover(); x != nil {
			config.Logger.Printf("[%s]: PANIC! sent error page for %s: %v\n", target.Name(), req.Path, x)
			resp = new(mongrel2.HttpResponse)
			resp.StatusCode = 500
			resp.StatusMsg = "Internal Server Error"
			b := fmt.Sprintf("Panic: %v\n", x)
			resp.ContentLength = len(b)
			resp.Body = strings.NewReader(b)
		}
	}()
	resp = target.ProcessRequest(req)
	config.Logger.Printf("[%s]: responded to %s with %d bytes of content\n", target.Name(), req.Path, resp.ContentLength)
	return
}

//Generate500Page returns an error page as a mongrel2.Response.  This includes a call stack of the point
//where the caller called this function.
func Generate500Page(err string, request *mongrel2.HttpRequest) *mongrel2.HttpResponse {
	fiveHundred := new(mongrel2.HttpResponse)

	fiveHundred.ServerId = request.ServerId
	fiveHundred.ClientId = []int{request.ClientId}

	fiveHundred.StatusCode = 500
	fiveHundred.StatusMsg = "Internal Server Error"
	b := generateStackTrace(fmt.Sprintf("%v", err))
	fiveHundred.Body = strings.NewReader(b)
	fiveHundred.ContentLength = len(b)
	return fiveHundred
}

//generateStackTrace knows how to create a string from an error by using runtime.Caller and filtering
//out a couple of calls, such as itself.
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

//Shutdown here is a bit trickier than it might look.  This sends the shutdown message
//to the write loop.  The read loop would never see the read message if you closed it here
//because it is blocked waiting on the socket.  So, when the context is closed the 
//read loop will catch the ETERM error and close the channel.
func (self *HttpRunnerDefault) Shutdown() {
	//this check is needed because if you call shutdown before things get rolling, you'll
	//try to close a nil channel
	if self.Out != nil {
		close(self.Out)
	}
}
