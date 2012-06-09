package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"seven5"
	"seven5/cmd"
	"seven5/util"
	"time"
)

//wire is our connection to the seven5 binary
type roadieState struct {
	cfg         *configWatcher
	src         *sourceWatcher
	haveLibrary bool
	types       *cmd.ExplodeTypeResult
	Wire        *seven5.Wire
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
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: unable to create directory monitor on %s\n",
			workingDirectory)
		return
	}

	state.src, err = newSourceWatcher(workingDirectory)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: unable to create directory monitor on app source\n",
			os.Args[0])
		return
	}

	modifiedGoPath := os.Getenv("GOPATH") + string(filepath.ListSeparator) + workingDirectory
	if err = os.Setenv("GOPATH", modifiedGoPath); err != nil {
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

	fmt.Fprintf(os.Stderr, "roadie error waiting on connections: %s",
		s.ListenAndServe().Error())
}

// buildPhase1 is responsible for getting seven5 built, if possible, and
// for building the user library, if possible.
func buildPhase1(writer http.ResponseWriter, request *http.Request) {
	builtSeven5 := false

	//nice symmetry: the util package contributes a parameter and the
	//command package contributes one.  both are needed to actually dispatch
	//commands to the wire.
	logger := util.NewHtmlLogger(util.INFO, writer, true)
	clientCap := cmd.NewDefaultClientCapability(request)

	fullStart := time.Now()
	defer func() {
		diff := time.Since(fullStart)
		logger.Info("Complete build sequence too %s", diff)
	}()
	//do we have a wire?
	if state.Wire == nil {
		logger.Info("No seven5 built... so building now")
		path, err := seven5.BuildSeven5(logger)
		if err != nil {
			return //can't build seven5
		}
		//we have already built seven5 on this pass, so we want to avoid
		//doing it again
		builtSeven5 = true

		//we are going to use the JsonStringExec strategy...
		state.Wire = wireFromSeven5Path(path)
	}

	//we now have a wire, do we need to rebuild anything because of config
	//changes
	appChanged, commandChanged, err := state.cfg.poll(writer, request, logger)
	if err != nil {
		return //some kind of io problem
	}
	//maybe they changed the seven5 app commands?
	if commandChanged && !builtSeven5 {
		logger.Info("command.json configuration files changed, rebuilding seven5")
		path, err := seven5.BuildSeven5(logger)
		if err != nil {
			state.Wire = nil //just to be sure
			return           //can't build seven5
		}
		state.Wire = wireFromSeven5Path(path)
	}
	//at this point, seven5 has been updated if needed and we have a wire
	//capable of actually reaching our executable
	if appChanged {
		logger.Info("Validating that the project has the right project structure.")
		//note that it is ok to ignore the value here because the only thing is
		//an error which will get converted to WIRE_SEMANTIC_ERROR in the err
		_, err := state.Wire.Dispatch(seven5.VALIDATEPROJECT, writer,
			request, logger, clientCap)
		if err != nil {
			return //app does not check out... abandon hope
		}
	}
	//app in now in a sensible state, did the source code change?
	source, err := state.src.pollAllSource(writer, request, logger)
	if err != nil {
		return // can't read source dir or other IO problem
	}
	//loop until source has stopped changing
	for source || !state.haveLibrary {
		state.types = nil //signal that we need help from types later

		if _, err:=state.Wire.Dispatch(seven5.DESTROYGENERATEDFILE,
			writer, request, logger, clientCap); err != nil {
			return
		}
		if !state.haveLibrary {
			logger.Info("No user library present, rebuilding")
		}
		if source {
			logger.Info("Detected change to source code, building user library in %s",
				workingDirectory)
		}
		//ok to return result because this command just returns a yes/no value
		//and it gets put in the err value
		if _, err:=state.Wire.Dispatch(seven5.BUILDUSERLIB, 
			writer, request, logger, clientCap); err!= nil {
			return // can't build .a for user code
		}
		state.haveLibrary = true
		//check again, might be another change
		source, err = state.src.pollAllSource(writer, request, logger)
		if err != nil {
			return // can't read source dir or other IO problem
		}
	}

	logger.Info("User library is ok.")
	buildPhase2(writer, request, logger, clientCap)
}

//buildPhase2 is responsible for exploding types and if possible building
//the necessary add-on source code.
func buildPhase2(writer http.ResponseWriter, request *http.Request, logger util.SimpleLogger,
	clientCap cmd.ClientSideCapability) {
	var vocab []string
	var err error
	var rawResult interface{}
	exploded := false
	
	if state.types!=nil {
		//did vocabs change?
		vocab, err = state.src.pollVocab(writer, request, logger)
		if err != nil {
			logger.Error("Error trying to determine if vocabularies changed:%s", err)
			return // can't read source dir or other IO problem
		}
		if len(vocab) == 0 {
			logger.Debug("No vocab has changed (size of vocab is zero)")
		} else {
			dumpVocabs(vocab, logger)
		}
		if len(vocab) > 0 {
			logger.Info("Vocabularies to be processed: %v\n", vocab)
			state.types = nil
		}
	}
	if state.types == nil ||len(vocab)!=0 {
		exploded = true
		if state.types==nil {
			logger.Info("Need to rebuild and explode types because we don't know anything.")
		}
		rawResult, err = state.Wire.Dispatch(seven5.EXPLODETYPE, 
		writer, request, logger, clientCap)
		result:=rawResult.(*cmd.ExplodeTypeResult)
		if err != nil {
			logger.Debug("Error trying to explode type: %s", err)
			return
		}
		//we now have the results
		state.types = result
	} else {
		//look for other changed parts of the source that need type data
	}

	//make sure that any future command knows our types, recomputed or not
	clientCap.SetTypeInfo(state.types);

	if exploded {
		logger.Debug("Need to rebuild the vocabulary support code since we " +
			"just exploded the types.")
		_, err = state.Wire.Dispatch(seven5.PROCESSVOCAB, writer, request,
			logger, clientCap)
		if err != nil {
			return
		}
	}
	
}

func dumpVocabs(vocab []string, logger util.SimpleLogger) {
	vlist := ""
	for _, v := range vocab {
		vlist = vlist + "'" + v + "' "
	}
	logger.Debug("Changed vocabs (%d): %s", len(vocab), vlist)
}

//wireFromSeven5Path converts the path to a seven5 executable into a Wire that
//can talk to it, via the JsonStringExecStrategy
func wireFromSeven5Path(path string) *seven5.Wire {
	//we are going to use the JsonStringExec strategy...
	strategy := seven5.NewJsonStringStrategy(path)
	return seven5.NewWire(strategy)
}
