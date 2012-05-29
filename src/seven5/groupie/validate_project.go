package groupie

import (
	"os"
	"path/filepath"
	"seven5/util"
	"strings"
	"net/http"
)

//ProjectValidatorResult is the result type of a call on the ProjectValidator.
//Must be public for json encoding.
type ValidateProjectResult struct {
	CommandResult
}

// Default project validator looks for the directory structure
// app
type DefaultValidateProject struct {
}

//verifyDirectory is used to make sure that a particular project has
//a filesystem entry with this name.  true is used to check that it is
//a directory, otherwise checks for file.
func (self *DefaultValidateProject) verifyFSEntry(log util.SimpleLogger,
	isDirectory bool, path string, candidate ...string) bool {

	var err error
	var stat os.FileInfo

	proposed := filepath.Join(path, filepath.Join(candidate...))
	if stat, err = os.Stat(proposed); err != nil {
		return false
	}

	if isDirectory {
		return stat.IsDir()
	}
	return !stat.IsDir()

}
func (self *DefaultValidateProject) Exec(ignored1 string, 
	dir string, ignored2 *ApplicationConfig,
	ignored3 *http.Request, log util.SimpleLogger) interface{} {

	dirForHuman := dir
	parts := strings.SplitAfter(dir, string(filepath.Separator))
	if len(parts) > 3 {
		parts = parts[len(parts)-3:]
		dirForHuman = filepath.Join(parts...)
	}
	log.Debug("Using DefaultProjectValidator in %s", dirForHuman)
	names := []string{"client", "public", "src", "app.json"}
	directory := []bool{true, true, true, false}
	for i, n := range names {
		if !self.verifyFSEntry(log, directory[i], dir, n) {
			log.Error("failed to find %s/%s: invalid project", dir, n)
			return &ValidateProjectResult{ErrorResult()}
		}
	}

	//everything is ok so we return no error
	return &ValidateProjectResult{}
}
