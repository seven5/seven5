package main

import (
	"os"
	"fmt"
	"os/exec"
	"time"
	"seven5"
)

type Rock struct {
	projectDir *os.File
	dirMonitor *seven5.DirectoryMonitor
	buildProcess *exec.Cmd
	webProcess *exec.Cmd
}

// This is the main loop in which Rock controls the build and server process and watches for source changes
func (rock *Rock) Run() {
	if rock.Build(){
		rock.StartServer()
	}
	for {
		if rock.NeedsRebuild(){
			rock.StopServer()
			if rock.Build() {
				rock.StartServer()
			}
		}
		fmt.Print(".")
		time.Sleep(1e9)
	}
}

func (rock *Rock) Build() (success bool) {
	// TODO: Run the code generation command
	// TODOL Run gb 
	return true
}

func (rock *Rock) NeedsRebuild() bool {
	// TODO: Poll the directory monitor
	return false
}
 
func (rock *Rock) StartServer() (success bool) {
	// TODO: Run webapp_start and save to rock.webProcess
	return true
}

func (rock *Rock) StopServer() (success bool) {
	//TODO: kill the rock.webProcess
	return true
}

func NewRock(projectDirPath string) (rock *Rock, err error) {
	dir, err := os.Open(projectDirPath)
	if err != nil { return }
	dirMon, err := seven5.NewDirectoryMonitor(projectDirPath)
	if err != nil { return }
	return &Rock{dir, dirMon, nil, nil}, nil
}

func main() {
	rock, err := NewRock(".")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Rock on!")
	rock.Run()
}
