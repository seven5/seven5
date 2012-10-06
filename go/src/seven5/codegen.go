package seven5

import (
	"text/template"
	"bytes"
)

var t *template.Template

func init() {
	t = template.Must(template.New("DART_CLASS_TMPL").Parse(DART_CLASS_TMPL))
}

func (self *ResourceDescription) generateDartDecl() string {
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
}
`
