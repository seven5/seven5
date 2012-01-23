package seven5

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"seven5/store"
)

const (
	//When doing route testing, this is the result value the framework will give in
	//response to a GET request on a raw handler.
	route_test_response_code = 209
	//When doing route testing, requests should be marked with this header and the
	//header's value should be the name of the handler expected to get it.  When
	//generating a successful response to a test, seven5 also sets this header in the 
	//response.
	route_test_header = "X-Seven5-Route-Test"
)
//WebAppConfig is a structure that represents some parameters of the application that are
//passed in "from the outside".  Typically, the tune application generates this structure
//based on analyzing the source code and perhaps some options passed to tune itself.
type WebAppConfig struct {
	//true if it is ok for superuser to shutdown the server
	AllowShutdown bool
	//storage implementation for this application
	Store store.T
	//private initialization function, normally part of pwd.go--can be nil
	PrivateInit func(*log.Logger, store.T) error
}

//Routable is use to indicate that your object has a name method and should be placed into the
//URL space.  This interface is separate from the Httpified interface because of the need to
//treat http and json services as the same type of thing at the mongrel2 level.  This class
//is typically not needed by client code as they get this by using one of the Default 
//implementations such as RestHandlerDefault.
type Routable interface {
	//Name should return the name of the handler.
	Name() string
	//Shutdown is called to give an opportunity for resources to be closed as the application is
	//going down.  This most likely gonig to be called when a test is shutting down as web applications
	//typically run forever.
	Shutdown()
	//Should return "" if you want to be placed in the default place in the URL space (usually /api/Name()).
	//Most user code should return "" and only system code should muck around with this value.
	Pattern() string
	//AppStarting is called at startup time to allow startup actions to occur.  The logger
	//provided is connected to the seven5.log and the store is the store in use for this
	//application.
	AppStarting(*log.Logger, store.T) error
	//BindToTransport is used to connect your routable to the lower level transport that 
	//mucks about with sockets and other grungy stuff.  For now, this is
	//always mongrel2 and the transport parameter is a zeromq Context object.
	BindToTransport(name string, transport Transport) error
}

//Httpified is an interface indicating that the object in question can process http
//messages.  Note that this uses http.Request and http.Response but that does not indicate
//that the other "Go-ish" parts of the http packages are in use.  Normally user code
//implements this interface via one of the default implementations nad shouldn't see this
//directly.
type Httpified interface {
	Routable
	//ProcessRequest is the critical method.  All the HttpRequests are pushed through this method
	//and each one should generate a response object.  Implementors of this interface are handling
	//HTTP "raw". 
	ProcessRequest(request *http.Request) *http.Response
}

//builtin guises that are enabled by default
var systemGuise = []Routable{newFaviconGuise(), newLoginGuise(), newModelGuise()}

//startUp starts Routables running. It starts all the system guises plus the Routables
//(user-level) that are provided as a parameter.  Typically user-level
//code should never need this.
//
//The first parameter is the transport cruft (zmq context) for this application and it
//is to bind each routable to this transport. This function starts all the Routable provided
//via a goroutine.  If something went wrong this returns nil and most web
//apps will just want to exit since the error has already been printed to stderr.  If you are
//calling this from test code, you will want to set the third parameter to the proposed
//project directory; otherwise pass "" .  The second parameter is the private init function
//for this app, if any; normally the privateInit func is determined by the tune program.
func startUp(transport Transport, privateInit func(*log.Logger, store.T) error, store store.T,
	conf *projectConfig, routeable []Routable) bool {

	if privateInit != nil {
		if err := privateInit(conf.Logger, store); err != nil {
			fmt.Fprintf(os.Stderr, "private init failed:%v\n", err)
			return false
		}
	}

	for _, h := range routeable {
		h.BindToTransport(h.Name(), transport)
		h.AppStarting(conf.Logger, store)
		switch x := h.(type) {
		case Httpified:
			go x.(httpRunner).runHttp(conf, x)
		default:
			panic(fmt.Sprintf("unknown handler type! %T is not Httpified or Jsonified!", h))
		}
	}
	//fmt.Printf("done\n")

	return true
}


//WebAppRun takes the named handlers (often empty) and begins driving HTTP through them.
//Most webapps will call this method to start their app running and it will never return.
//Any return is probably an error or a shutdown request.  This interrogates the backbone
//support code to look for restful services that were registered with BackboneService()
//so you only need to pass Routable objects to this method if you have non-restful things
//to start.  The first parameter is typically supplied by the tune application so
//changing the values supplied should be done elsewhere.
func WebAppRun(webConfig *WebAppConfig, nonRest ...Routable) error {
	var err error

	//allRoutable is the set of all routable in the system. 
	var allRoutable []Routable
	//transport is the transport in use.
	var transport Transport

	AllowShutdown(webConfig.AllowShutdown)

	//add backbone-REST services into the set of routables
	bboneSvc := backboneServices()
	allRoutable = make([]Routable, len(systemGuise)+len(nonRest)+len(bboneSvc))
	copy(allRoutable, systemGuise)
	userRoutable := allRoutable[len(systemGuise):]
	copy(userRoutable, nonRest)
	copy(userRoutable[len(nonRest):], bboneSvc)

	projConfig, err := webAppDefaultConfig(allRoutable)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error generating mongrel2 configuration:%v\n", err)
		return err
	}

	//setup the network
	if transport, err = createNetworkResources(projConfig); err != nil {
		fmt.Fprintf(os.Stderr, "error creating networking/transport resources:%v\n", err)

		return errors.New(fmt.Sprintf("error starting transport (networking or mongrel):%s", err.Error()))
	}
	if transport == nil {
		fmt.Fprintf(os.Stderr, "unable to create networking/transport (zmq context)!")

		return errors.New("No networking/transport was created.\n")
	}
	//this does the work of shutdown... we use a defer func in case the stack is unwound
	//due to a panic.
	defer func() {
		for _, r := range allRoutable {
			r.Shutdown()
		}
		transport.Shutdown()
	}()

	//this uses the logger from the config, so no need to print error messages, it's handled
	//by the callee... 
	if !startUp(transport, webConfig.PrivateInit, webConfig.Store, projConfig, allRoutable) {
		fmt.Fprintf(os.Stderr, "error starting up handlers! Exiting!")

		return errors.New(fmt.Sprintf("error starting up the routables:%s", err))
	}

	//just block until somebody requests a shutdown
	_, ok := <-shutdownChannel
	if ok {
		panic("Should never be sending values through the shutdown channel!")
	}
	//fmt.Printf("channel closed for shutdown\n")
	
	//only happens if there is a shutdown request... our defer func does the work.
	return nil
}

//webAppDefaultConfig builds the mongrel2 database needed to run in the default seven5
//configuration for logs, pid files, URL namespace, etc.  The return value is a project
//config or nil plus an error. 
func webAppDefaultConfig(named []Routable) (*projectConfig, error) {
	var config *projectConfig
	var err error
	host := "localhost"

	//create config and zeromq context
	if config = bootstrap(); config == nil {
		return nil, errors.New("unable to bootstrap seven5")
	}

	//this accepts all the defaults for log placement, pid files, etc.
	if err = generateServerHostConfig(config, host, test_port); err != nil {
		return nil, errors.New(fmt.Sprintf("error writing mongrel2 config: server/host: %s", host))
	}

	//walk the handlers given
	for _, n := range named {
		if err = generateHandlerAddressAndRouteConfig(config, host, n); err != nil {
			return nil, errors.New(fmt.Sprintf("error writing mongrel2 config: address/route %s: %s", n.Name(), err))
		}
	}

	//walk the guises
	for _, g := range systemGuise {
		if err = generateHandlerAddressAndRouteConfig(config, host, g); err != nil {
			return nil, errors.New(fmt.Sprintf("error writing mongrel2 config: address/route (%s guise):%s", g.Name(), err))
		}
	}

	//static content at /static
	if err = generateStaticContentConfig(config, host, static_dir); err != nil {
		return nil, errors.New(fmt.Sprintf("error writing mongrel2 config: static content:%s", err))
	}

	//normally this does nothing unless the DB is completely empty
	if err = generateMimeTypeConfig(config); err != nil {
		return nil, errors.New(fmt.Sprintf("error writing mongrel2 config: mime types:%s", err))
	}

	//finish writing data to disk
	if err = finishConfig(config); err != nil {
		return nil, errors.New(fmt.Sprintf("error finishing mongrel2 config: db close:%s", err))
	}

	return config, nil
}


//BufferCloser is a convenience type for using a buffer as the body of an HTTP response.  This
//is being exposed beacuse bytes.Buffer is a Reader but not a Closer().  This is a simple
//wrapper around bytes.Buffer
type BufferCloser struct {
	*bytes.Buffer
}

//NewBufferCloser creates a new, empty buffer, ala new(bytes.Buffer)
func NewBufferCloser() *BufferCloser {
	return &BufferCloser{new(bytes.Buffer)}
}

//NewBufferCloserFromBytes creates a new buffer pointed at the parameter, ala bytes.NewBuffer([]byte)
func NewBufferCloserFromBytes(b []byte) *BufferCloser {
	return &BufferCloser{bytes.NewBuffer(b)}
}

//NewBufferCloserFromString creates a new buffer pointed at the parameter, ala bytes.NewBuffer(string)
func NewBufferCloserFromString(s string) *BufferCloser {
	return &BufferCloser{bytes.NewBufferString(s)}
}

//Close dumps the storage for the underlying buffer (to prevent inadvertent reuse) 
//but otherwise does nothing.
func (self *BufferCloser) Close() error {
	self.Buffer = nil
	return nil
}

//Len is changed to return an int64 for convenince of using it with Content-Length field
//in an HTTP response.
func (self *BufferCloser) Len() int64 {
	return int64(self.Buffer.Len())
}

//shutdownChannel is used ONLY for signaling that a shutdown has been requested
var shutdownChannel = make(chan int)

//getShutdownNeeded returns channel that can be closed by someone who wants the system to
//go down.  No point in calling this more than once.  Use close(getShutdownChannel()) blow
//everything up.
func getShutdownChannel() chan int {
	return shutdownChannel
}
