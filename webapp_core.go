package seven5

import (
	"errors"
	"fmt"
	"github.com/seven5/gozmq"
	"github.com/bradfitz/gomemcache/memcache"
	"log"
	"github.com/seven5/mongrel2"
	"os"
	"seven5/store"
	"time"
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

//Named is use to indicate that your object has a name method and should be placed into the
//URL space.  This is used to avoid specific knowlege of the concrete type in some places.
type Named interface {
	//Name should return the name of the handler.
	Name() string
	//Shutdown is called to give an opportunity for resources to be closed as the application is
	//going down.  This most likely gonig to be called when a test is shutting down as web applications
	//typically run forever.
	Shutdown()
	//Should return "" if you want to be placed in the default place in the URL space (usually /api/Name()).
	//Most user code should return "" and only system code should muck around with this value.
	Pattern() string
}

//AppStarter means that the object in question would like to receive a message at application 
//startup time.  The passed objects are a logger that is connected to seven5.log and the storage
//in use for the application, if you need to do some initalization.
type AppStarter interface {
	AppStarting(*log.Logger, store.T) error
}

var systemGuise = []Named{newFaviconGuise(), newLoginGuise(), newModelGuise()}

//StartUp starts handlers running. It starts all the system guises plus the Named
//handlers (user-level) that are provided as a parameter.  Typically user-level
//code should never need this.
//
//The return parameter is the zmq context for this application and this should be closed 
// on shutdown (usually using defer). This functions runs all the Named provided
//via a goroutine.  If something went wrong this returns nil and most web
//apps will just want to exit since the error has already been printed to stderr.  If you are
//calling this from test code, you will want to set the second parameter to the proposed
//project directory; otherwise pass "" .
func startUp(ctx gozmq.Context, privateInit func(*log.Logger,store.T) error, conf *projectConfig, named []Named) bool {

	store := &store.MemcacheGobStore{memcache.New(store.LOCALHOST)}
	
	allNamed := make([]Named, len(systemGuise)+len(named))
	for i, n := range systemGuise {
		allNamed[i] = n
	}
	for i, n := range named {
		allNamed[i+len(systemGuise)] = n
	}

	if privateInit!=nil {
		if err:=privateInit(conf.Logger, store); err!=nil {
			fmt.Fprintf(os.Stderr,"private init failed:%v\n",err)
			return false
		}
	}

	for _, h := range allNamed {
		rh := h.(mongrel2.RawHandler)
		if err := rh.Bind(h.Name(), ctx); err != nil {
			fmt.Fprintf(os.Stderr, "unable to bind %s to socket! %s\n", h.Name(), err)
			return false
		}
		starter, ok := h.(AppStarter)
		if ok {
			starter.AppStarting(conf.Logger, store)
		}
		switch x := h.(type) {
		case Httpified:
			go x.(HttpRunner).RunHttp(conf, x)
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
//so you only need to pass Named objects to this method if you have non-restful things
//to start.  The first parameter should be nil or a function with the signature appropriate
//to be a pwd.PrivateInit() function--normally this is discovered by the "tune" command
//and selected automatically.
func WebAppRun(privateInit func(*log.Logger,store.T) error, named ...Named) error {
	var ctx gozmq.Context
	var err error
	
	//add backbone-REST services into the set of named
	bboneSvc:= backboneServices()
	allNamed:=make([]Named,len(named)+len(bboneSvc))
	for i, n:=range named {
		allNamed[i]=n
	}
	for i, n:=range bboneSvc {
		allNamed[i+len(named)]=n
	}
	
	config,err:=webAppDefaultConfig(allNamed)
	if err!=nil {
		fmt.Fprintf(os.Stderr,"error generating configuration:%v\n",err)
		return err
	}

	//setup the network
	if ctx, err = createNetworkResources(config); err != nil {
		fmt.Fprintf(os.Stderr,"error creating network resources:%v\n",err)
		
		return errors.New(fmt.Sprintf("error starting 0MQ or mongrel:%s", err.Error()))
	}
	if ctx == nil {
		fmt.Fprintf(os.Stderr,"unable to create 0MQ context!")
		
		return errors.New("No ctx was created.\n")
	}
	defer ctx.Close()

	//this uses the logger from the config, so no need to print error messages, it's handled
	//by the callee... 
	if !startUp(ctx, privateInit, config, allNamed) {
		fmt.Fprintf(os.Stderr,"error starting up handlers! Exiting!")
		
		return errors.New(fmt.Sprintf("error starting up the handers:%s", err))
	}
	
	//wait forever in 10 sec increments... need to keep this function alive because when
	//it exits (such as control-c) the context gets closed
	for {
		time.Sleep(10000000000)
	}

	return nil //will never happen
}

//webAppDefaultConfig builds the mongrel2 database needed to run in the default seven5
//configuration for logs, pid files, URL namespace, etc.  The return value is a project
//config or nil plus an error. 
func webAppDefaultConfig(named []Named) (*projectConfig, error) {
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

//Shutdown causes the channels to be closed so the various goroutines (including the system
//guises) can close down their resources.  Normal web applications run forever so they don't
//need this, but tests do.
func Shutdown(named ...Named) {

	allNamed := make([]Named, len(systemGuise)+len(named))
	for i, n := range systemGuise {
		allNamed[i] = n
	}
	for i, n := range named {
		allNamed[i+len(systemGuise)] = n
	}

	for _, bye := range allNamed {
		bye.Shutdown()
	}
}
