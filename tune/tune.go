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

func GenerateMain(importPath string, name ...string) (main_go string, err error) {

	base := filepath.Clean(filepath.Base(importPath))

	myFuncs := make(map[string]interface{})
	myFuncs["upper"]=toUpperFirst

	t := template.Must(template.New("main").Funcs(myFuncs).Parse(seven5.WEBAPP_TEMPLATE))

	data := make(map[string]interface{})

	data["import"] = importPath
	data["package"] = base
	data["handler"] = name //array

	buff := bytes.NewBufferString("")
	if err = t.Execute(buff, data); err != nil {
		return
	}

	return string(buff.Bytes()), nil
}

func WriteMain(main_go string) (err error) {
	dir, err := os.Stat(seven5.WEBAPP_START_DIR)
	if err != nil { return }
	if !dir.IsDirectory() {
		err = errors.New(seven5.WEBAPP_START_DIR + " exists but is not a directory")
		return
	}
	main_file,err := os.Create(filepath.Join(seven5.WEBAPP_START_DIR, "main.go"))
	if err != nil { return }
	wr := bufio.NewWriter(main_file) 
	wr.WriteString(main_go) 
	wr.Flush()
	fmt.Println("Generated " + filepath.Join(seven5.WEBAPP_START_DIR, "main.go"))
	return
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: tune <project-package-name> <handler 1> <handler 2>")
		os.Exit(1)
	}

	main_go, err := GenerateMain(os.Args[1], os.Args[2:]...)
	if err != nil {
		fmt.Printf("Error processing template: %s\n",err.Error())
		os.Exit(1)
	}
	
	err = WriteMain(main_go)
	if err != nil { 
		fmt.Printf("Error writing main.go: %s\n",err.Error())
		os.Exit(1)
	}
	
	return

}