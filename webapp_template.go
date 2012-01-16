package seven5


//WEBAPP_TEMPLATE is the go code necessary to create a main.main() that glues the user-level package
//that has the webapp to the seven5 package.  It is used by the tune tool, so must be public.
const WEBAPP_TEMPLATE = `
package main

import (
	{{importIfNeeded .import}}
	"seven5"
)

//Because you can't dynamically load go code yet, you have to use this
//bit of boilerplate. 
func main() {

    seven5.BackboneService("user", seven5.NewUserSvc(), &seven5.User{})

	{{$pkg = .import}}
	{{range .model}}
    seven5.BackboneService("{{lower .}}",{{$pkg}}.New{{.}}Svc(), &{{$pkg}}.{{.}}{})
	{{end}}
	
	{{$pkg = .package}}
	//derive from filenames
	{{range .handler}}
	{{.}} := {{$pkg}}.New{{upper .}}()
	{{end}}

	seven5.WebAppRun()
}`
