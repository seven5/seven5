package seven5

import (
	"bytes"
	"fmt"
	"mongrel2"
	"strings"
	"runtime"
)

type HttpRunner interface {
	mongrel2.M2HttpHandler
	RunHttp(config *ProjectConfig, target Httpified)
}

//Httpified is an interface indicating that the object in question can process mongrel2
//messages as expected for a mongrel2 handler.  
type Httpified interface {
	Named
	ProcessRequest(request *mongrel2.M2HttpRequest) *mongrel2.M2HttpResponse
}

type HttpRunnerDefault struct {
	*mongrel2.M2HttpHandlerDefault
}


//Called to launch the processing of the HTTP protocol via goroutines.  This call will 
//never return.  It will repeatedly call ProcessRequest() as messages arrive and need
//to be sent to the mongrel2 server.  It is careful to protect itself from code that
//might call panic() even though this is _not_ advised for implementors; it is preferred
//to have implementations that detect a problem use the HTTP error code set.
func (self *HttpRunnerDefault) RunHttp(config *ProjectConfig, target Httpified) {

	in := make(chan *mongrel2.M2HttpRequest)
	out := make(chan *mongrel2.M2HttpResponse)

	go self.ReadLoop(in)
	go self.WriteLoop(out)

	for {
		//block until we get a message from the server
		req := <-in

		if req == nil {
			config.Logger.Printf("[%s]: close of mongrel2 connection in raw handler!", target.Name())
			return
		} else {
			config.Logger.Printf("[%s]: serving %s", target.Name(), req.Path)
		}

		//note: mongrel converts this to lower case!
		testHeader := req.Header[strings.ToLower(ROUTE_TEST_HEADER)]
		if target.Name() == testHeader {
			testResp := new(mongrel2.M2HttpResponse)
			config.Logger.Printf("[ROUTE TEST] %s : %s\n", target.Name(), req.Path)
			testResp.ClientId = []int{req.ClientId}
			testResp.ServerId = req.ServerId
			testResp.StatusCode = ROUTE_TEST_RESPONSE_CODE
			testResp.Header = map[string]string{ROUTE_TEST_HEADER: target.Name()}
			testResp.StatusMsg = "Thanks for testing with seven5"
			out <- testResp
			continue
		}

		resp := protectedProcessRequest(config, req, target)
		out <- resp
	}
}

//protectedProcessRequest is present to allow us to trap panic's that occur inside the web 
//application.  The web application should not really ever do this, it should generate 500
//pages instead but a nil pointer dereference or similar is possible.
func protectedProcessRequest(config *ProjectConfig, req *mongrel2.M2HttpRequest, target Httpified) (resp *mongrel2.M2HttpResponse) {
	defer func() {
		if x := recover(); x != nil {
			config.Logger.Printf("[%s]: PANIC! sent error page for %s: %v\n", target.Name(), req.Path, x)
			resp = new(mongrel2.M2HttpResponse)
			resp.StatusCode = 500
			resp.StatusMsg = "Internal Server Error"
			resp.Body = fmt.Sprintf("Panic: %v\n", x)
		}
	}()
	resp = target.ProcessRequest(req)
	config.Logger.Printf("[%s]: responded to %s with %d bytes of content\n", target.Name(), req.Path, len(resp.Body))
	return
}

//Generate500Page returns an error page as a mongrel2.Response.  This includes a call stack of the point
//where the caller called this function.
func Generate500Page(err string, request *mongrel2.M2HttpRequest) *mongrel2.M2HttpResponse {
	fiveHundred := new(mongrel2.M2HttpResponse)

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
