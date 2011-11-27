package seven5


const WEBAPP_TEMPLATE = `package main

import (
	"{{.import}}"
	"seven5"
	"time"
	"github.com/alecthomas/gozmq"
	"fmt"
	"os"
)

//Because you can't dynamically load go code yet, you have to use this
//bit of boilerplate. 
func main() {

	{{$pkg=.package}}
	//derive from filenames
	{{range .handler}}
	{{.}} := {{$pkg}}.New{{upper .}}()
	{{end}}
	
	//usually want localhost for testing but maybe should be configurable
	host := seven5.LOCALHOST

	var err error
	var msg string
	var ctx gozmq.Context
	var config *seven5.ProjectConfig

	//create config and zeromq context
	if config, ctx = seven5.Bootstrap(); ctx==nil {
		msg="unable to bootstrap seven5 and create 0MQ context"
		goto abort
	}
	defer ctx.Close()
	
	//this accepts all the defaults for log placement, pid files, etc.
	if err = seven5.GenerateServerHostConfig(config, host, seven5.TEST_PORT); err!=nil {
		msg="error writing mongrel2 config: server/host"
		goto abort
	}
	

	//derive from filenames... need to know if it is json or not
	//host much match the one above
	if err= seven5.GenerateHandlerAddressAndRouteConfig(config, host, echo, false); err!=nil{
		msg="error writing mongrel2 config: address/route (echo)"
		goto abort
	}
	if err=seven5.GenerateHandlerAddressAndRouteConfig(config, host, chat, true); err!=nil{
		msg="error writing mongrel2 config: address/route (chat)"
		goto abort
	}

	//static content at /static
	if err=seven5.GenerateStaticContentConfig(config, host, seven5.STATIC); err!=nil {
		msg="error writing mongrel2 config: static content"
		goto abort
	}

	//normally this does nothing unless the DB is completely empty
	if err=seven5.GenerateMimeTypeConfig(config); err!=nil {
		msg="error writing mongrel2 config: mime types"
		goto abort
	}

	//this uses the logger from the config, so no need to print error messages, it's handled
	//by the callee... 
	if !seven5.StartUp(ctx,config,{{range .handler}}{{.}},{{end}}) {
		goto abort
	}

	//wait forever in 10 sec increments... need to keep this function alive because when
	//it exits (such as control-c) the context gets closed
	for {
		time.Sleep(10000000000)
	}
	
	//cleanup operations, if any.
abort:
	fmt.Fprintf(os.Stderr,"%s\n",msg)
	return
}`
