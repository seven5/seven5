package cmd

import (
	"os"
	"path/filepath"
	"seven5/util"
	"strings"
)

//
// Destroy generated files is used to clean up before generating a new set of
// generated files.  This is useful in case the names of the "key" files change,
// we don't want to leave files dangling.  Public because it is referenced by
// the Seven5 pill.
//
var DestroyGeneratedFiles = &CommandDecl{
	Arg: []*CommandArgPair{
		ProjectSrcDir, //project source code dir
	},
	Ret: SimpleReturn,
	Impl: defaultDestroyGeneratedFiles,
}


func defaultDestroyGeneratedFiles(log util.SimpleLogger, v...interface{}) interface{} {

	appPath := v[0].(string)
	
	log.Info("Destroying Seven5 generated files.")
	f, err := os.Open(appPath)
	if err != nil {
		log.Error("Unable to read directory contents: %s", appPath)
		return &SimpleErrorReturn{Error: true}
	}
	name, err := f.Readdir(-1)
	if err != nil {
		log.Error("Error reading directory contents: %s", appPath)
		return &SimpleErrorReturn{Error: true}
	}
	for _, n := range name {
		if strings.HasSuffix(n.Name(), "_generated.go") {
			die := filepath.Join(appPath, n.Name())
			err = os.Remove(die)
			if err != nil {
				log.Error("Error deleting generated code: %s", die)
				return &SimpleErrorReturn{Error: true}
			}
		}
	}
		
	return &SimpleErrorReturn{Error:false}
}
