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



	t = template.Must(template.New("CLASSDECL_TMPL").Funcs(fnMap).Parse(classdecl_tmpl))
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

func generateDartForResource(r *APIDoc) string {



	var buffer bytes.Buffer



	if err := t.ExecuteTemplate(&buffer, "CLASSDECL_TMPL", r); err != nil {



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
