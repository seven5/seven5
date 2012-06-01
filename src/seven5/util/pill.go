package util

import (
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

//this is ok because it's not used for crypto just generating random filenames
var pillRand = rand.New(rand.NewSource(time.Now().Unix()))

//MakePillDir is a  convenience for making a random pill directory and 
//returning the name of it.
func MakePillDir(logger SimpleLogger) (string, error) {
	x := pillRand.Intn(1000)
	y := pillRand.Intn(1000)
	dir := os.TempDir()
	name := filepath.Join(dir, fmt.Sprintf("pill%d-%d", x, y))
	var dirPerm os.FileMode = os.ModeDir | 0777
	logger.Debug("Creating pill directory %s", name)

	if err := os.Mkdir(name, dirPerm); err != nil {
		logger.Error("Can't create temp dir for pill: %s", err)
		return "", err
	}
	return name, nil
}

//CompilePill takes the text provided and generates the main and compiles it. It
//returns the name of the executable or "" in the firs return value. It returns
//a non-empty second value if the compile has failed containing the error text.
func CompilePill(mainCode string, logger SimpleLogger) (string, string, error) {
	var file *os.File
	var err error
	var previousCwd string
	var dir string
	
	if dir, err= MakePillDir(logger); err!=nil {
		return "", "", nil
	}
	mainName := filepath.Join(dir, "main.go")
	file, err = os.Create(mainName)
	if err != nil {
		logger.Error("Unable to create main file: %s", err)
		return "", "", err
	}

	if _, err = file.WriteString(mainCode); err != nil {
		logger.Error("Unable to write to main file in bootstrap pill: %s", err)
		return "", "", err
	}

	if err = file.Close(); err != nil {
		logger.Error("Unable to close main file in bootstrap pill: %s", err)
	return "", "", err
	}

	if previousCwd, err = os.Getwd(); err != nil {
		logger.Error("Unable get working dir before chdir: %s", err)
		return "", "", err
	}

	if err = os.Chdir(dir); err != nil {
		logger.Error("Unable to change to bootstrap pill dir: %s", err)
		return "", "", err
	}
	defer func() {
		if err = os.Chdir(previousCwd); err != nil {
			logger.Error("Unable to change back to previous dir after creating pill: %s", err)
		}

	}()

	cmd := exec.Command("go", "build")
	var buf []byte
	if buf, err = cmd.CombinedOutput(); err != nil {
		return "", string(buf)+"\n"+ err.Error(), nil
	}
	//this weird construction represents the result of running go build on a main
	return filepath.Join(dir, filepath.Base(dir)), "", nil
}
