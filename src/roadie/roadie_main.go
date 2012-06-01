package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"seven5/util"
	"time"
)

//wire is our connection to the seven5 binary
type roadieState struct {
	cfg *configWatcher
	src *sourceWatcher
	haveLibrary bool
	action *roadieAction //includes the wire
}


var workingDirectory string

var state = &roadieState{}

//main is the entry point of the roadie 
func main() {
	var err error

	if len(os.Args) > 2 {
		fmt.Fprintf(os.Stderr, "%s usage: %s [directory]\n", os.Args[0], os.Args[0])
		return
	}

	if len(os.Args) > 1 {
		if err = os.Chdir(os.Args[1]); err != nil {
			fmt.Fprintf(os.Stderr, "%s: unable to change to %s: %s\n", os.Args[0], os.Args[1], err)
			return
		}
	}

	if workingDirectory, err = os.Getwd(); err != nil {
		fmt.Fprintf(os.Stderr, "%s: unable to get working directory: %s\n", os.Args[0], err)
		return
	}
	
	state.cfg, err = newConfigWatcher(workingDirectory)
	if err!=nil {
		fmt.Fprintf(os.Stderr, "%s: unable to create directory monitor on %s\n",
			workingDirectory)
		return
	}

	state.src, err = newSourceWatcher(workingDirectory)
	if err!=nil {
		fmt.Fprintf(os.Stderr, "%s: unable to create directory monitor on app source\n",
			os.Args[0])
		return
	}
	
	state.action = &roadieAction{}
	
	modifiedGoPath := os.Getenv("GOPATH")+string(filepath.ListSeparator)+workingDirectory
	if err = os.Setenv("GOPATH", modifiedGoPath); err!=nil {
		fmt.Fprintf(os.Stderr, "%s: unable to set GOPATH: %s\n", os.Args[0], err)
		return
	}

	
	s := &http.Server{
		Addr:         ":9009",
		ReadTimeout:  2 * time.Second,
		WriteTimeout: 2 * time.Second,
	}
	
	http.HandleFunc("/seven5/build", build)

	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(os.Stderr, "internal seven5 err:%s\n", r)
		}
	}()
	
	fmt.Fprintf(os.Stderr,"roadie error waiting on connections: %s",
		s.ListenAndServe().Error())
}


// called to build seven5 bythe roadie or by the user
func build(writer http.ResponseWriter, request *http.Request) {
	firstPass:=false
	builtSeven5:=false
	
	logger:=util.NewHtmlLogger(util.DEBUG, writer, true) 
	//do we have a wire?
	if state.action.wire==nil {
		logger.Info("No seven5 built yet, building now")
		firstPass=true
		if !state.action.buildSeven5(writer, request, logger) {
			return //can't build seven5
		}
		builtSeven5=true
	}
	//we now have a wire, do we need to rebuild anything because of config
	//changes
	appChanged, groupieChanged, err := state.cfg.poll(writer,request,logger)
	if err!=nil {
		return //some kind of io problem
	}
	//maybe they changed the seven5 app groupies?
	if groupieChanged && !builtSeven5 {
		logger.Info("Groupie.json configuration files changed, rebuilding seven5")
		if !state.action.buildSeven5(writer,request,logger) {
			return //can't build seven5
		}
	}
	//at this point, seven5 has been updated if needed and we have a wire
	//capable of actually reaching our executable
	if appChanged || firstPass{
		logger.Info("Validating that the project has the right project structure.")
		if state.action.validateProject(workingDirectory, writer, request, logger)!=nil{
			return //app does not check out abandon hope
		}
	}
	//app in now in a sensible state, did the source code change?
	source, err := state.src.pollAllSource(writer, request, logger)
	if err!=nil {
		return // can't read source dir or other IO problem
	}
	//loop until source has stopped changing
	for source || !state.haveLibrary {
		if !state.haveLibrary {
			logger.Info("No library present, rebuilding")
		}
		if source {
			logger.Info("Detected change to source code, building user library in %s", 
				workingDirectory)
		}
		if state.action.buildUserLib(workingDirectory, writer, request, logger)!=nil {
			return // can't build .a for user code
		}
		state.haveLibrary = true
		//check again, might be another change
		source, err = state.src.pollAllSource(writer, request, logger)
		if err!=nil {
			return // can't read source dir or other IO problem
		}		
	}
	
	if state.haveLibrary {
		logger.Info("User library is ok.")
	}
}


