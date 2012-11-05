package main

import (
	"github.com/seven5/seven5"
	//"seven5"
	"net/http"
)

//sadly, we must run a server here to allow us to access resources via the network
//no cross origin resource sharing (CORS) between file: and http:, where our api doc is
func main() {
	h:=seven5.NewSimpleHandler();
	http.ListenAndServe(":3004", seven5.DefaultProjectBindings(h, "docbrowser"))
}