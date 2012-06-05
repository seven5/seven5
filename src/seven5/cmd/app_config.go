package cmd

import (
	"strings"
	"encoding/json"
	"seven5/util"
	"path/filepath"
)

const (
	APP_CONFIG_FILE = "app.json"
	APPNAME = "AppName"
	DARTCOMPILERPATH = "DartCompilerPath"
)


//ApplicationConfig is read from the app.json file and passed to every 
//command.
type ApplicationConfig struct {
	AppName          string
	DartCompiler     string
}


//decodeAppConfig is a convienence routine for reading in the application
//configuration file.  typically it is used on the seven5 side by commands
//that need configuration information.
func decodeAppConfig(dir string) (*ApplicationConfig, error) {
	var app ApplicationConfig
	var err error
	var jsonBlob string
	
	jsonBlob, err = util.ReadIntoString(dir, APP_CONFIG_FILE)
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

//getAppPath is used by application's that need to know where the source code
//is for the target application.
func getAppSourceDir(dir string) (string, error) {
	cfg, err:=decodeAppConfig(dir)
	if err!=nil {
		return "",err
	}
	return filepath.Join(dir, "src", cfg.AppName), nil
}
