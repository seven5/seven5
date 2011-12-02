/*
Tune generates all of the necessary pre-compile information for this project.
During development you'd usually use rock_on instead directly running tune.
For deployments, run tune in your package directory before running gb.
*/ 
package main

import (
	"os"
	"fmt"
	"bytes"
	"bufio"
	"errors"
	"seven5"
	"strings"
	"text/template"
	"path/filepath"
)

func toUpperFirst(x string) string {
	return strings.ToUpper(x[0:1])+x[1:]
}

func GenerateMain(importPath string, base string, handlers []string, cssSources []string, htmlSources []string) (main_go string, err error) {
	
	myFuncs := make(map[string]interface{})
	myFuncs["upper"]=toUpperFirst

	t := template.Must(template.New("main").Funcs(myFuncs).Parse(seven5.WEBAPP_TEMPLATE))

	data := make(map[string]interface{})

	data["import"] = importPath
	data["package"] = base
	data["handler"] = handlers 
	data["css"] = cssSources 
	data["html"] = htmlSources 

	buff := bytes.NewBufferString("")
	if err = t.Execute(buff, data); err != nil {
		return
	}

	return string(buff.Bytes()), nil
}

func WriteMain(main_go string, projectName string) (err error) {
	dir, err := os.Stat(seven5.WEBAPP_START_DIR)
	if err != nil { 
		_,ok:=err.(*os.PathError)
		if !ok {
			return
		}
		var perm uint32 = 0777 
		err = os.Mkdir(seven5.WEBAPP_START_DIR,perm)
		if err!=nil {
			return
		}
		//try again
		dir, err = os.Stat(seven5.WEBAPP_START_DIR)
		if err!=nil {
			return
		}
	}
	if !dir.IsDirectory() {
		err = errors.New(seven5.WEBAPP_START_DIR + " exists but is not a directory")
		return
	}
	mainPath := filepath.Join(seven5.WEBAPP_START_DIR, projectName+".go")
	main_file,err := os.Create(mainPath)
	if err != nil { return }
	wr := bufio.NewWriter(main_file) 
	wr.WriteString(main_go) 
	wr.Flush()
	fmt.Println("Generated " + mainPath)
	return
}

func main() {
	if len(os.Args) !=5 {
		fmt.Fprintln(os.Stderr,"Usage: tune <project-package-name> <handlers> <css sources> <html sources> (quote groups, space separated--empty groups must be empty-quoted)")
		os.Exit(1)
	}

	imp:=os.Args[1]
	
	handlers:=strings.Split(os.Args[2], " ")
	if len(handlers[0])==0 {
		handlers=[]string{}
	}
	cssSources:=strings.Split(os.Args[3]," ")
	if len(cssSources[0])==0 {
		cssSources=[]string{}
	}
	htmlSources:=strings.Split(os.Args[4]," ")
	if len(htmlSources[0])==0 {
		htmlSources=[]string{}
	}
		
	base := filepath.Clean(filepath.Base(imp))
	
	main_go, err := GenerateMain(imp,base,handlers,cssSources,htmlSources)
	
	if err != nil {
		fmt.Printf("Error processing template: %s\n",err.Error())
		os.Exit(1)
	}
	
	err = WriteMain(main_go,base)
	if err != nil { 
		fmt.Printf("Error writing main.go: %s\n",err.Error())
		os.Exit(1)
	}
	
	return
}
