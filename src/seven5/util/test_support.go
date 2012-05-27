package util

import (
	"os"
	"strings"
	"path/filepath"
	"bytes"
)

//FindTestDataPath looks for the "testdata" directory inside the GOPATH 
//environment variable components and then tries to locate the 
//insideTestData directory within that.
func FindTestDataPath(logger SimpleLogger, comp...string) string {
	dirs:= strings.Split(os.Getenv("GOPATH"),string(filepath.ListSeparator))
	userPart:=filepath.Join(comp...)
	path := filepath.Join("testdata",userPart)
	for _,d:= range(dirs) {
		candidate := filepath.Join(d,path)
		if _,err:=os.Stat(candidate); err==nil {
			return candidate
		}
	}
	msg :=
`Your GOPATH environment variable does not include the Seven5 source code
tree so we cannot find the testdata and thus cannot run tests. Some element
of GOPATH should include testdata as its direct child.`	
	logger.Panic(msg)
	return "" //wont happen
}

//ReadTestData reads a test data file into a string.
func ReadTestData(logger SimpleLogger, pathcomponent... string,) string {
	var buffer bytes.Buffer
	var err error
	var file *os.File
	fullPath:=FindTestDataPath(logger,pathcomponent...) 
	file, err = os.Open(fullPath)
	if err!=nil {
		logger.Panic("Unable to read file %s:%s",fullPath,err)
	}
	if _, err = buffer.ReadFrom(file); err != nil {
		logger.Panic("Unable to find a test data file: %s:%s",file,err)
	}
	
	return buffer.String()
}
