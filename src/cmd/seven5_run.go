package main

import (
	"flag"
	"fmt"
	"os"
	"seven5"
)

func main() {
	flag.Parse()
	if flag.NArg()==0 {
		fmt.Fprintf(os.Stderr,"usage: seven5_run [project name]\n")
		return
	}
	projectDir, p := seven5.LocateProject(flag.Arg(0)) 
	if projectDir=="" {
		fmt.Fprintf(os.Stderr,"usage: seven5_run [project name]\n")
		fmt.Fprintf(os.Stderr,"your current directory is %s\n",p)
		fmt.Fprintf(os.Stderr,"seven5_run expects you to be in :\n")
		fmt.Fprintf(os.Stderr,"your project directory, the parent of your project directory such that [current directory]/[project name] has your project's files,\n")
		fmt.Fprintf(os.Stderr,"or the eclipse bin directory such that ../../src/pkg/[project name] is your project\n")
	}
	
	if !seven5.VerifyProjectLayout(projectDir) {
		fmt.Fprintf(os.Stderr,"project %s does not have the standard seven5 project structure")
		fmt.Fprintf(os.Stderr,"see http://seven5.github.com/seven5/project_layout.html")
	}
}