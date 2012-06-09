package cmd

import (
	"seven5/util"
	"os"
	"path/filepath"
	"strings"
)

//Validate project checks out the structure of a project. It is public because
//it is referenced from the Seven5 pill.
var ValidateProject = &CommandDecl{
	Arg: []*CommandArgPair{
		ProjectRootDir, //root of the user project
		ProjectConfiguration, //contents of project.json
	},
	Ret: SimpleReturn,
	Impl: defaultValidateProject,
}


func defaultValidateProject(log util.SimpleLogger, v...interface{}) interface{} {
	dir := v[0].(string)
	cfg := v[1].(*ProjectConfig)
	
	dirForHuman := dir
	parts := strings.SplitAfter(dir, string(filepath.Separator))
	if len(parts) > 3 {
		parts = parts[len(parts)-3:]
		dirForHuman = filepath.Join(parts...)
	}
	log.Debug("Using DefaultProjectValidator in %s", dirForHuman)
	names := []string{"client", "public", "src", "project.json"}
	directory := []bool{true, true, true, false}
	for i, n := range names {
		if !verifyFSEntry(log, directory[i], dir, n) {
			log.Error("failed to find %s/%s: invalid project", dir, n)
			return &SimpleErrorReturn{Error:true}
		}
	}

	//check the parent dir, src subdir, and .go entry point based on config
	parent:=filepath.Dir(dir)
	if !verifyFSEntry(log, true, parent, cfg.AppName) {
		log.Error("cant find expected app root directory %s/%s",parent,cfg.AppName) 
		return &SimpleErrorReturn{Error:true}
	}

	if filepath.Base(dir)!=cfg.AppName {
		log.Error("root directory is %s but expected %s", filepath.Base(dir),
			cfg.AppName);
		return &SimpleErrorReturn{Error:true}
	}
	
	src:=filepath.Join(dir,"src")
	if !verifyFSEntry(log, true, src, cfg.AppName) {
		log.Error("to build properly with go tools, src should have subdirectory %s",
			cfg.AppName)
		return &SimpleErrorReturn{Error:true}
	}
	
	codeDir := filepath.Join(src,cfg.AppName)
	goFile := cfg.AppName + ".go"
	if !verifyFSEntry(log, false, codeDir, goFile) {
		log.Error("can't find app main entry point, expected it to be %s",
			filepath.Join(codeDir,goFile))
		return &SimpleErrorReturn{Error:true}
	}	
	
	//everything is ok so we return no error
	return &SimpleErrorReturn{Error:false}
	
}



//verifyDirectory is used to make sure that a particular project has
//a filesystem entry with this name.  true is used to check that it is
//a directory, otherwise checks for file.
func verifyFSEntry(log util.SimpleLogger, isDirectory bool, path string, 
	candidate ...string) bool {

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