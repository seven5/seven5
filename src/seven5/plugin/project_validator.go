package plugin

import (
	"os"
	"path/filepath"
	"seven5/util"
)

//ProjectValidatorArgs is passed to the ProjectValidator to do its
//work.Must be public for json encoding.
type ProjectValidatorArgs struct {
	Path string
}

//ProjectValidatorResult is the result type of a call on the ProjectValidator.
//Must be public for json encoding.
type ProjectValidatorResult struct {
	Result
	Ok bool
}

// ProjectValidator checks to see if the layout of the project is
// acceptable for the next phase.
type ProjectValidator interface {
	Validate(args ProjectValidatorArgs, log util.SimpleLogger) ProjectValidatorResult
}

// Default project validator looks for the directory structure
// app
type DefaultProjectValidator struct {
}

func (self *DefaultProjectValidator) Validate(cwd string, log util.SimpleLogger) bool {
	var err error
	var stat os.FileInfo

	log.Debug("Using DefaultProjectValidator in %s", cwd)

	appPath := filepath.Join(cwd, "app")

	if stat, err = os.Stat(appPath); err != nil {
		log.Error("failed to find app directory: %s", err)
		return false
	}

	if !stat.IsDir() {
		log.Error("found %s but it is not a directory!", appPath)
		return false
	}

	return true
}
