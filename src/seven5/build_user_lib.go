package seven5

import (
	"net/http"
	"os/exec"
	"seven5/util"
	"strconv"
	"strings"
	"path/filepath"
)

// Return type of build is a yes/no
type BuildUserLibResult struct {
	Error bool
}

// DefaultBuildUser lib uses go build to test that the library is buildable. If
// it builds, it installs into your gopath as .a file.
type DefaultBuildUserLib struct {
}

func (self *DefaultBuildUserLib) GetArg() interface{} {
	return nil
}

func (self *DefaultBuildUserLib) Exec(command string, dir string,
	config *ApplicationConfig, request *http.Request, ignored interface{},
	log util.SimpleLogger) interface{} {

	cmd := exec.Command("go", "install", config.AppName)
	out, err := cmd.CombinedOutput()
	if err != nil {
		//msg := fmt.Sprintf("Unable to compile %s:%s", config.AppName, err.Error())
		//log.DumpTerminal(util.ERROR, msg, string(out))
		log.FileError(compilerErrorsToList(string(out),dir,log))
		return &BuildUserLibResult{Error: true}
	}

	return &BuildUserLibResult{Error: false}
}

//compilerErrorToList returns a list of the FilErrorLogItem structs for use
//by the logger
func compilerErrorsToList(compErr string, dir string, logger util.SimpleLogger) *util.BetterList {
	result := util.NewBetterList()
	compileResult := strings.Split(compErr, "\n")
	
	for _, line := range compileResult {
		if strings.HasPrefix(line, "#") || line=="" {
			continue
		}
		msg := strings.Split(line,":")
		if len(msg)<3 {
			logger.Error("cant understand compiler output: %s\n!",line)
			continue
		}
		lineNum, err:= strconv.Atoi(msg[1])
		if err!=nil {
			logger.Error("cant understand compiler output (line#): %s\n",line)
			continue
		}
		longMessage :=""
		for i:=2; i<len(msg); i++ {
			longMessage = longMessage + msg[i]
		}
		item:=&util.FileErrorLogItem{Path:filepath.Join(dir,msg[0]), Msg:longMessage,
			Line: lineNum}
		result.PushBack(item)
	}	
	return result
}

