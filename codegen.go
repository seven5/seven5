package seven5

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
	"os"
	"net/http"
)

var codegenTemplate *template.Template

//init creates the template needed for Dart code generation.
func init() {
	fnMap := template.FuncMap{
		"tolower": strings.ToLower,
	}
	codegenTemplate = template.Must(template.New("CLASSDECL_TMPL").Funcs(fnMap).Parse(classdecl_tmpl))
}

//contains type checks a slice of FieldDescriptions to see if a candidate type is
//present in the slice. Comparison is based on StructName.
func containsType(all []*FieldDescription, candidate *FieldDescription) bool {
	for i := 0; i < len(all); i++ {
		if all[i].StructName == candidate.StructName {
			return true
		}
	}
	return false
}

//collectStructs is a recursive walk of FieldDescription structure to find all 
//types that are known as structs.  This full list is needed to generate code.
func collectStructs(current *FieldDescription) []*FieldDescription {
	result := []*FieldDescription{}
	if current.StructName != "" {
		result = append(result, current)
		for _, c := range current.Struct {
			for _, f := range collectStructs(c) {
				if !containsType(result, f) {
					result = append(result, f)
				}
			}
		}
	}
	return result
}

//fdWrapper is a wrapper and field description that adds a field to help the code generator know what
//the rest prefix is.
type fdWrapper struct {
	*FieldDescription
	RestPrefix string
}
func generateDartForResource(f *FieldDescription, prefix string) string {
	var buffer bytes.Buffer
	w:=&fdWrapper{f,prefix+"/"}
	if err := codegenTemplate.ExecuteTemplate(&buffer, "CLASSDECL_TMPL", w); err != nil {
		return err.Error()
	}
	return buffer.String()
}

func generateDartForSupportStruct(f *FieldDescription) string {
	var buffer bytes.Buffer
	if err := codegenTemplate.ExecuteTemplate(&buffer, "SUPPORT_STRUCT_TMPL", f); err != nil {
		return err.Error()
	}
	return buffer.String()
}

//Dart returns the Dart name for a particular _type_ name or panics if it does not understand.
func (self *FieldDescription) Dart() string {
	switch self.TypeName {
	case "Boolean":
		return "bool"
	case "Integer":
		return "int"
	case "Floating":
		return "double"
	case "String255":
		return "String"
	case "Textblob":
		return "String"
	case "Id":
		return "int"
	}
	if self.Array != nil {
		return "List<" + self.Array.Dart() + ">"
	}
	if self.StructName != "" {
		return self.StructName
	}
	panic(fmt.Sprintf("unable to convert type %s to Dart type!", self.TypeName))
}

//HasId returns true if this struct has a field Id of type seven5.Id.
func (self *FieldDescription) HasId() bool {
	if len(self.Struct) == 0 {
		return false
	}
	ok := false
	for i := 0; i < len(self.Struct); i++ {
		if self.Struct[i].Name == "Id" && self.Struct[i].TypeName == "Id" {
			ok = true
			break
		}
	}
	return ok
}
//dartPrettyPrint is a very naive dart formatter. It doesn't understand much of the lexical
//structure of dart but it's enough for our generated code (which doesn't do things like embed
//{ inside a string and does has too many, not too few, line breaks)
func dartPrettyPrint(raw string) []byte {
	state := WAITING_ON_NON_WS
	indent := 0
	var result bytes.Buffer
	for i := 0; i < len(raw); i++ {
		c := raw[i]
		switch state {
		case WAITING_ON_NON_WS:
			if c == '\t' || c == ' ' || c == '\n' {
				continue
			}
			switch c {
			case '{':
				indent += 2
			case '}':
				indent -= 2
			}
			for j := 0; j < indent; j++ {
				result.WriteString(" ")
			}
			result.Write([]byte{c})
			state = WAITING_ON_EOL
			continue
		case WAITING_ON_EOL:
			if c == '\n' {
				result.WriteString("\n")
				state = WAITING_ON_NON_WS
				continue
			}
			switch c {
			case '{':
				indent += 2
			case '}':
				indent -= 2
			}
			result.Write([]byte{c})
			continue
		}
	}
	return result.Bytes()
}



//GeneratedDartContent adds an http handler for a particular path.  The restPrefix must be the same one
//used by the TypeHolder (probably a dispatcher) to map its rest resources.
//func GeneratedDartContent(mux *ServeMux, holder TypeHolder, urlPath string, restPrefix string) {
//	mux.HandleFunc(fmt.Sprintf("%sdart", urlPath), generateDartFunc(holder, restPrefix))
//}


const LIBRARY_INFO = `
library generated;
import 'package:seven5/support';
import 'dart:json';
`


//generateDartFunc returns a function that outputs text string for all the wire types associated
//with all the resources known to the type holder.  Note that the number of classes output may be
//different than the number of resources because the wire types may nest structures inside.
func generateDartFunc(holder TypeHolder, prefix string) func(http.ResponseWriter,*http.Request){
	return func(w http.ResponseWriter, r *http.Request) {
		text:=wrappedCodeGen(holder, prefix)
		if _,err:=w.Write(text.Bytes()); err!=nil {
			fmt.Fprintf(os.Stderr, "Unable to write result of code generation to the client: %s\n", err)
		}
	}
} 

//wrappedCodeGenFunc is the top level of the code generation.
func wrappedCodeGen(holder TypeHolder,prefix string) bytes.Buffer{
	var text bytes.Buffer
	resourceStructs := []*FieldDescription{}
	supportStructs := []*FieldDescription{}
	text.WriteString(LIBRARY_INFO)
	for _, d := range holder.All() {
		text.WriteString(generateDartForResource(d, prefix))
		resourceStructs = append(resourceStructs, d)
	}
	for _, d := range holder.All() {
		candidates := collectStructs(d)
		for _, s := range candidates {
			if !containsType(resourceStructs, s) && !containsType(supportStructs, s) {
				supportStructs = append(supportStructs, s)
			}
		}
	}
	for _, i := range supportStructs {
		text.WriteString(generateDartForSupportStruct(i))
	}
	return text	
}
