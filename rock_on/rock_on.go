package main

import (
	"os"
	"fmt"
	"time"
	"seven5"
	"os/exec"
	"path/filepath"
)

type Rock struct {
	projectDir *os.File
	dirMonitor *seven5.DirectoryMonitor
	buildCmd *exec.Cmd
	webCmd *exec.Cmd
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
		time.Sleep(1e9)
	}
}

func (rock *Rock) Build() (success bool) {
	if rock.buildCmd != nil {
		rock.buildCmd.Process.Kill()
		rock.buildCmd = nil
	}
	packageName := CurrentDirName()

	// Tune
	rock.buildCmd  = exec.Command("tune", packageName)
	output, err := rock.buildCmd.CombinedOutput()
	if output != nil {
		fmt.Print(string(output))
	}
	if err != nil{
		fmt.Println("Error tuning: ", err)
		return false
	}

	// Build
	rock.buildCmd  = exec.Command("gb", "-c", "-t", ".")
	output, err = rock.buildCmd.CombinedOutput()
	rock.buildCmd = nil
	if output != nil {
		fmt.Print(string(output))
	}
	if err != nil{
		fmt.Println("Error building: ", err)
		return false
	}
	return true
}

func (rock *Rock) NeedsRebuild() bool {
	// TODO: Poll the directory monitor
	return false
}
 
func (rock *Rock) StartServer() (success bool) {
	fmt.Println("Running the web server")
	// TODO: Run webapp_start and save to rock.webProcess
	return true
}

func (rock *Rock) StopServer() (success bool) {
	fmt.Println("Stopping the web server")
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

func CurrentDirName() string{
	abs,_ := filepath.Abs(".")
	return filepath.Base(abs)
}

func main() {
	rock, err := NewRock(".")
	if err != nil {
		fmt.Println(err)
		return
	}
	rock.Run()
}
