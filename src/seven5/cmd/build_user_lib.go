package cmd

import (
	"os/exec"
	"path/filepath"
	"seven5/util"
	"strconv"
	"strings"
)

//
// BuildUserLib is a command to build the user's library.  It dosen't care
// at all what is in the library, it just tries to build it.  If you need
// to perform special operations on teh source, do them before calling
// this command.  BuildUserLib goes to some trouble to display nice error
// messages if the build fails.
//
var BuildUserLib = &CommandDecl{
	Arg: []*CommandArgPair{
		ClientSideWd, //root of the user project
	},
	Ret: BuiltinSimpleReturn,
	Impl: defaultBuildUserLib,
}


func defaultBuildUserLib(log util.SimpleLogger, v...interface{}) interface{} {
	dir:=v[0].(string)
	config,err :=decodeAppConfig(dir)
	if err!=nil {
		log.Error("Couldn't understand the configuration for the app: %s",err)
		return &SimpleErrorReturn{Error:true}
	}
	cmd := exec.Command("go", "install", config.AppName)
	out, err := cmd.CombinedOutput()
	if err != nil {
		//msg := fmt.Sprintf("Unable to compile %s:%s", config.AppName, err.Error())
		//log.DumpTerminal(util.ERROR, msg, string(out))
		log.FileError(compilerErrorsToList(string(out), dir, log))
		return &SimpleErrorReturn{Error: true}
	}

	return &SimpleErrorReturn{Error: false}
}

//compilerErrorToList returns a list of the FilErrorLogItem structs for use
//by the logger
func compilerErrorsToList(compErr string, dir string, logger util.SimpleLogger) *util.BetterList {
	result := util.NewBetterList()
	compileResult := strings.Split(compErr, "\n")

	for _, line := range compileResult {
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}
		msg := strings.Split(line, ":")
		if len(msg) < 3 {
			logger.Error("cant understand compiler output: %s\n!", line)
			continue
		}
		lineNum, err := strconv.Atoi(msg[1])
		if err != nil {
			logger.Error("cant understand compiler output (line#): %s\n", line)
			continue
		}
		longMessage := ""
		for i := 2; i < len(msg); i++ {
			longMessage = longMessage + msg[i]
		}
		item := &util.FileErrorLogItem{Path: filepath.Join(dir, msg[0]), Msg: longMessage,
			Line: lineNum}
		result.PushBack(item)
	}
	return result
}
