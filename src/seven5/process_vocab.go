package seven5

import (
	"net/http"
	"seven5/util"
	"os"
	"path/filepath"
)


//ProcessVocabResult is the return value after we execute this seven5
//command.
type ProcessVocabResult struct {
	Error bool
}
//ProcessVocabArg is the argument that gets called to tell us what code
//to generate.
type ProcessVocabArg struct {
	Info []*VocabInfo
}

// DefaultProcessVocab looks for vocabulary definitions in the source,
// ending in _vocab.go, and builds a package level initalizer that will
// initalize that vocabulary.  It needs to be supplied with the names
// of the files and if they have itializers.
type DefaultProcessVocab struct {
}


//GetArg returns  an instance of our argument type
func (self *DefaultProcessVocab) GetArg() interface{} {
	return &ProcessVocabArg{}
}

func (self *DefaultProcessVocab) Exec(command string, dir string,
	config *ApplicationConfig, request *http.Request, raw interface{},
	log util.SimpleLogger) interface{} { 

	arg:=raw.(*ProcessVocabArg)
	result:=&ProcessVocabResult{}
	
	for _,vocab := range arg.Info {
		filename := util.TypeNameToFilename(vocab.Name)
		log.Debug("filename conversion %s",filename)
		p:=filepath.Join(dir, "src", config.AppName, filename+"_generated.go");
		log.Debug("path %s",p)
		file, err := os.Create(p)
		if err!=nil {
			log.Error("Unable to write file %s:%s",p,err.Error())
			result.Error=true
			return result
		}
		err = file.Close()
		if err!=nil {
			log.Error("Unable to close file %s:%s",p,err.Error())
			result.Error=true
			return result
		}
	}
	
	return result
}
