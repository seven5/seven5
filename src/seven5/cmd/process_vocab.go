package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"encoding/json"
	"seven5/util"
)

//
// Process a vocab is responsible for taking a collection of types that have
// been verified to be vocabularies and generating the necessary code to allow
// them to be used as such in a user project.  Public because it is referenced
// from the Seven5 pill.
//
var ProcessVocab = &CommandDecl{
	Arg: []*CommandArgPair{
		ProjectConfiguration, // project config file contents
		ProjectSrcDir, // project config file contents
		ParamFromFiles("_vocab",false),
	},
	Ret: SimpleReturn,
	Impl: defaultProcessVocab,
}


func defaultProcessVocab(log util.SimpleLogger, v...interface{}) interface{} {
	config:=v[0].(*ProjectConfig)
	srcDir:=v[1].(string)
	arg := v[2].([]string)

	for _, vocab := range arg {
		typeName := util.FilenameToTypeName(vocab)
		log.Debug("filename conversion %s", typeName)
		p :=  filepath.Join(srcDir, typeName+"_generated.go")
		log.Debug("path %s", p)
		file, err := os.Create(p)
		if err != nil {
			return &SimpleErrorReturn{Error:true}
		}
		err = generate(file, srcDir, typeName, config)
		if err != nil {
			log.Error("Unable to generate code in %s:%s", p, err.Error())
			return &SimpleErrorReturn{Error:true}
		}
		err = file.Close()
		if err != nil {
			log.Error("Unable to close file %s:%s", p, err.Error())
			return &SimpleErrorReturn{Error:true}
		}
	}

	return &SimpleErrorReturn{Error:false}

}

func generate(file *os.File, dirName string, vocabName string, 
	config *ProjectConfig) error {

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
