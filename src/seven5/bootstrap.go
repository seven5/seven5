package seven5

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"seven5/util"
	"strings"
	"fmt"
)

const (
	SEVEN5_CONFIG = "seven5.json"
)

// Bootstrap is responsible for two tasks.  First, insuring that the
// user project can compile as a library.  Second, for building and then
// invoking a working Seven5Drumkit
type Bootstrap struct {
	Writer  http.ResponseWriter
	Request *http.Request
	Logger  util.SimpleLogger
}

//Run is invoked from the webserver to tell us that the user wants to 
//try to build and run their project.  It kicks-off all the processing
//that Bootstrap is responsible for.  
func (self *Bootstrap) Run() {
	self.Logger = util.NewHtmlLogger(util.DEBUG, true, self.Writer)
	var cwd string
	var err error
	var file *os.File


	if cwd, err = os.Getwd(); err != nil {
		self.Logger.Panic("unable to get the current working directory!")
	}

	configPath := filepath.Join(cwd, "seven5.json")
	self.Logger.Debug("checking cwd (%s) for config file %s", cwd, configPath)

	if file, err = os.Open(configPath); err != nil {
		self.Logger.Error("cannot bootstrap without a seven five configuration file!")
		self.Logger.Error("%s!", err)
		return
	}

	var jsonBuffer bytes.Buffer
	if _, err = jsonBuffer.ReadFrom(file); err != nil {
		self.Logger.Error("error trying to read seven5 config file: %s!", err)
		return
	}

	self.Logger.Debug("Json configuration:")
	self.Logger.DumpJson(jsonBuffer.String())
	config := bootstrapConfiguration(jsonBuffer.String(), self.Logger)
	if config == nil {
		self.Logger.Error("Bad json format for project config, aborting")
	}

}

// bootstrapConfiguration is called to read a set of configuration values
// into a json structore. It returns nil if the format is not satisfactory.
// Note that this does not check semantics!
func bootstrapConfiguration(jsonBlob string, logger util.SimpleLogger) *ProjectConfig {
	decoder := json.NewDecoder(strings.NewReader(jsonBlob))
	var project ProjectConfig
	decoder.Decode(&project)
	//if any plugins are not defined, we definitely have a problem
	seemsOk := true
	switch {
	case project.Plugins.ProjectValidator == "":
		logger.Error("No ProjectValidator plugin found!")
		seemsOk = false
	}
	if !seemsOk {
		return nil
	}
	return &project
}


//pill generates the pill in a temp directory and compiles it.  It returns
//the name of the seven5 command or "" if it failed.
func bootstrapSeven5(config *ProjectConfig, logger util.SimpleLogger) string {
	var cmd string
	var errText string
	
	mainCode := fmt.Sprintf(bootstrapTemplate,
		config.Plugins.ProjectValidator)

	if cmd,errText = util.CompilePill(mainCode, logger); cmd=="" {
		logger.DumpTerminal(errText)
		logger.Panic("Unable to compile the seven5pill! Aborting!")
	}	
	logger.Info("Seven5 is now %s",cmd)	
	return cmd
}

const bootstrapTemplate =
`
package main
import (
	"os"
	"seven5/plugins"
)

func main() {
	plugins.Seven5PillConfig(&%s{}/*ProjecValidator*/)
	plugins.Seven5PillGo(os.Args...)
}
`