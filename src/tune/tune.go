/*
Tune generates all of the necessary pre-compile information for this project.
During development you'd usually use rock_on instead directly running tune.
For deployments, run tune in your package directory before running gb.
*/
package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"seven5"
	"strings"
	"text/template"
)

func toUpperFirst(x string) string {
	return strings.ToUpper(x[0:1]) + x[1:]
}

func importIfNeeded(x string,exported *seven5.ExportedSeven5Objects) string {
	if len(exported.Model)==0 {
		return ""
	}
	return "\""+x+"\""
}
func GenerateMain(importPath string, base string, exported *seven5.ExportedSeven5Objects) (string,error) {

	myFuncs := make(map[string]interface{})
	myFuncs["upper"] = toUpperFirst
	//WEIRD that this must take interface{} when the toUpperFirst is ok with string
	myFuncs["importIfNeeded"] = func (x interface{}) string {
		return importIfNeeded(fmt.Sprintf("%s",x),exported)
	}
	t := template.Must(template.New("main").Funcs(myFuncs).Parse(seven5.WEBAPP_TEMPLATE))

	data := make(map[string]interface{})

	data["import"] = importPath
	data["package"] = base
	data["model"] = exported.Model

	buff := bytes.NewBufferString("")
	if err := t.Execute(buff, data); err != nil {
		return "",err
	}

	return string(buff.Bytes()), nil
}

func WriteMain(main_go string, projectName string) (err error) {
	dir, err := os.Stat(seven5.WEBAPP_START_DIR)
	if err != nil {
		_, ok := err.(*os.PathError)
		if !ok {
			return
		}
		var perm uint32 = 0777
		err = os.Mkdir(seven5.WEBAPP_START_DIR, perm)
		if err != nil {
			return
		}
		//try again
		dir, err = os.Stat(seven5.WEBAPP_START_DIR)
		if err != nil {
			return
		}
	}
	if !dir.IsDir() {
		err = errors.New(seven5.WEBAPP_START_DIR + " exists but is not a directory")
		return
	}
	mainPath := filepath.Join(seven5.WEBAPP_START_DIR, projectName+".go")
	main_file, err := os.Create(mainPath)
	if err != nil {
		return
	}
	wr := bufio.NewWriter(main_file)
	wr.WriteString(main_go)
	wr.Flush()
	fmt.Println("Generated " + mainPath)
	return
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "Usage: tune <project-package-name>")
		os.Exit(1)
	}

	imp := os.Args[1]
	base := filepath.Clean(filepath.Base(imp))
	
	cwd,err :=os.Getwd()
	if err!=nil {
		fmt.Fprintf(os.Stderr,"cannot get working directory!\n")
		return
	}
	
	var exported seven5.ExportedSeven5Objects
	seven5.CheapASTAnalysis(cwd,&exported)

	main_go, err := GenerateMain(imp, base, &exported)

	if err != nil {
		fmt.Printf("Error processing template: %s\n", err.Error())
		os.Exit(1)
	}

	err = WriteMain(main_go, base)
	if err != nil {
		fmt.Printf("Error writing main.go: %s\n", err.Error())
		os.Exit(1)
	}

	return
}
