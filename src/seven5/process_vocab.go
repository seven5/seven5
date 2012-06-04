package seven5

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"encoding/json"
	"seven5/util"
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

	arg := raw.(*ProcessVocabArg)
	result := &ProcessVocabResult{}

	for _, vocab := range arg.Info {
		filename := util.TypeNameToFilename(vocab.Name)
		log.Debug("filename conversion %s", filename)
		srcDir := filepath.Join(dir, "src", config.AppName)
		p :=  filepath.Join(srcDir, filename+"_generated.go")
		log.Debug("path %s", p)
		file, err := os.Create(p)
		if err != nil {
			log.Error("Unable to write file %s:%s", p, err.Error())
			result.Error = true
			return result
		}
		err = generate(file, srcDir, vocab.Name, config)
		if err != nil {
			log.Error("Unable to generate code in %s:%s", p, err.Error())
			result.Error = true
			return result
		}
		err = file.Close()
		if err != nil {
			log.Error("Unable to close file %s:%s", p, err.Error())
			result.Error = true
			return result
		}
	}

	return result
}

func generate(file *os.File, dirName string, vocabName string, 
	config *ApplicationConfig) error {

	load := fmt.Sprintf(LOAD_DATA_CODE,config.AppName, vocabName, vocabName, 
		filepath.Join(dirName,util.TypeNameToFilename(vocabName)+".json"), 
		vocabName, vocabName) 
	_,err:=file.WriteString(load);
	if err!=nil {
		return err
	}
	
	return nil
}
//LoadVocab is the run-time support for loading json from a example set into
//a particular vocab.
func LoadAndSaveVocab(path string, v interface{}, saveFn func (interface{})error) error {
	file, err := os.Open(path)
	if err!=nil {
		return err
	}
	decoder := json.NewDecoder(file)
	for {
		err=decoder.Decode(v)
		if err==io.EOF {
			break;
		}
		if err!=nil {
			return err
		}
		
		//read it ok, now process it
		err= saveFn(v)
		if err!=nil {
			return err
		}
	}
	
	return file.Close()
}

func SaveEntry(vocabName string, v interface{}) error {
	return nil
}

const LOAD_DATA_CODE = `
package %s

import "seven5"

func save%s(obj interface{}) error {
	err := seven5.SaveEntry("%s",obj)
	return err
}

func init()  {
	var fname = "%s.json"
	err := seven5.LoadVocab(fname,&%s{},save%s)  
	if err!=nil {
		panic("error loading vocab data "+fname+" at applicatin startup!")
	}
	return nil
}
`