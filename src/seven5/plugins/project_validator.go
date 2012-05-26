package plugins

import (
	"path/filepath"
	"os"
	"seven5/util"
)

// ProjectValidator checks to see if the layout of the project is
// acceptable for the next phase.
type ProjectValidator interface{
	Validate(currentWorkingDir string, log util.SimpleLogger) bool
}

// Default project validator looks for the directory structure
// app
type DefaultProjectValidator struct {
}

func (self *DefaultProjectValidator) Validate(cwd string, log util.SimpleLogger) bool {
	var err error
	var stat os.FileInfo
	
	log.Debug("Using DefaultProjectValidator in %s",cwd)
	
	appPath := filepath.Join(cwd,"app")
	
	if stat,err = os.Stat(appPath); err!=nil {
		log.Error("failed to find app directory: %s", err)
		return false
	}
	
	if !stat.IsDir() {
		log.Error("found %s but it is not a directory!", appPath)
		return false
	}
	
	return true
}
