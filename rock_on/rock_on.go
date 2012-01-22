package main

import (
	"os"
	"fmt"
	"time"
	"seven5"
	"os/exec"
	"path/filepath"
	"seven5/store"
)

type Rock struct {
	projectDir string
	dirMonitor *seven5.DirectoryMonitor
	buildCmd *exec.Cmd
	webCmd *exec.Cmd
}

type RockListener struct {
}

func (self *RockListener) FileChanged(fileInfo os.FileInfo) {
	fmt.Printf("[ROCK ON] %s changed\n",fileInfo.Name())
}
func (self *RockListener) FileAdded(fileInfo os.FileInfo) {
	fmt.Printf("[ROCK ON] %s added\n",fileInfo.Name())
}
func (self *RockListener) FileRemoved(fileInfo os.FileInfo) {
	fmt.Printf("[ROCK ON] %s removed\n",fileInfo.Name())
}

//destroyStore is a utility routine that is useful for clearing all the data when you are
//running tests.
func destroyStore() error {
	return store.NewStoreImpl(store.MEMCACHE_LOCALHOST).DestroyAll(store.MEMCACHE_LOCALHOST)
}

// This is the main loop in which Rock controls the build and server process and watches for source changes
func (rock *Rock) Run(projectName string) {
	if rock.Build(projectName){
		rock.StartServer(projectName)
	}
	for {
		if rock.NeedsRebuild(projectName){
			start:=time.Now()
			rock.StopServer(projectName)
			stop:=time.Now()
			if rock.Build(projectName) {
				build:=time.Now()
				rock.StartServer(projectName)
				end:=time.Now()
				duration:=end.Sub(start)
				stopDuration:=stop.Sub(start)
				buildDuration:=build.Sub(stop)
				startDuration:=end.Sub(build)
				fmt.Printf("[ROCK ON] restart took %4.4f seconds (%4.4f to stop, %4.4f to build, %4.4f to start)\n",
				duration.Seconds(), stopDuration.Seconds(), buildDuration.Seconds(), startDuration.Seconds())
			}
		}
		time.Sleep(1e9)
	}
}

func (rock *Rock) Build(packageName string) (success bool) {
	cmd:="tune"
	// Tune
	start:=time.Now()
	rock.buildCmd  = exec.Command(cmd, packageName)
	output, err := rock.buildCmd.CombinedOutput()
	if output != nil {
		fmt.Print(string(output))
	}
	if err != nil{
		fmt.Println("[ROCK ON] Error tuning: ", err)
		return false
	}
	codeGen:=time.Now()
	
	// Build library
	rock.buildCmd  = exec.Command("go", "install", ".")
	rock.buildCmd.Dir = rock.projectDir
	output, err = rock.buildCmd.CombinedOutput()
	
	if output != nil {
		fmt.Print(string(output))
	}
	if err != nil{
		fmt.Println("[ROCK ON] Error building "+packageName+": ", err)
		return false
	}
	library:=time.Now()

	//Build executable
	rock.buildCmd  = exec.Command("go", "build", "-o", packageName, ".")
	rock.buildCmd.Dir = filepath.Join(rock.projectDir,seven5.WEBAPP_START_DIR)
	output, err = rock.buildCmd.CombinedOutput()
	if output != nil {
		fmt.Print(string(output))
	}
	if err != nil{
		fmt.Println("[ROCK ON] Error building executable (_seven5/"+packageName+"): ", err)
		return false
	}
	
	
	executable:=time.Now()
	
	genDuration:=codeGen.Sub(start)
	libraryDuration:=library.Sub(codeGen)
	executableDuration:=executable.Sub(library)
	
	fmt.Printf("[ROCK ON] Build stats: %4.4f secs for code generation, %4.4f secs %s library, %4.4f secs executable\n",
		genDuration.Seconds(), libraryDuration.Seconds(), packageName, executableDuration.Seconds())

	return true
}

func (rock *Rock) NeedsRebuild(projectName string) bool {
	// TODO: Poll the directory monitor
	changed, err := rock.dirMonitor.Poll()
	if err != nil {
		fmt.Println("[ROCK ON] Error monitoring the directory", err)
		os.Exit(1)
	} else if changed {
		fmt.Printf("\n--------------------------------------\nProject %s changed\n--------------------------------------\n", projectName)
	}
	return changed
}
 
func (rock *Rock) StartServer(projectName string) (success bool) {
	rock.StopServer(projectName)
	
	cmdName:=filepath.Clean(filepath.Base(projectName))
	rock.webCmd  = exec.Command(filepath.Join(rock.projectDir, seven5.WEBAPP_START_DIR, cmdName), projectName)
	rock.webCmd.Dir = filepath.Clean(filepath.Join(rock.projectDir,".."))
	

	fmt.Printf("[ROCK ON] running '%s' from '%s'\n",cmdName,rock.webCmd.Dir)
	
	rock.webCmd.Stdout = os.Stdout
	rock.webCmd.Stderr = os.Stderr
	err := rock.webCmd.Start()
	if err != nil{
		fmt.Println("[ROCK ON] Error starting the project application: ", err)
		return false
	}
	err =rock.webCmd.Wait()
	if err != nil{
		fmt.Println("[ROCK ON] Executable _seven5/"+projectName+" exited with an error: ", err)
		return false
	} else {
		fmt.Println("[ROCK ON] Executable _seven5/"+projectName+" was shut down")
	}
	
	return true
}

func (rock *Rock) StopServer(projectName string) (success bool) {
	if rock.webCmd == nil { return true }
	fmt.Printf("[ROCK ON] Killing old version of %s\n", projectName)
	rock.webCmd.Process.Kill()
	rock.webCmd = nil
	return true
}

func NewRock(projectDirPath string) (rock *Rock, err error) {
	_, err = os.Open(projectDirPath)
	if err != nil { return }
	dirMon, err := seven5.NewDirectoryMonitor(projectDirPath, ".go")
	if err != nil { return }
	dirMon.Listen(&RockListener{})
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
		fmt.Fprintf(os.Stdout,"[ROCK ON] no directory given, hope project name '%s' is ok\n", projName)
	} else {
		projName = os.Args[1]
	}
	
	fmt.Printf("[ROCK ON] monitoring directory '%s' for project '%s'\n",dir,projName)
	rock, err := NewRock(dir)
	if err != nil {
		fmt.Println(err)
		return
	}
	rock.Run(projName)
}
