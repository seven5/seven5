package plugin

import (
	"os"
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
}

// ProjectValidator checks to see if the layout of the project is
// acceptable for future phases.
type ProjectValidator interface {
	Validate(cmd *Command, args *ProjectValidatorArgs, 
		log util.SimpleLogger) *ProjectValidatorResult
}

// Default project validator looks for the directory structure
// app
type DefaultProjectValidator struct {
}

func (self *DefaultProjectValidator) Validate(cmd *Command, args *ProjectValidatorArgs, 
log util.SimpleLogger) *ProjectValidatorResult {
	var err error
	var stat os.FileInfo

	log.Debug("Using DefaultProjectValidator in %s", cmd.AppDirectory)

	if stat, err = os.Stat(cmd.AppDirectory); err != nil {
		log.Error("failed to find app directory: %s", err)
		return &ProjectValidatorResult{Result:Result{Error:false}}
	}

	if !stat.IsDir() {
		log.Error("found %s but it is not a directory!", cmd.AppDirectory)
		return &ProjectValidatorResult{Result:Result{Error:true}}
	}

	return &ProjectValidatorResult{Result:Result{Error:false}};
}
