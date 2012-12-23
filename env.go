package seven5

import (
	"os"
	"fmt"
	"strings"
)

//Environment encodes information that cannot be obtained from the source code, either because
//it can't be known or it is undesirable to commit these values to the source code repository.
type Environment interface {
	IsTest() bool
	ClientId(AuthServiceConnector) string
	ClientSecret(AuthServiceConnector) string
	
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
	return os.Getenv(fmt.Sprintf("%s_%s_CLIENT_ID", self.name,
		strings.ToUpper(conn.Name())))
}

//ClientSecret returns the value of the environment value of the client secret that
//has been given out the by the service associated with conn.  The environment variable
//is APPNAME_SERVICENAME_CLIENT_SECRET and is read only once.
func (self *EnvironmentVars) ClientSecret(conn AuthServiceConnector) string {
	return os.Getenv(fmt.Sprintf("%s_%s_CLIENT_SECRET", self.name,
		strings.ToUpper(conn.Name())))
}
