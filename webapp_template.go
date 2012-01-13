package seven5


const WEBAPP_TEMPLATE = `//target:{{.package}}
package main

import (
	{{importIfNeeded .import}}
	"seven5"
	"fmt"
	"os"
)

//Because you can't dynamically load go code yet, you have to use this
//bit of boilerplate. 
func main() {

	seven5.BackboneModel(&seven5.User{})

	{{range $key,$val := .model}}
    seven5.BackboneModel("{{$key}}",{{range .}}"{{.}}",{{end}})
	{{end}}


	//derive from filenames
	{{range .handler}}
	{{.}} := {{$pkg}}.New{{upper .}}()
	{{end}}

	var config *seven5.ProjectConfig
	var err error

	// accept all defaults for project layout, etc... note parameters are from filenames
	if config, err = seven5.WebAppDefaultConfig({{range .handler}}{{.}},{{end}}); err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		return
	}
	//run handlers and guises
	if err = seven5.WebAppRun(config, {{range .handler}}{{.}},{{end}}); err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		return
	}
}`
