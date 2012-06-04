package seven5

import (
	"net/http"
	"os"
	"path/filepath"
	"seven5/util"
	"strings"
)

//DestroyGeneratedFileResult tells the caller if we were successful in 
//destroying the code.
//Must be public for json encoding.
type DestroyGeneratedFileResult struct {
	Error bool
}

// Default echo plugin just prints an echo of the values it
// received in the request.
type DefaultDestroyGeneratedFile struct {
}

func (self *DefaultDestroyGeneratedFile) GetArg() interface{} {
	return nil
}

func (self *DefaultDestroyGeneratedFile) Exec(ignored string, dir string,
	config *ApplicationConfig, ignored2 *http.Request, ignored3 interface{},
	log util.SimpleLogger) interface{} {

	log.Info("Destroying Seven5 generated files so user code can be compiled alone.")
	appPath := filepath.Join(dir, "src", config.AppName)
	f, err := os.Open(appPath)
	if err != nil {
		log.Error("Unable to read directory contents: %s", appPath)
		return &DestroyGeneratedFileResult{Error: true}
	}
	name, err := f.Readdir(-1)
	if err != nil {
		log.Error("Error reading directory contents: %s", appPath)
		return &DestroyGeneratedFileResult{Error: true}
	}
	for _, n := range name {
		if strings.HasSuffix(n.Name(), "_generated.go") {
			die := filepath.Join(appPath, n.Name())
			err = os.Remove(die)
			if err != nil {
				log.Error("Error deleting generated code: %s", die)
				return &DestroyGeneratedFileResult{Error: true}
			}
		}
	}
	return &DestroyGeneratedFileResult{}

}
