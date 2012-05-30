package util

import (
	"path/filepath"
	"bytes"
	"errors"
	"fmt"
	"os"
)

func ReadIntoString(dir string, filename string) (string,error) {
	var file *os.File
	var info os.FileInfo
	var err error
	var buffer bytes.Buffer
	
	fpath := filepath.Join(dir, filename)
	if  info, err =os.Stat(fpath); err!=nil {
		return "", err
	} 
	if info.IsDir() {
		return "", errors.New(fmt.Sprintf("in %s, %s is a directory not a file!",
			dir, filename))
	}	
	
	if file, err = os.Open(fpath); err != nil {
		return "", err
	}
	
	if _, err = buffer.ReadFrom(file); err != nil {
		return "", err
	}
	
	return buffer.String(),nil
}