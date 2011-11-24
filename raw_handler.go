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

type Named interface {
	Name() string
	M2Handler() mongrel2.M2RawHandler
}

//==StartUp is what most web apps will want to use as an entry point. ===
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
		m2h:=h.M2Handler()
		if m2h==nil {
			m2h=new(mongrel2.M2RawHandlerDefault)
		}
		fmt.Fprintf(os.Stderr,"in Startup %s about to bind\n",h.Name())
		if err:=m2h.Bind(h.Name(),ctx); err!=nil {
			fmt.Fprintf(os.Stderr,"unable to bind %s to socket! %s\n", h.Name(),err)
			return nil
		}
		switch x:=h.(type) {
		case Httpified:
			runner:=x.HttpRunner()
			if runner==nil {
				runner=new(HttpRunnerDefault)
			}
			runner.RunHttp(conf,x)
		default:
			panic(fmt.Sprintf("unknown handler type! %T",h))
		}
	}

	return ctx
}
