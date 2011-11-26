package seven5

import (
	"fmt"
	"github.com/alecthomas/gozmq"
	"os"
	"mongrel2"
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
}

//StartUp is what most web apps will want to use as an entry point. 
// 
//The return parameter is the zmq context for this application and this should be closed 
// on shutdown (usually using defer). This functions runs all the seven5.RawHandlers provided
//via a goroutine.  If something went wrong this returns nil and most web
//apps will just want to exit since the error has already been printed to stderr.  If you are
//calling this from test code, you will want to set the second parameter to the proposed
//project directory; otherwise pass "" and it will be retreived from the command line args.
func StartUp(raw []Named, proposedDir string) gozmq.Context {
	var conf *ProjectConfig
	var ctx gozmq.Context

	if proposedDir != "" {
		conf, ctx = BootstrapFromDir(proposedDir)
	} else {
		conf, ctx = Bootstrap()
	}
	if conf == nil {
		return nil
	}


	for _, h := range raw {
		rh:=h.(mongrel2.RawHandler)
		if err:=rh.Bind(h.Name(),ctx); err!=nil {
			fmt.Fprintf(os.Stderr,"unable to bind %s to socket! %s\n", h.Name(),err)
			return nil
		}
		switch x:=h.(type) {
		case Httpified:
			go x.(HttpRunner).RunHttp(conf,x)
		case Jsonified:
			go x.(JsonRunner).RunJson(conf,x)
		default:
			panic(fmt.Sprintf("unknown handler type! %T is not Httpified or Jsonified!",h))
		}
	}

	return ctx
}
