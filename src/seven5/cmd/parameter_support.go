package cmd

import (
	"net/http"
	"os"
	"path/filepath"
	"seven5/util"
	"strings"
)

//clientSideCapability represents operations that can be performed on the client
//side when generating parameter values.  Because the client is running
//in the root directory of the project, it has easy access to 
//values that would be possible, but a hassle, to calculate completely on the 
//server side.  It also has access to the web browser request, so it can pass the
//current web request through to the Seven5 process.  Note that marshalling
//code can assume that errors have already been logged so simply returning
//in case of an error is ok.  Public so others can write commands.
type ClientSideCapability interface {
	ProjectRootDir(util.SimpleLogger) (string, error)
	ProjectConfiguration(util.SimpleLogger) (*ProjectConfig, error)
	ProjectSrcDir(util.SimpleLogger) (string, error)
	CollectFiles(util.SimpleLogger, string, bool) ([]string, error)
	CurrentWebRequest(util.SimpleLogger) (*util.BrowserRequest, error)
	SetTypeInfo(*ExplodeTypeResult)
	TypeInfo() *ExplodeTypeResult
}

//projectRootDir is suitable for use in a Command declaration as an argument.
//It wraps all the gunk around the client side capability of the same name.
//Public so others can write commands.
var ProjectRootDir = &CommandArgPair{
	Unmarshalled: nil,
	Generator: func(cl ClientSideCapability, log util.SimpleLogger) (interface{}, error) {
		return cl.ProjectRootDir(log)
	},
}

//ProjectConfig is suitable for use in a Command declaration as an argument.
//It wraps all the gunk around the client side capability of the same name.
//Public so others can write commands.
var ProjectConfiguration = &CommandArgPair{
	Unmarshalled: func() interface{} { return &ProjectConfig{} },
	Generator: func(cl ClientSideCapability, log util.SimpleLogger) (interface{}, error) {
		return cl.ProjectConfiguration(log)
	},
}

//ProjectSrcDir is suitable for use in a Command declaration as an argument.
//It wraps all the gunk around the client side capability of the same name.
//Public so others can write commands.
var ProjectSrcDir = &CommandArgPair{
	Unmarshalled: nil,
	Generator: func(cl ClientSideCapability, log util.SimpleLogger) (interface{}, error) {
		return cl.ProjectSrcDir(log)
	},
}

//CurrentWebRequest is suitable for use in a Command declaration as an argument.
//It wraps all the gunk around the client side capability of the same name.
//Public so others can write commands.
var CurrentWebRequest = &CommandArgPair{
	Unmarshalled: func() interface{} { return &util.BrowserRequest{} },
	Generator: func(cl ClientSideCapability, log util.SimpleLogger) (interface{}, error) {
		return cl.CurrentWebRequest(log)
	},
}

//ProjectRootDir is just the wd of the roadie
func (self *defaultClientSideCapability) ProjectRootDir(log util.SimpleLogger) (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		log.Error("Cannot get working directory:%s", err)
		return "", err
	}
	return wd, nil
}

//Get the current project config and return it
func (self *defaultClientSideCapability) ProjectConfiguration(log util.SimpleLogger) (*ProjectConfig, error) {
	wd, err := self.ProjectRootDir(log)
	if err != nil {
		return nil, err
	}
	cfg, err := getProjectConfig(wd)
	if err != nil {
		log.Error("Cannot get application working directory:%s", err)
		return nil, err
	}

	return cfg, nil
}

//compute the project's source code dir
func (self *defaultClientSideCapability) ProjectSrcDir(log util.SimpleLogger) (string, error) {
	cfg, err := self.ProjectConfiguration(log)
	if err != nil {
		return "", err
	}
	root, err := self.ProjectRootDir(log)
	if err != nil {
		return "", err
	}
	return filepath.Join(root, "src", cfg.AppName), nil
}
//SetTypeInfo just sets the current type knowlege that we have into this
//client capability.
func (self *defaultClientSideCapability) SetTypeInfo(t *ExplodeTypeResult) {
	self.typeInfo = t
}
//GetTypeInfo just returns the current type knowlege that we have inside this
//client capability.
func (self *defaultClientSideCapability) TypeInfo() *ExplodeTypeResult {
	return self.typeInfo 
}

//CurrentWebRequest converts the "real" type of a web request into something
//that can be easily marshalled in Json.
func (self *defaultClientSideCapability) CurrentWebRequest(log util.SimpleLogger) (*util.BrowserRequest, error) {
	browserReq, err := util.MarshalRequest(self.req, log)
	if err != nil {
		return nil, err
	}
	return browserReq,nil
}

//CollectFiles returns a slice of names that are either filenames
//or type names based on a particular suffix.  it only works on .go files so
//the suffix should not include this.  if the second param is true, it will
//convert the filenames to type names.
func (self *defaultClientSideCapability) CollectFiles(log util.SimpleLogger,
	suffix string, wantTypeNames bool) ([]string, error) {

	srcDir, err := self.ProjectSrcDir(log)
	if err != nil {
		return nil, err //already logged
	}
	f, err := os.Open(srcDir)
	if err != nil {
		return nil, err
	}
	raw, err := f.Readdir(0)
	if err != nil {
		return nil, err
	}
	result := []string{}
	fullSuffix := suffix + ".go"
	for _, n := range raw {
		if !n.IsDir() && strings.HasSuffix(n.Name(), fullSuffix) {
			result = append(result, n.Name())
		}
	}
	return result, nil
}

//ParamFromFiles is a convenience routine for using CollectFiles.  
//don't include .go in the suffix and pass true if you want type names not filenames.  
//Public so others can write commands.
func ParamFromFiles(suffix string, wantTypeNames bool) *CommandArgPair {
	return &CommandArgPair{
		func() interface{} {
			return ([]string{})
		},
		func(cl ClientSideCapability, log util.SimpleLogger) (interface{}, error) {
			return cl.CollectFiles(log, suffix, wantTypeNames)
		},
	}
}

//our implementation of the client side capabilty
type defaultClientSideCapability struct {
	req *http.Request
	typeInfo *ExplodeTypeResult
}

//NewDefaultClientCapability is the way to get an implementation of ClientCapabilities.
//Marshalling is in a different package, so must be public
func NewDefaultClientCapability(req *http.Request) ClientSideCapability {
	return &defaultClientSideCapability{req, nil}
}
