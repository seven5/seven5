package main

import (
	"flag"
	"fmt"
	"os"
	"seven5"
)

func main() {
	flag.Parse()
	if flag.NArg() == 0 {
		fmt.Fprintf(os.Stderr, "usage: seven5_run [project name]\n")
		return
	}
	projectDir, p := seven5.LocateProject(flag.Arg(0))
	if projectDir == "" {
		fmt.Fprintf(os.Stderr, "usage: seven5_run [project name]\n")
		fmt.Fprintf(os.Stderr, "your current directory is %s\n", p)
		fmt.Fprintf(os.Stderr, "seven5_run expects you to be in :\n")
		fmt.Fprintf(os.Stderr, "your project directory, the parent of your project directory such that [current directory]/[project name] has your project's files,\n")
		fmt.Fprintf(os.Stderr, "or the eclipse bin directory such that ../../src/pkg/[project name] is your project\n")
		return
	}

	if err := seven5.VerifyProjectLayout(projectDir); err != "" {
		fmt.Fprintf(os.Stderr, "%s does not have the standard seven5 project structure!\n", projectDir)
		fmt.Fprintf(os.Stderr, "%s", err)
		fmt.Fprintf(os.Stderr, "for project structure details, see http://seven5.github.com/seven5/project_layout.html")
		return
	}

	logger, path, err := seven5.CreateLogger(projectDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to create logger at path %s\n", path)
		return
	}

	fmt.Printf("Seven5 is logging to %s\n", path)

	config := seven5.NewProjectConfig(projectDir, logger)
	config.Logger.Printf("Starting to run with project %s at %s", config.Name, config.Path)

	err = seven5.ClearTestDB(config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error clearing the test db configuration:%s\n", err.Error())
		return
	}

	_,err = seven5.DiscoverHandlers(config)
	if err!=nil {
		fmt.Fprintf(os.Stderr, "unable to discover mongrel2 handlers:%s\n", err.Error())
		return
	}
}
