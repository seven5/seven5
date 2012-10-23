package seven5

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
)

var t *template.Template

//init creates the template needed for Dart code generation.
func init() {
	fnMap := template.FuncMap{
		"tolower": strings.ToLower,
	}
	t = template.Must(template.New("DART_CLASS_TMPL").Funcs(fnMap).Parse(DART_CLASS_TMPL))
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

func generateDartForResource(r *ResourceDescription) string {
	var buffer bytes.Buffer
	if err := t.ExecuteTemplate(&buffer, "DART_CLASS_TMPL", r); err != nil {
		return err.Error()
	}
	return buffer.String()
}

func generateDartForSupportStruct(f *FieldDescription) string {
	var buffer bytes.Buffer
	if err := t.ExecuteTemplate(&buffer, "SUPPORT_STRUCT_TMPL", f); err != nil {
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

const DART_CLASS_TMPL = `
{{define "FIELD_DECL"}}
	{{range .Struct}} 
		{{.Dart}} {{.Name}};
	{{end}}	
{{end}}

{{define "COPY_JSON_FIELDS"}}
	{{range .Struct}}

		{{if .StructName}}
			this.{{.Name}} = new {{.StructName}}.fromJson(json["{{.Name}}"]);
		{{else}}
			this.{{.Name}} = json["{{.Name}}"];
			{{end}}{{/* if */}}
		{{end}} {{/* range */}}
{{end}} {{/* define */}}

class {{.Name}} {
	{{template "FIELD_DECL" .Field}}

	static String resourceURL = "/{{tolower .Name}}/";

	static List<{{.Name}}> Index(successFunc, [errorFunc, headers, requestParameters]) {
		Seven5Support.Index(resourceURL, ()=>new List<{{.Name}}>(), ()=>new {{.Name}}(), successFunc, errorFunc, headers, requestParameters);
	}

	void Find(int Id, successFunc, [errorFunc, headers, requestParameters]) {
		Seven5Support.Find(Id, resourceURL, new {{.Name}}(), successFunc, errorFunc, headers, requestParameters);
	}
	
	
	//convenience constructor
	{{.Name}}.fromJson(Map json) {
		copyFromJson(json);
	}
	
	//nothing to do in default constructor
	{{.Name}}();
	
	//this is the "magic" that changes from untyped Json to typed object
	copyFromJson(Map json) {
		{{template "COPY_JSON_FIELDS" .Field}}
		return this;
	}
}

{{define "SUPPORT_STRUCT_TMPL"}}
	class {{.StructName}} {
		{{template "FIELD_DECL" .}}

		//convenience constructor
		{{.StructName}}.fromJson(Map json) {
			copyFromJson(json);
		}
	
		//nothing to do in default constructor
		{{.StructName}}();
	
		//this is the "magic" that changes from untyped Json to typed object
		copyFromJson(Map json) {
			{{template "COPY_JSON_FIELDS" .}}
			return this;
		}
	}
{{end}} {{/*define*/}}
`
