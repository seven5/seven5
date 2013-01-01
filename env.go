package seven5

import (
	"os"
	"fmt"
	"strings"
	"path/filepath"
)

//FileFlavor is used to "find" parts of your application at run-time. Use of these constants
//means that if the default project layout changes you only have to move your files around,
//not change your code.
type FileFlavor int

const (
	GO_SOURCE_FLAVOR = iota
	DART_FLAVOR
	ASSET_FLAVOR
	TOP_LEVEL_FLAVOR
)

//Environment encodes information that cannot be obtained from the source code, either because
//it can't be known or it is undesirable to commit these values to the source code repository.
//It is also used to "figure out" the location of project entities based on the standard
//project layout.

//ProjectObject computes a path inside a seven5 project that has the default
//layout.  You need to supply a project name and the directory you are looking for.  If
//you set
//
//For project foo, the default layout is:
// foo/
//    Procfile
//    .godir
//    dart/
//					foo/
//					      web/
//					      			assets/
//                pubspec.yaml
//                pubspec.lock
//                packages/
//                ...
//    db/
//    go/
//         bin/
//         pkg/
//         src/
//               foo/
//                     runfoo/
//                     				main.go

type Environment interface {
	IsTest() bool
	ClientId(AuthServiceConnector) string
	ClientSecret(AuthServiceConnector) string
	GetValueOrPanic(string) string
	ProjectObject(target string, projectName string, flavor FileFlavor) (string, error) 
}

//EnvironmentVars is an Environment implementation that reads values from a standard arrangement
//of unix-ish environment variables.  Typically the enivornment variables are prefixed with the
//application name.
type EnvironmentVars struct {
	name string
}

//NewEnvironmentVars returns an initialied EnvironmentVars based on the name provided.
func NewEnvironmentVars(appName string) Environment {
	return &EnvironmentVars {
		name: strings.ToUpper(appName),
	}
}

//IsTest returns true if the variable APPNAME_TEST is set to a non empty value. This value
//is read only once. 
func (self *EnvironmentVars) IsTest() bool {
	return os.Getenv(fmt.Sprintf("%s_TEST", self.name)) != ""
}

//ClientId returns the value of the client id that
//has been given out the by the service associated with conn.  The environment variable
//is APPNAME_SERVICENAME_CLIENT_ID and is read only once.
func (self *EnvironmentVars) ClientId(conn AuthServiceConnector) string {
	return self.GetValueOrPanic(fmt.Sprintf("%s_%s_CLIENT_ID", self.name, strings.ToUpper(conn.Name())))
}

//ClientSecret returns the value of the environment value of the client secret that
//has been given out the by the service associated with conn.  The environment variable
//is APPNAME_SERVICENAME_CLIENT_SECRET and is read only once.
func (self *EnvironmentVars) ClientSecret(conn AuthServiceConnector) string {
	return self.GetValueOrPanic(fmt.Sprintf("%s_%s_CLIENT_SECRET", self.name, strings.ToUpper(conn.Name())))
}

//ProjectObject defaults to looking at the GOPATH environment variable to work out the location
//of other objects in the project.  It calls GetValueOrPanic("GOPATH").
func (self *EnvironmentVars) ProjectObject(target string, projectName string, flavor FileFlavor) (string, error) {
	env := self.GetValueOrPanic("GOPATH")
	pieces := strings.Split(env, ":")
	if len(pieces) > 1 {
		env = pieces[0]
	}
	env = filepath.Clean(env)
	if strings.HasSuffix(env,"/") && env!="/"{
		env = env[0:len(env)-1]
	}
	switch flavor {
	case GO_SOURCE_FLAVOR:
		return filepath.Join(env, "src", target), nil
	case DART_FLAVOR:
		return filepath.Join(filepath.Dir(env), "dart", projectName, target), nil
	case ASSET_FLAVOR:
		return filepath.Join(filepath.Dir(env), "dart", projectName, "assets", target), nil
	case TOP_LEVEL_FLAVOR:
		return filepath.Join(filepath.Dir(env), target), nil
	}
	panic("unknown type of object searched for in the project!")
}

func (self *EnvironmentVars) GetValueOrPanic(n string) string {
	value:=os.Getenv(n)
	if value=="" {
		panic(fmt.Sprintf("expected to find environment variable %s but did not!",n))
	}
	return value
}