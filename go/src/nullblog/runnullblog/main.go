package main

import (
	"fmt"
	"github.com/seven5/seven5"
	"net/http"
	"nullblog" //your library
)

const (
	NAME = "nullblog"
	REST = "/rest/"
)

func main() {

	//create a project
	project := NewProject("nullblog")

	//implementation of the resources in this application
	articleRez := &nullblog.ArticleResource{}
	project.BaseDisptacher.ResourceSeparate("article", &nullblog.ArticleWire{}, articleRez, articleRez,
		nil, nil, nil)

	//generate the static dart content in to a file in the filesystem... note this
	//generates code for resources added to the dispatcher above
	project.GenerateCode()

	//start the server. this never returns.
	err := http.ListenAndServe(fmt.Sprintf(":%d", project.Heroku.Port()), project.ServeMux)
	fmt.Printf("Listen And Serve should not return! %v", err)
}
