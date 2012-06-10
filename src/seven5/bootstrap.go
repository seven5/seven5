package seven5

import (
	"bytes"
	"fmt"
	"path/filepath"
	"seven5/util"
	"strings"
	"encoding/json"
)

const (
	COMMAND_CONFIG_FILE = "command.json"
)

// A Command plays a role in the system. These roles are well defined and
// bound to structures that will be passed to the command playing that
// role.  Public for json-encoding.
type Command struct {
	Role string
	Info CommandInfo
}

//CommandInfo is extra information about a particular groupie. This is 
//information used/needed at bootstrap time.  Public for json encoding.
type CommandInfo struct {
	TypeName      string
	ImportsNeeded []string
}

//CommandWrapper is used just to allow to have only one top level declaration
//in the COMMAND_CONFIG_FILE.
type CommandWrapper struct {
	CommandConfig []Command
}

//CommandConfig is the result of parsing the json.
type commandConfig map[string]*CommandInfo

//simulate const array
func DEFAULT_IMPORTS() []string {
	return []string{"fmt", "seven5", "os"}
}

// Bootstrap is responsible for buliding the current seven5 executable
// based on the command configuration.  He logs all errors so callers can
// just return if they receive an error.
type bootstrap struct {
	logger  util.SimpleLogger
}


//configureSeven5 checks for a goroupie config file and returns a config or
//nil in the error case. pass "" to use current working dir.
func (self *bootstrap) configureSeven5(dir string) (commandConfig,error) {

	var commandJson string
	var err error
	var result commandConfig

	commandJson, err = util.ReadIntoString(dir, COMMAND_CONFIG_FILE)
		
	if err != nil {
		self.logger.Error("unable find or open the command config:%s", err)
		return nil, err
	}
	self.logger.DumpJson(util.DEBUG, "Command configuration", commandJson)

	if result, err = self.parseCommandConfig(commandJson); err != nil {
		self.logger.DumpJson(util.ERROR,"Command configuration", commandJson)
		self.logger.Error("could not understand "+COMMAND_CONFIG_FILE+"! aborting!")
		return nil, err
	}

	return result, nil
}

//parseCommandConfig takes a bunch of json and turns it into a commandConfig.
//It returns an error if you don't supply a sensible configuration.
func (self *bootstrap) parseCommandConfig(jsonBlob string) (commandConfig, error) {
	result := make(map[string]*CommandInfo)
	dec := json.NewDecoder(strings.NewReader(jsonBlob))
	var wrapper CommandWrapper
	if err := dec.Decode(&wrapper); err != nil {
		return nil, err
	} 
	for _, raw := range wrapper.CommandConfig {
		result[raw.Role] = &CommandInfo{raw.Info.TypeName, raw.Info.ImportsNeeded}
	}
	return result, nil
}


//takeSeven5 generates the pill in a temp directory and compiles it.  It returns
//the name of the new seven5 command or "" if it failed.
func (self *bootstrap) takeSeven5Pill(config commandConfig) (string,error) {
	var cmd string
	var errText string
	var imports bytes.Buffer
	var setStatement bytes.Buffer
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

	//walk all the configed groupies
	for k, v := range config {
		setStatement.WriteString(fmt.Sprintf(
			"\tseven5.Seven5app[seven5.%s]=%s\n",
			strings.ToUpper(k), v.TypeName))
	}

	mainCode := fmt.Sprintf(seven5pill,
		imports.String(),
		setStatement.String())

	self.logger.DumpTerminal(util.DEBUG, "Main code for seven5 pill", mainCode)

	if cmd, errText, err = util.CompilePill(mainCode, self.logger); cmd == "" {
		self.logger.DumpTerminal(util.ERROR, "Bogus seven5 pill code", mainCode)
		if errText!="" {
			self.logger.DumpTerminal(util.ERROR, "Unable to compile the seven5pill!",
				errText)
		}
		if err!=nil {
			self.logger.Error("Internal seven5 error: %s",err)
		}
		return "", err
	}
	path := strings.Split(cmd, string(filepath.Separator))
	self.logger.Info("Seven5 is now [tmpdir]/%s", path[len(path)-1])

	return cmd, nil
}

//seven5pill is the text of the pill
const seven5pill = `
package main
%s

func main() {
%s
	if len(os.Args)<3 {
		os.Exit(1)
	}
	//double percent bceause run through sprintf twice
	fmt.Fprintf(os.Stdout,"%%s\n",seven5.RunCommand(os.Args[1], os.Args[2], os.Args[3:]...))
	os.Stdout.Sync()
}
`
