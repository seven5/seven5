package util

import (
	"os"
	"strings"
	"path/filepath"
	"bytes"
	"launchpad.net/gocheck"
)

//FindTestDataPath looks for the "testdata" directory inside the GOPATH 
//environment variable components and then tries to locate the 
//insideTestData directory within that.
func FindTestDataPath(c *gocheck.C, comp...string) string {
	dirs:= strings.Split(os.Getenv("GOPATH"),string(filepath.ListSeparator))
	userPart:=filepath.Join(comp...)
	path := filepath.Join("testdata",userPart)
	for _,d:= range(dirs) {
		candidate := filepath.Join(d,path)
		if _,err:=os.Stat(candidate); err==nil {
			return candidate
		}
	}
	c.Errorf("Unable to find test data %s on GOPATH: %s",
		path,os.Getenv("GOPATH"))
	c.Fail()
	return "" //wont happen
}

//ReadTestData reads a test data file into a string.
func ReadTestData(c *gocheck.C, pathcomponent... string,) string {
	var buffer bytes.Buffer
	var err error
	var file *os.File
	fullPath:=FindTestDataPath(c,pathcomponent...) 
	file, err = os.Open(fullPath)
	if err!=nil {
		c.Fatal("Unable to read file %s:%s",fullPath,err)
	}
	if _, err = buffer.ReadFrom(file); err != nil {
		c.Fatal("Unable to find a test data file: %s:%s",file,err)
	}
	
	return buffer.String()
}
