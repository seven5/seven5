package seven5

import (
	"strings"
	"encoding/json"
	"seven5/util"
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

//substitute for constant array
func MANDATORY_PARAMS() []string {
	return []string{
		APPNAME,
	}
}

//substitute for constant array
func ALL_PARAMS() []string {
	return []string{
		APPNAME,
		DARTCOMPILERPATH,
	}
}

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
