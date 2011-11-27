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
	rock.StopServer()
	rock.webCmd  = exec.Command(filepath.Join(".", seven5.WEBAPP_START_DIR, seven5.WEBAPP_START_DIR), ".")
	rock.webCmd.Stdout = os.Stdout
	rock.webCmd.Stderr = os.Stderr
	err := rock.webCmd.Start()
	if err != nil{
		fmt.Println("Error starting the web server: ", err)
		return false
	}
	return true
}

func (rock *Rock) StopServer() (success bool) {
	if rock.webCmd == nil { return true }
	rock.webCmd.Process.Kill()
	rock.webCmd = nil
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
