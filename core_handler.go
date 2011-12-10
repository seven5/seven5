package seven5

import (
	"fmt"
	"github.com/alecthomas/gozmq"
	"os"
	"mongrel2"
	"errors"
	"time"
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

//Named is use to indicate that your object has a name method.  This is used to avoid specific
//knowlege of the concrete type in some places.
type Named interface {
	Name() string
	IsJson() bool
}

//For guises, they don't need names but they do need app startup info
type Guise interface {
	Named
	AppStarting(*ProjectConfig) error
	Pattern() string
}

var SystemGuise = []Guise{NewCssGuise(),NewFaviconGuise(), NewHtmlGuise()}

//StartUp starts handlers running. It starts all the system guises plus the Named
//handlers (user-level) that are provided as a parameter.
//
//The return parameter is the zmq context for this application and this should be closed 
// on shutdown (usually using defer). This functions runs all the Named provided
//via a goroutine.  If something went wrong this returns nil and most web
//apps will just want to exit since the error has already been printed to stderr.  If you are
//calling this from test code, you will want to set the second parameter to the proposed
//project directory; otherwise pass "" and it will be retreived from the command line args.
func StartUp(ctx gozmq.Context, conf *ProjectConfig, named []Named) bool {
	
	allNamed := make([]Named,len(SystemGuise)+len(named))
	for i,n:=range SystemGuise {
		allNamed[i]=n
	}
	for i,n:=range named {
		allNamed[i+len(SystemGuise)]=n
	}
	
	//fmt.Printf("Starting...")
	for _, h := range allNamed {
		rh:=h.(mongrel2.RawHandler)
		if err:=rh.Bind(h.Name(),ctx); err!=nil {
			fmt.Fprintf(os.Stderr,"unable to bind %s to socket! %s\n", h.Name(),err)
			return false
		}
		//fmt.Printf("%s...",h.Name())
		switch x:=h.(type) {
		case Httpified:
			go x.(HttpRunner).RunHttp(conf,x)
		case Jsonified:
			go x.(JsonRunner).RunJson(conf,x)
		default:
			panic(fmt.Sprintf("unknown handler type! %T is not Httpified or Jsonified!",h))
		}
	}
	//fmt.Printf("done\n")

	return true
}

//WebAppRun takes the named handlers and begins driving HTTP or Json requests through them.
//Most webapps will call this method to start their app running and it will never return.
//Any return is probably an error.
func WebAppRun(config *ProjectConfig, named ... Named) error {
    fmt.Println("Web app running")
	var ctx gozmq.Context
	var err error
	
	//setup the network
	if ctx, err=CreateNetworkResources(config); err!=nil {
		return errors.New(fmt.Sprintf("error starting 0MQ or mongrel:%s",err.Error()))
	}
	if ctx == nil {
		return errors.New("No ctx was created.\n")
	}
	defer ctx.Close()

	//this uses the logger from the config, so no need to print error messages, it's handled
	//by the callee... 
	if !StartUp(ctx, config, named) {
		return errors.New(fmt.Sprintf("error starting up the handers:%s",err))
	}

	//wait forever in 10 sec increments... need to keep this function alive because when
	//it exits (such as control-c) the context gets closed
	for {
		time.Sleep(10000000000)
	}

	return nil//will never happen
}

//WebAppDefaultConfig builds the mongrel2 database needed to run in the default seven5
//configuration for logs, pid files, URL namespace, etc.  The return value is a project
//config or nil plus an error.  Most webapps will want to accept this default configuration
//unless they are doing mongrel2-level hacks.
func WebAppDefaultConfig(named ... Named) (*ProjectConfig,error) {
	var config *ProjectConfig
	var err error
	host:="localhost"
	
	
	//create config and zeromq context
	if config= Bootstrap(); config == nil {
		return nil, errors.New("unable to bootstrap seven5")
	}

	//this accepts all the defaults for log placement, pid files, etc.
	if err = GenerateServerHostConfig(config, host, TEST_PORT); err != nil {
		return nil,errors.New(fmt.Sprintf("error writing mongrel2 config: server/host: %s",host))
	}

	//walk the handlers given
	for _,n:=range named {
		if err = GenerateHandlerAddressAndRouteConfig(config, host, n); err != nil {
			return nil, errors.New(fmt.Sprintf("error writing mongrel2 config: address/route %s: %s",n.Name(),err))
		}
	}

	//walk the guises
	for _,g:=range SystemGuise {
 		if err = GenerateHandlerAddressAndRouteConfig(config, host, g); err != nil {
			return nil, errors.New(fmt.Sprintf("error writing mongrel2 config: address/route (%s guise):%s",g.Name(),err))
		}
	}

	//static content at /static
	if err = GenerateStaticContentConfig(config, host, STATIC); err != nil {
		return nil,errors.New(fmt.Sprintf("error writing mongrel2 config: static content:%s",err))
	}

	//normally this does nothing unless the DB is completely empty
	if err = GenerateMimeTypeConfig(config); err != nil {
		return nil,errors.New(fmt.Sprintf("error writing mongrel2 config: mime types:%s",err))
	}

	//finish writing data to disk
	if err = FinishConfig(config); err!=nil {
		return nil,errors.New(fmt.Sprintf("error finishing mongrel2 config: db close:%s",err))
	}
	
	return config,nil
}

//ShutdownGuises releases the network resources associated with the system Guises.  Note that this
//does not shutdown user defined handlers and shutting down such handlers is the responsibility
//of test structures or user code.
func ShutdownGuises() error {
	for _,g:=range SystemGuise {
		rh:=g.(mongrel2.RawHandler)
		if err:=rh.Shutdown(); err!=nil {
			return err
		}
	}
	return nil
}
