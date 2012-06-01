package seven5

import (
	"net/http"
	"seven5/util"
)


//ProcessVocabResult is the return value after we execute this seven5
//command.
type ProcessVocabResult struct {
	Error bool
}

// DefaultProcessVocab looks for vocabulary definitions in the source,
// ending in _vocab.go, and builds a package level initalizer that will
// initalize that vocabulary.  It needs to be supplied with the names
// of the files and if they have itializers.
type DefaultProcessVocab struct {
}

func (self *DefaultProcessVocab) Exec(command string, dir string,
	config *ApplicationConfig, request *http.Request, 
	log util.SimpleLogger) interface{} { 
	
	return nil
}
