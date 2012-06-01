package seven5

import (
	"fmt"
	"net/http"
	"os/exec"
	"seven5/util"
)

// Return type of build is a yes/no
type BuildUserLibResult struct {
	Error bool
}

// DefaultBuildUser lib uses go build to test that the library is buildable. If
// it builds, it installs into your gopath as .a file.
type DefaultBuildUserLib struct {
}

func (self *DefaultBuildUserLib) Exec(command string, dir string,
	config *ApplicationConfig, request *http.Request,
	log util.SimpleLogger) interface{} {

	cmd := exec.Command("go", "install", config.AppName)
	out, err := cmd.CombinedOutput()

	if err != nil {
		msg := fmt.Sprintf("Unable to compile %s:%s", config.AppName, err.Error())
		log.DumpTerminal(util.ERROR, msg, string(out))
		return &BuildUserLibResult{Error: true}
	}

	return &BuildUserLibResult{Error: false}
}
