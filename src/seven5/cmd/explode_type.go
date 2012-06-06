package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"seven5/util"
	"strings"
	"reflect"
	"os"
)


//FieldInfo gives a description of the fields in a structure that we
//understand.  Note that the structure may have many fields we do not
//understand.
type FieldInfo struct {
	Name       string
	TypeName   string
}

//VocabInfo is a struct that tells you about each vocab.
type VocabInfo struct {
	Name  string
	Field []*FieldInfo
}

type PillVocabWrapper struct {
	Error bool
	ErrorMsg string
	Vocab []*VocabInfo
}

// Return type of the explode type object
type ExplodeTypeResult struct {
	Error bool
	Vocab []*VocabInfo
}

//
// ExplodeTypes is used to send back determine key information about types in
// the client application. It is also interesting if you want to implement
// a command that uses a pill or has custom argument/return value marshalling.
// Public because it is referenced by the Seven5 pill.
var ExplodeType = &CommandDecl{
	Arg: []*CommandArgPair{
		ProjectConfiguration, // project config
		ParamFromFiles("vocab",true), //list of the vocabulary files
	},
	Ret: ExplodeTypeReturn,
	Impl: defaultExplodeType,
}


//ExplodeTypeReturn is the necessary cruft to tell the unmarshalling code on
//the client side how to handle our return value.  We don't return a body
//so the last field is nil.  Marshalling code is not in our package so this
//must be public.
var ExplodeTypeReturn =  &CommandReturn {
	func() interface{} { return &ExplodeTypeResult{}},
	func(v interface{}) bool { return v.(*ExplodeTypeResult).Error },
	nil,
}


//ProbeVocabAll is the driver routine for the pill. Input is all the named
//vocabs and it returns the json output for this command after repeatedly
//calling ProbeVocab.  Referenced by the pill, so must be public.
func ProbeVocabAll(vocabs...interface{}) string {
	result := &PillVocabWrapper{}
	
	for _, v := range vocabs {
		info, err := probeVocab(v)
		if err!="" {
			return err
		}
		result.Vocab = append(result.Vocab, info)
	}
	return formatProbeVocabResult(result)
}

//probeVocab runs _inside_ the pill that determines the type information about
//a user library.  It returns json that represents the result of it's work,
//even if that is an error message.
func probeVocab(candidate interface{}) (*VocabInfo, string) {
	
	t:=reflect.TypeOf(candidate)
	if t.Kind()!=reflect.Struct {
		return nil, sendErrorMessageAsPillResult("type %s is a %s, not a struct! vocab definitations "+
			"must be a struct!", t.Name(), t.Kind());
	}

	result := &VocabInfo{Name: t.Name()}
	
	hasId := false
	
	for i:=0; i<t.NumField(); i++ {
		field := t.Field(i)
		if ((field.Name == "Id") && (field.Type.Kind()==reflect.Int64)) {
			hasId=true;
			continue;
		}
		vocabInfo, errorResult := fieldToFieldInfo(field)
		if errorResult!="" {
			return nil, errorResult
		}
		result.Field = append(result.Field, vocabInfo)
	}
	
	if !hasId {
		return nil, sendErrorMessageAsPillResult("all structs that represent a dictionary "+
			"must have an 'Id int64' field: %s does not", t.Name())
	}
	return result, ""
}

func fieldToFieldInfo(field reflect.StructField) (*FieldInfo, string){

	if !knownTypes(field.Type) {
		return nil, sendErrorMessageAsPillResult("seven5 doesn't know this type yet: %s", 
			field.Type.Kind().String())
	}
	result := &FieldInfo{Name: field.Name, TypeName: field.Type.Kind().String()}
	return result, ""
}

func knownTypes(t reflect.Type) bool {
	switch t.Kind() {
	case reflect.String:
		return true
	case reflect.Int64:
		return true
	}
	return false
}

//formErrorMessageAsPillResult computes a json bundle that expresses the 
//error message passed (as to Sprintf) for the type exploder pill
func sendErrorMessageAsPillResult(fmtString string, v ...interface{}) string {
	result := &PillVocabWrapper{}
	result.Error=true
	result.ErrorMsg=fmt.Sprintf(fmtString,v...)
	return formatProbeVocabResult(result)
}

//formatProbeVocabResult returns either an error string or it exist the 
//program with an exit code, which is ok. It runs inside the pill.
func formatProbeVocabResult(result *PillVocabWrapper) string {
	var buffer bytes.Buffer
	encoder:=json.NewEncoder(&buffer)
	error:=encoder.Encode(result)
	if error!=nil {
		fmt.Printf("Failed to encode json result: %s", error);
		os.Exit(1);
	}
	return buffer.String()
}

func defaultExplodeType(log util.SimpleLogger, v...interface{}) interface{} { 
	config := v[0].(*ProjectConfig)
	vocab := v[1].([]string)
	
	//construct a function call that has all the names of the vocabs
	var expectedList bytes.Buffer
	var probeBuffer bytes.Buffer
	probeBuffer.WriteString("\tseven5.ProbeVocabAll(")
	for _, v := range vocab {
		if expectedList.Len()>0 {
			expectedList.WriteString(", ");
		}
		expectedList.WriteString(v)
		call := fmt.Sprintf("%s.%s{},",config.AppName, v)
		probeBuffer.WriteString(call)
	}
	probeBuffer.WriteString(")")
	
	//construct the pill
	pill:=fmt.Sprintf(EXPLODE_TYPE_PILL, config.AppName, probeBuffer.String())
	p, compileMessage,err :=util.CompilePill(pill,log)
	if err!=nil {
		log.DumpTerminal(util.DEBUG, "Failed To Compile Type Exploder Pill",
			err.Error());
	}
	if compileMessage!="" {
		log.DumpTerminal(util.DEBUG, "Compiler Error Building Type Exploder Pill",
			compileMessage);
	}
	if err!=nil || compileMessage!="" {
		log.Error("Failed to understand vocabulary type: expecting to see these types: %s",
			expectedList.String())
		return &ExplodeTypeResult{Error:true}
	}
	//we got a pill, lets run it
	cmd:=exec.Command(p)
	out, err:= cmd.CombinedOutput()
	if err!=nil {
		log.DumpTerminal(util.ERROR, "Failed To Run Type Exploder Pill",
			err.Error());
		return &ExplodeTypeResult{Error:true}
	}
	output:=string(out)
	log.DumpJson(util.DEBUG, "Type Exploder Pill Output", output)
	dec:= json.NewDecoder(strings.NewReader(output))
	wrapper:=&PillVocabWrapper{}
	err=dec.Decode(wrapper)
	if err!=nil {
		log.Error("Unable to decode into VocabWrapper: %s", err)
		return &ExplodeTypeResult{Error:true}
	}
	if wrapper.Error {
		log.Error(wrapper.ErrorMsg)
		return &ExplodeTypeResult{Error:true}
	}
	
	//everything went ok
	result:=&ExplodeTypeResult{Error:false}
	result.Vocab=wrapper.Vocab
	return result
}

const EXPLODE_TYPE_PILL = `
package main
import "seven5"
import "fmt"
import "os"

import "%s"

func main() {
	fmt.Println(%s)
	os.Stdout.Sync()
}
`
