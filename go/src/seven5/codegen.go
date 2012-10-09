package seven5

import (
	"text/template"
	"bytes"
)

var t *template.Template

func init() {
	t = template.Must(template.New("DART_CLASS_TMPL").Parse(DART_CLASS_TMPL))
}

func (self *ResourceDescription) generateDart() string {
	var buffer bytes.Buffer
	if err:=t.ExecuteTemplate(&buffer, "DART_CLASS_TMPL", self); err!=nil {
		return err.Error()
	}
	return buffer.String()
}

const DART_CLASS_TMPL=`
class {{.Name}} {
	{{range .Fields}} {{.DartType}} {{.Name}};
	{{end}}
	
	static String findURL = "{{.GETSingular}}";
	static String indexURL = "{{.GETPlural}}";

	static List<{{.Name}}> Index(successFunc, [errorFunc, headers, requestParameters]) {
		Seven5Support.Index(indexURL, ()=>new List<{{.Name}}>(), 
			()=>new {{.Name}}(), successFunc, errorFunc, headers, requestParameters);
	}

	void Find(int Id, successFunc, [errorFunc, headers, requestParameters]) {
		Seven5Support.Find(Id, findURL, new ItalianCity(), successFunc, errorFunc, headers,
			requestParameters);
	}
	
	//convenience constructor
	{{.Name}}.fromJson(Map json) {
		copyFromJson(json);
	}
	
	//nothing to do in default constructor
	{{.Name}}();
	
	//this is the "magic" that changes from untyped Json to typed object
	{{.Name}}.copyFromJson(Map json) {
		{{range .Fields}} this.{{.Name}} = json["{{.Name}}"]
		{{end}}
		return this;
	}
}
`
