package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"seven5"
	"seven5/util"
	"time"
)

//wire is our connection to the seven5 binary
type roadieState struct {
	cfg *configWatcher
	src *sourceWatcher
	haveLibrary bool
	types *seven5.ExplodeTypeResult
	action *seven5.BuildAction //includes the wire
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
	
	state.action = &seven5.BuildAction{}
	
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
	
	http.HandleFunc("/seven5/build", buildPhase1)

	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(os.Stderr, "internal seven5 err:%s\n", r)
		}
	}()
	
	fmt.Fprintf(os.Stderr,"roadie error waiting on connections: %s",
		s.ListenAndServe().Error())
}


// buildPhase1 is responsible for getting seven5 built, if possible, and
// for building the user library, if possible.
func buildPhase1(writer http.ResponseWriter, request *http.Request) {
	firstPass:=false
	builtSeven5:=false
	
	logger:=util.NewHtmlLogger(util.DEBUG, writer, true) 
	//do we have a wire?
	if state.action.Wire==nil {
		logger.Info("No seven5 built yet, building now")
		firstPass=true
		if !state.action.BuildSeven5(writer, request, logger) {
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
		if !state.action.BuildSeven5(writer,request,logger) {
			return //can't build seven5
		}
	}
	//at this point, seven5 has been updated if needed and we have a wire
	//capable of actually reaching our executable
	if appChanged || firstPass{
		logger.Info("Validating that the project has the right project structure.")
		if state.action.ValidateProject(workingDirectory, writer, request, logger)!=nil{
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
		state.types = nil //signal that we need help from types later
		if !state.haveLibrary {
			logger.Info("No library present, rebuilding")
		}
		if source {
			logger.Info("Detected change to source code, building user library in %s", 
				workingDirectory)
		}
		if state.action.BuildUserLib(workingDirectory, writer, request, logger)!=nil {
			return // can't build .a for user code
		}
		state.haveLibrary = true
		//check again, might be another change
		source, err = state.src.pollAllSource(writer, request, logger)
		if err!=nil {
			return // can't read source dir or other IO problem
		}		
	}
	
	logger.Info("User library is ok.")
	buildPhase2(writer, request, logger)
}

//buildPhase2 is responsible for exploding types and if possible building
//the necessary add-on source code.
func buildPhase2(writer http.ResponseWriter, request *http.Request, logger util.SimpleLogger) {	
	explodeArg:=&seven5.ExplodeTypeArg{}
	
	if state.types==nil {
		//get everythnig
		logger.Debug("nothing known about types, need to explode all types.")
		explodeArg.Vocab = state.src.vocab.GetFileList()
		dumpVocabs(explodeArg.Vocab, logger)
	} else {
		vocab, err:=state.src.pollVocab(writer,request,logger)
		if err!=nil {
			logger.Error("Error trying to determine if vocabularies changed:%s",err)
			return // can't read source dir or other IO problem
		}	
		if len(vocab)==0 {
			logger.Debug("No vocab has changed (size of vocab is zero)")
		} else {
			dumpVocabs(vocab, logger)
		}	
		if len(vocab)>0 {
			logger.Info("Vocabularies to be processed: %v\n", vocab)
			explodeArg.Vocab = vocab
			state.types = nil
		}
	}
	var err error
	if state.types==nil {
		logger.Info("Need to rebuild and explode types because we don't know anything.")
		state.types, err=state.action.ExplodeType(workingDirectory,writer,request,
			explodeArg,logger)
		if err!=nil {
			logger.Debug("Error trying to explode type: %s",err)
			return 
		}
	} else {
		//look for changed parts of the source
	}
	
	if len(explodeArg.Vocab)>0 {
		logger.Debug("Need to rebuild the vocabulary support code since we " +
			"just exploded the types.")
		pvArg := &seven5.ProcessVocabArg{}
		pvArg.Info = state.types.Vocab
		_, err =state.action.ProcessVocab(workingDirectory,writer,request,
			pvArg,logger)
		if err!=nil {
			return
		}
	}
}


func dumpVocabs(vocab []string, logger util.SimpleLogger) {
	vlist:=""
	for _, v:= range vocab {
		vlist=vlist+"'"+v+"' "
	}
	logger.Debug("Changed vocabs (%d): %s", len(vocab), vlist)
}


