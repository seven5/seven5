package seven5

import (
	"bytes"
	"fmt"
	"net/http"
	"seven5/util"
	"os"
)

const ()

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
	var conf groupieConfig

	self.Logger = util.NewHtmlLogger(util.DEBUG, true, self.Writer)

	self.Logger.Debug("checking for groupies config file...")
	groupieJson, err = findGroupieConfigFile()
	if err != nil {
		self.Logger.Error("unable find or open the groupies config:%s", err)
		return
	}
	self.Logger.Debug("Groupies configuration:")
	self.Logger.DumpJson(groupieJson)
	if conf, err = getGroupies(groupieJson, self.Logger); err != nil {
		return
	}

	bootstrapSeven5(conf, self.Logger)
}

// bootstrapConfiguration is called to read a set of groupie values
// from json to a config structures. It returns nil if the format is not 
// satisfactory.  Note that this does not check semantics!
func getGroupies(jsonBlob string, logger util.SimpleLogger) (groupieConfig, error) {
	var result groupieConfig
	var err error
	if result, err = parseGroupieConfig(jsonBlob); err != nil {
		logger.Error(err.Error())
		return nil, err
	}
	return result, nil
}

//simulate const array
func DEFAULT_IMPORTS() []string {
	return []string{"fmt", "seven5/plugin", "os"}
}

//pill generates the pill in a temp directory and compiles it.  It returns
//the name of the seven5 command or "" if it failed.
func bootstrapSeven5(config groupieConfig, logger util.SimpleLogger) string {
	var cmd string
	var errText string
	var imports bytes.Buffer
	var cwd string
	var err error
	
	seen := util.NewBetterList()
	for _, i := range DEFAULT_IMPORTS() {
		seen.PushBack(i)
	}
	//gather all includes
	for _, v := range config {
		for _, i := range v.ImportsNeeded {
			if seen.Contains(i) {
				continue
			}
			seen.PushBack(i)
		}
	}
	for e := seen.Front(); e != nil; e = e.Next() {
		imports.WriteString(fmt.Sprintf("import \"%s\"\n", e.Value))
	}

	if cwd, err = os.Getwd(); err!=nil {
		logger.Panic("Unable to get the current working directory!")
	}
	mainCode := fmt.Sprintf(seven5pill,
		imports.String(),
		config["ProjectValidator"].TypeName,cwd)

	if cmd, errText = util.CompilePill(mainCode, logger); cmd == "" {
		logger.DumpTerminal(errText)
		logger.Warn("Unable to compile the seven5pill! Aborting!")
	}
	logger.Info("Seven5 is now %s", cmd)
	return cmd
}

const seven5pill = `
package main
%s

func main() {
	plugin.Seven5App.SetProjectValidator(&%s{})
	if len(os.Args)<3 {
		os.Exit(1)
	}
	fmt.Println(plugin.Run("%s",os.Args[1], os.Args[2]))
}
`
