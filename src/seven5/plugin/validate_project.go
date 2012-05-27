package plugin

import (
	"os"
	"seven5/util"
	"path/filepath"
)

//ValidateProjectArgs is passed to the ProjectValidator to do its
//work.Must be public for json encoding.
type ValidateProjectArgs struct {
}

//ProjectValidatorResult is the result type of a call on the ProjectValidator.
//Must be public for json encoding.
type ValidateProjectResult struct {
	CommandResult
}

// ProjectValidator checks to see if the layout of the project is
// acceptable for future phases.
type ValidateProject interface {
	Validate(cmd *Command, args *ValidateProjectArgs, 
		log util.SimpleLogger) *ValidateProjectResult
}

// Default project validator looks for the directory structure
// app
type DefaultValidateProjectImpl struct {
}

//verifyDirectory is used to make sure that a particular project has
//a filesystem entry with this name.  true is used to check that it is
//a directory, otherwise checks for file.
func (self *DefaultValidateProjectImpl) verifyFSEntry(log util.SimpleLogger,
	isDirectory bool, path string, candidate... string) bool {
	
	var err error
	var stat os.FileInfo
	
	proposed := filepath.Join(path,filepath.Join(candidate...))
	if stat, err = os.Stat(proposed); err != nil {
		log.Error("failed to find fs entry: %s", err)
		return false
	}
	
	if isDirectory {
		return stat.IsDir()
	}
	return !stat.IsDir()

} 

func (self *DefaultValidateProjectImpl) Validate(cmd *Command, args *ValidateProjectArgs, 
log util.SimpleLogger) *ValidateProjectResult {

	log.Debug("Using DefaultProjectValidator in %s", cmd.AppDirectory)
	names := []string{"client","public","src","app.json"}
	dir := []bool{true, true ,true,false}
	for i, n := range(names) {
		if !self.verifyFSEntry(log, dir[i], cmd.AppDirectory, n) {
			log.Error("failed to find %s/%s: invalid project",cmd.AppDirectory,n)
			return &ValidateProjectResult{ErrorResult()}
		}
	}

	//everything is ok so we return no error
	return &ValidateProjectResult{};
}
