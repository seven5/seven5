package main

import (
	"github.com/seven5/seven5"
	"italy"
	"net/http"
	"fmt"
)

const (
	NAME = "italy"
)

func main() {

	//create a new empty URL space ... we are not using a cookie handler
	h := seven5.NewSimpleHandler(nil)
	
	//add our resource
	h.AddResource(italy.ItalianCity{}, &italy.ItalianCityResource{})

	//the GOPATH var is used to find where the parts of our project are and how to find our
	//Google maps key
	env := seven5.NewEnvironmentVars(NAME)
	
	//we are using the default layout and want the default bindings... we don't use the
	//environment for this application
	seven5.DefaultProjectBindings(h, NAME, env)

	//normal http calls for running a server in go... ListenAndServe never should return
	//err:=http.ListenAndServe(":3003",logHTTP(asHttp))

	//use this verson, not the one above, if you want to log HTTP requests to the terminal
	err := http.ListenAndServe(":3003", logHTTP(http.Handler(h)))

	fmt.Printf("Error! Returned from ListenAndServe(): %s", err)
}

// tiny wrapper around all the HTTP dispatching that can be nice to help with debugging
func logHTTP(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("%s %s %s\n", r.RemoteAddr, r.Method, r.URL)
		handler.ServeHTTP(w, r)
	})
}
