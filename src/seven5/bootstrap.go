package seven5

import (
	"bytes"
	"net/http"
	"seven5/util"
	"fmt"
)

const (
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
	var groupieJson string
	var err error
	
	self.Logger = util.NewHtmlLogger(util.DEBUG, true, self.Writer)

	self.Logger.Debug("checking for groupies config file...")
	groupieJson, err = FindGroupieConfigFile()	
	if  err != nil {
		self.Logger.Error("unable find or open the groupies config:%s", err)
		return
	}
	self.Logger.Debug("Groupies configuration:")
	self.Logger.DumpJson(groupieJson)
	if _, err := getGroupies(groupieJson, self.Logger); err!=nil {
		return
	}
}

// bootstrapConfiguration is called to read a set of groupie values
// from json to a config structures. It returns nil if the format is not 
// satisfactory.  Note that this does not check semantics!
func getGroupies(jsonBlob string, logger util.SimpleLogger) (GroupieConfig, error){
	var result GroupieConfig
	var err error
	if result, err = ParseGroupieConfig(jsonBlob); err!=nil {
		logger.Error(err.Error())		
		return nil, err
	} 
	return result,nil
}


//pill generates the pill in a temp directory and compiles it.  It returns
//the name of the seven5 command or "" if it failed.
func bootstrapSeven5(config GroupieConfig, logger util.SimpleLogger) string {
	var cmd string
	var errText string
	var imports bytes.Buffer
	
	//gather all includes
	for _,v := range(config) {
		for _, i:= range(v.ImportsNeeded) {
			imports.WriteString(fmt.Sprintf("include \"%s\"\n",i))
		}
	}
	mainCode := fmt.Sprintf(bootstrapTemplate,
		imports, 
		config["ProjectValidator"].TypeName)

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
%s

func main() {
	plugins.Seven5PillConfig(&%s{}/*ProjecValidator*/)
	plugins.Seven5PillGo(os.Args...)
}
`