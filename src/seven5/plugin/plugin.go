package plugin

import (

)

//all commands
const (
	VALIDATE_PROJECT = "ValidateProject"
	ECHO = "Echo"
)

//PluginSet is the set of all plugins making up the Seven5 app.  It has
//to be visible to the Seven5 pill.
type PluginSet struct {
	Validator ValidateProject
	Echo Echo
}

//Seven5App this is the "application" that is Seven5.
var Seven5App = &PluginSet{}

//SetProjectValidator for this app
func (self *PluginSet) SetValidateProject(vp ValidateProject) {
	self.Validator=vp
}
//Set Echo plugin for this app
func (self *PluginSet) SetEcho(echo Echo) {
	self.Echo=echo
}
