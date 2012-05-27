package plugin

import (

)

//PluginSet is the set of all plugins making up the Seven5 app.  It has
//to be visible to the Seven5 pill.
type PluginSet struct {
	Validator ProjectValidator
}

//Seven5App this is the "application" that is Seven5.
var Seven5App = &PluginSet{}

//SetProjectValidator for this app
func (self *PluginSet) SetProjectValidator(pv ProjectValidator) {
	self.Validator=pv
}
