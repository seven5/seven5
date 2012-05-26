package util

import (
	"os"
	"math/rand"
	"time"
	"path/filepath"
	"fmt"
	"os/exec"
	"strings"
)


//this is ok because it's not used for crypto just generating random filenames
var pillRand = rand.New(rand.NewSource(time.Now().Unix()))


//MakePillDir is a  convenience for making a random pill directory and 
//returning the name of it.
func MakePillDir(logger SimpleLogger) string {
	x:=pillRand.Intn(1000)
	y:=pillRand.Intn(1000);
	dir:=os.TempDir()
	name := filepath.Join(dir,fmt.Sprintf("pill%d-%d",x,y))
	var dirPerm os.FileMode = os.ModeDir | 0777
	logger.Debug("Creating pill directory %s",name)
	
	if err:=os.Mkdir(name,dirPerm); err!=nil {
		logger.Panic("Can't create temp dir for pill: %s",err)
	}
	return name
}

//CompilePill takes the text provided and generates the main and compiles it. It
//returns the name of the executable or "" in the firs return value. It returns
//a non-empty second value if the compile has failed containing the error text.
func CompilePill(mainCode string, logger SimpleLogger) (string, string) {
	var file *os.File
	var err error

	dir:=MakePillDir(logger)
	mainName := filepath.Join(dir,"main.go")
	file, err = os.Create(mainName)
	if err!=nil {
		logger.Panic("Unable to create main file: %s",err)
	}

	if _,err = file.WriteString(mainCode); err!=nil {
		logger.Panic("Unable to write to main file in bootstrap pill: %s",err)
	}
		
	if err=file.Close(); err!=nil {
		logger.Panic("Unable to close main file in bootstrap pill: %s",err)
	}
	if err = os.Chdir(dir); err!=nil {
		logger.Panic("Unable to change to bootstrap pill dir: %s",err)
	}
	
	cmd := exec.Command("go","build")
	var buf []byte
	if buf, err = cmd.CombinedOutput(); err!=nil {
		return "", string(buf)
	} 
	slice := strings.SplitAfter(dir,string(filepath.Separator))
	return slice[len(slice)-1],""
}
