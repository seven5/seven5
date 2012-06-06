package cmd

import (
	"strings"
	"encoding/json"
	"seven5/util"
)

const (
	PROJECT_CONFIG_FILE = "project.json"
	APPNAME = "AppName"
	DARTCOMPILERPATH = "DartCompilerPath"
)


//ProjectConfig is a type representing the contents of the project.json file.
type ProjectConfig struct {
	AppName          string
	DartCompiler     string
}


//getProjectConfig is a convienence routine for reading in the application
//configuration file for a root project directory.
func getProjectConfig(dir string) (*ProjectConfig, error) {
	var app ProjectConfig
	var err error
	var jsonBlob string
	
	jsonBlob, err = util.ReadIntoString(dir, PROJECT_CONFIG_FILE)
	if err!=nil {
		
		return nil, err
	}
	decoder:=json.NewDecoder(strings.NewReader(jsonBlob))
	err = decoder.Decode(&app)
	if err!=nil{
		return nil, err
	}
	return &app, nil
}
