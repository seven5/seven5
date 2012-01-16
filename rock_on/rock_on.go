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
	projectDir string
	dirMonitor *seven5.DirectoryMonitor
	buildCmd *exec.Cmd
	webCmd *exec.Cmd
}

// This is the main loop in which Rock controls the build and server process and watches for source changes
func (rock *Rock) Run(projectName string) {
	if rock.Build(projectName){
		rock.StartServer(projectName)
	}
	for {
		if rock.NeedsRebuild(projectName){
			rock.StopServer()
			if rock.Build(projectName) {
				rock.StartServer(projectName)
			}
		}
		time.Sleep(1e9)
	}
}

func (rock *Rock) Build(packageName string) (success bool) {
	cmd:="tune"
	// Tune
	rock.buildCmd  = exec.Command(cmd, packageName)
	output, err := rock.buildCmd.CombinedOutput()
	if output != nil {
		fmt.Print(string(output))
	}
	if err != nil{
		fmt.Println("Error tuning: ", err)
		return false
	}

	// Build library
	rock.buildCmd  = exec.Command("go", "install", ".")
	rock.buildCmd.Dir = rock.projectDir
	output, err = rock.buildCmd.CombinedOutput()
	
	if output != nil {
		fmt.Print(string(output))
	}
	if err != nil{
		fmt.Println("Error building "+packageName+": ", err)
		return false
	}

	//Build executable
	rock.buildCmd  = exec.Command("go", "build", "-o", packageName, ".")
	rock.buildCmd.Dir = filepath.Join(rock.projectDir,seven5.WEBAPP_START_DIR)
	output, err = rock.buildCmd.CombinedOutput()
	

	return true
}

func (rock *Rock) NeedsRebuild(projectName string) bool {
	// TODO: Poll the directory monitor
	changed, err := rock.dirMonitor.Poll()
	if err != nil {
		fmt.Println("Error monitoring the directory", err)
		os.Exit(1)
	} else if changed {
		fmt.Printf("\n--------------------------------------\nProject %s changed\n--------------------------------------\n", projectName)
	}
	return changed
}
 
func (rock *Rock) StartServer(projectName string) (success bool) {
	rock.StopServer()
	
	cmdName:=filepath.Clean(filepath.Base(projectName))
	rock.webCmd  = exec.Command(filepath.Join(rock.projectDir, seven5.WEBAPP_START_DIR, cmdName), "dungheap")
	rock.webCmd.Dir = rock.projectDir
	

	fmt.Printf("---- '%s'  '%s'\n",rock.webCmd.Dir, rock.projectDir)
	
	rock.webCmd.Stdout = os.Stdout
	rock.webCmd.Stderr = os.Stderr
	err := rock.webCmd.Start()
	if err != nil{
		fmt.Println("Error starting the project application: ", err)
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
	_, err = os.Open(projectDirPath)
	if err != nil { return }
	dirMon, err := seven5.NewDirectoryMonitor(projectDirPath, ".go")
	if err != nil { return }
	return &Rock{projectDirPath, dirMon, nil, nil}, nil
}

func CurrentDirName() string{
	abs,_ := filepath.Abs(".")
	return filepath.Base(abs)
}

func main() {
	dir,err:=os.Getwd()
	if err!=nil {
		fmt.Fprintf(os.Stderr,"unable to get current directory: aborting:%s\n",err)
		return
	}
	projName:=filepath.Clean(filepath.Base(dir))
	if len(os.Args) == 1 {	
		fmt.Fprintf(os.Stdout,"no directory given, hope project name '%s' is ok\n", projName)
	} else {
		projName = os.Args[1]
	}
	
	fmt.Printf("monitoring directory '%s' for project '%s'\n",dir,projName)
	rock, err := NewRock(dir)
	if err != nil {
		fmt.Println(err)
		return
	}
	rock.Run(projName)
}
