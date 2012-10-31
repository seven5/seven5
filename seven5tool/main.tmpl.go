package seven5tool
const main_tmpl=`
package main

import (
	"github.com/seven5/seven5"
	"{{.base}}"
	"net/http"
	"fmt"
)

func main() {

	h := seven5.NewSimpleHandler()
	h.AddResource({{.base}}.{{.name}}{}, &{{.base}}.{{.name}}Resource{})

	//the result object is the _same object_ as h, but just using a different type to make
	//it more "clean" when used with the built in http package.
	asHttp := seven5.DefaultProjectBindings(h, "{{.base}}")

	//normal http calls for running a server in go... ListenAndServe never should return
	//err:=http.ListenAndServe(":3003",logHTTP(asHttp))

	//use this verson, not the one above, if you want to log HTTP requests to the terminal
	err := http.ListenAndServe(":3003", logHTTP(asHttp))

	fmt.Printf("Error! Returned from ListenAndServe(): %s", err)
}

// tiny wrapper around all the HTTP dispatching that can be nice to help with debugging
func logHTTP(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("%s %s %s\n", r.RemoteAddr, r.Method, r.URL)
		handler.ServeHTTP(w, r)
	})
}
`
