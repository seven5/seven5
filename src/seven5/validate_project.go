package seven5

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
	Error bool
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
	ignored3 *http.Request, ignored4 interface{}, log util.SimpleLogger) interface{} {

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
			return ValidateProjectResult{Error: true}
		}
	}

	//ok, top level passed ok, let's ready in the app.json
	cfg, err := decodeAppConfig(dir)
	if err!=nil {
		log.Error("Error reading app configuration file %s: %s",
			dir,err.Error())
		return ValidateProjectResult{Error:true}
	}
	
	//check the parent dir, src subdir, and .go entry point based on config
	parent:=filepath.Dir(dir)
	if !self.verifyFSEntry(log, true, parent, cfg.AppName) {
		log.Error("cant find expected app root directory %s/%s",parent,cfg.AppName) 
		return ValidateProjectResult{Error:true}
	}

	if filepath.Base(dir)!=cfg.AppName {
		log.Error("root directory is %s but expected %s", filepath.Base(dir),
			cfg.AppName);
		return ValidateProjectResult{Error: true}
	}
	
	src:=filepath.Join(dir,"src")
	if !self.verifyFSEntry(log, true, src, cfg.AppName) {
		log.Error("to build properly with go tools, src should have subdirectory %s",
			cfg.AppName)
		return ValidateProjectResult{Error: true}
	}
	
	codeDir := filepath.Join(src,cfg.AppName)
	goFile := cfg.AppName + ".go"
	if !self.verifyFSEntry(log, false, codeDir, goFile) {
		log.Error("can't find app main entry point, expected it to be %s",
			filepath.Join(codeDir,goFile))
		return ValidateProjectResult{Error:true}
	}	
	
	//everything is ok so we return no error
	return ValidateProjectResult{Error:false}
}

func (self *DefaultValidateProject) GetArg() interface{} {
	return nil
}

