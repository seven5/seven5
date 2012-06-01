package seven5

import (
	"net/http"
	"seven5/util"
)

//these are the only types that we understand and that we will take action
//with in a model, timestamp is a number of nanos since epoch (time.time)
const (
   SEVEN5_STRING = iota
   SEVEN5_INT64 
   SEVEN5_BOOL 
   SEVEN5_FLOAT64
   SEVEN5_TIMESTAMP //UNIX time (int64 since jan 1, 1970) and always UTC
)

//FieldInfo gives a description of the fields in a structure that we
//understand.  Note that the structure may have many fields we do not
//understand.
type FieldInfo struct {
	Name string
	Seven5Type string
}

//VocabInfo is a struct that tells you about each vocab.
type VocabInfo struct {
	Name string
	Value map[string]int
}

// Return type of the explode type object
type ExplodeTypeResult struct {
	Error bool
	Vocab []*VocabInfo
}

// DefaultExplodeType dumps out all the info it can find about types that
// may be interesting to the larger application.
type DefaultExplodeType struct {
}


func (self *DefaultExplodeType) Exec(command string, dir string,
	config *ApplicationConfig, request *http.Request, 
	log util.SimpleLogger) interface{} { 

	return nil	
}