package seven5

import (
	"net/http"
	"os"
	"path/filepath"
	"bytes"
)

const (
	SEVEN5_CONFIG = "seven5.json"
)

// Bootstrap is responsible for two tasks.  First, insuring that the
// user project can compile as a library.  Second, for building and then
// invoking a working Seven5Drumkit
type Bootstrap struct {
	Writer http.ResponseWriter
	Request *http.Request
	Logger SimpleLogger
}

//Run is invoked from the webserver to tell us that the user wants to 
//try to build and run their project.  It kicks-off all the processing
//that Bootstrap is responsible for.  
func (self *Bootstrap) Run() {
	self.Logger = NewHtmlLogger(DEBUG,true,self.Writer)
	var cwd string
	var err error
	var file *os.File
	
	if cwd, err = os.Getwd(); err!=nil {
		self.Logger.Panic("unable to get the current working directory!");
	}
	
	configPath := filepath.Join(cwd,"seven5.json")
	self.Logger.Debug("checking cwd (%s) for config file %s",cwd, configPath)
	
	if file, err = os.Open(configPath); err!=nil {
		self.Logger.Error("cannot bootstrap without a seven five configuration file!");
		self.Logger.Error("%s!",err);
		return
	}
	
	var jsonBuffer bytes.Buffer 
	if _,err= jsonBuffer.ReadFrom(file); err!=nil {
		self.Logger.Error("error trying to read seven5 config file: %s!",err);
		return
	}
	
	self.bootstrapConfiguration(cwd,jsonBuffer.Bytes())
}

// bootstrapConfiguration is called to read a set of configuration values
// into a json structore and then try to build that as a binary with the
// appropriate slots fill in as in the json.
func (self *Bootstrap) bootstrapConfiguration(dir string, jsonBlob []byte) {
	
}