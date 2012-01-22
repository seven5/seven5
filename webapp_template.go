package seven5


//WEBAPP_TEMPLATE is the go code necessary to create a main.main() that glues the user-level package
//that has the webapp to the seven5 package.  It is used by the tune tool, so must be public.
const WEBAPP_TEMPLATE = `
package main

import (
	{{importIfNeeded .import}}
	"seven5"
	"seven5/store"
)

//Because you can't dynamically load go code yet, you have to use this
//bit of boilerplate. 
func main() {

    seven5.BackboneService("user", seven5.NewUserSvc(), &seven5.User{})

	cfg:=&seven5.WebAppConfig{
		AllowShutdown: {{.config.AllowShutdown}},
		PrivateInit: {{.config.PrivateInit}},
		Store: {{.config.Store}},
    }

	{{$pkg = .import}}
	{{range .model}}
    seven5.BackboneService("{{lower .}}",{{$pkg}}.New{{.}}Svc(), &{{$pkg}}.{{.}}{})
	{{end}}


	{{/* this section is for handling http and json level handlers... needs to be "recovered" */}}
	{{$pkg = .package}}
	{{range .handler}}
	{{.}} := {{$pkg}}.New{{upper .}}()
	{{end}}
	{{/* ************************************************************************************* */}}

	seven5.WebAppRun(cfg)
}`
